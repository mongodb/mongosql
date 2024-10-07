#!/bin/bash
set -e

pipeline_dir=$1
workload_filename=$2
dataset_filename=$3
EVERGREEN_SF0_01_LIMIT=$4
EVERGREEN_SF0_1_LIMIT=$5
skip_pipeline_generation=$6

function download {
    echo "Downloading artifact $1"
    curl -LO "$1" \
        --silent \
        --fail \
        --max-time 60 \
        --retry 5 \
        --retry-delay 0
}

function setup_yq {
    local yq_url="https://mongosql-noexpire.s3.us-east-2.amazonaws.com/tpch/yq_linux_amd64"
    download $yq_url

    mv ./yq_linux_amd64 ./yq || true
    chmod +x ./yq
    echo "yq setup complete"
}

function process_genny_output() {
    phase_names=($(./yq '.Actors[0].Phases[].LoadConfig.Key' benchmark/${pipeline_dir}/${workload_filename}))

    while read -r line; do
        if [[ $line =~ "Beginning phase" ]]; then
            phase_number=$(echo "$line" | grep -oP 'Beginning phase \K[0-9]+')
            timestamp=$(echo "$line" | grep -oP '\[.*?\]' | head -1 | tr -d '[]')
            start_time=$(date -d "$timestamp" +%s%N)
            start_times[$phase_number]=$start_time
        elif [[ $line =~ "Ended phase" ]]; then
            phase_number=$(echo "$line" | grep -oP 'Ended phase \K[0-9]+')
            timestamp=$(echo "$line" | grep -oP '\[.*?\]' | head -1 | tr -d '[]')
            end_time=$(date -d "$timestamp" +%s%N)
            duration=$(((end_time - start_times[$phase_number]) / 1000000))
            phase_durations[$phase_number]=$duration
        fi
    done <<<"$1"

    if [[ ${#phase_names[@]} != ${#phase_durations[@]} ]]; then
        echo "Error: The number of phase names does not match the number of phase durations." >&2
        exit 1
    fi

    for i in "${!phase_durations[@]}"; do
        if [[ ${phase_names[$i]} =~ "NoopPhase" ]] || [[ ${phase_names[$i]} =~ "TPCHNormalizedQuery15" ]]; then
            continue
        fi
        echo "${phase_names[$i]} ${phase_durations[$i]}"
    done
}

echo "setting up Genny"
git clone git@github.com:mongodb/genny ../genny || true
cd ../genny
./run-genny install -d ubuntu2204
cd -

echo "setting up yq"
setup_yq

echo "setting up mongodb-database-tools"
download https://fastdl.mongodb.org/tools/db/mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION.tgz
tar zxvf mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION.tgz

echo "setting up mongosh"
download https://downloads.mongodb.com/compass/mongosh-$MONGOSH_VERSION-linux-x64.tgz
tar zxf mongosh-$MONGOSH_VERSION-linux-x64.tgz

# call pipeline generator to create atlas_sql/(normalized/denormalized) phase files
if [[ -z "$skip_pipeline_generation" ]]; then
    if [[ "${pipeline_dir}" == *"validation"* ]]; then
        cargo run --bin pipeline_generator --features genny,validation
    else
        cargo run --bin pipeline_generator --features genny
    fi
fi

echo "Downloading: https://mongosql-noexpire.s3.us-east-2.amazonaws.com/tpch/${dataset_filename}"
download "https://mongosql-noexpire.s3.us-east-2.amazonaws.com/tpch/${dataset_filename}"
mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION/bin/mongorestore --drop --numInsertionWorkersPerCollection=8 --bypassDocumentValidation --gzip --archive=${dataset_filename}

echo "Starting SAR resource logging"
# Log resource usage every 5 seconds
SAR_INTERVAL=5
stdbuf -oL sar -r $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' >sar-mem.log &
SAR_MEM=$!
stdbuf -oL sar -d $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' >sar-disk.log &
SAR_DISK=$!
stdbuf -oL sar -P ALL $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' >sar-cpu.log &
SAR_CPU=$!

# add validation dataset separately if needed; load into `tpch` database
if [[ "${pipeline_dir}" == *"validation"* ]]; then
    echo "Downloading: https://dsi-donot-remove.s3.us-west-2.amazonaws.com/tpch/tpch-validation.archive.gz"
    download "https://dsi-donot-remove.s3.us-west-2.amazonaws.com/tpch/tpch-validation.archive.gz"
    mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION/bin/mongorestore --nsFrom "validation.*" --nsTo "tpch.*" --numInsertionWorkersPerCollection=8 --bypassDocumentValidation --gzip --archive=tpch-validation.archive.gz
fi

# run genny, and print out database profiling information
mongosh-$MONGOSH_VERSION-linux-x64/bin/mongosh tpch --eval "config.set('inspectDepth', Infinity); db.setProfilingLevel(2)"
echo "Running workload: benchmark/${pipeline_dir}/${workload_filename}"

../genny/run-genny workload benchmark/"${pipeline_dir}"/"${workload_filename}" 2>&1 | tee genny_output.log
echo "Genny Results:"
cat genny_output.log

mongosh-$MONGOSH_VERSION-linux-x64/bin/mongosh tpch --eval "db.system.profile.find().toArray().forEach(doc => printjson(doc))" >profile.log

if [[ "${pipeline_dir}" != *"validation"* ]]; then
    query_run_results=$(process_genny_output "$(<genny_output.log)")

    if [[ $pipeline_dir == *"bic"* ]]; then
        type="BIC"
    else
        type="AtlasSQL"
    fi

    phase=0
    result_str=""
    total_duration=0
    while IFS= read -r line; do
        query=$(echo $line | cut -d' ' -f1)
        duration=$(echo $line | cut -d' ' -f2)
        total_duration=$((total_duration + duration))
        duration_sec=$(awk "BEGIN {print $duration/1000}")
        json=$(jq -n \
            --arg phase "$phase" \
            --arg query "$query" \
            --arg duration_sec "$duration_sec" \
            '{
        "info": {
          "test_name": $query,
          "args": {}
        },
        "metrics": [{
          "name": ("duration_" + $query),
          "type": "SUM",
          "value": '"$duration_sec"'
        }]
      }')
        if [ "$result_str" != "" ]; then
            result_str="$result_str,"
        fi
        result_str="$result_str$json"
        ((phase++)) || true
    done <<<"$query_run_results"

    total_duration_s=$(awk "BEGIN {print $total_duration/1000}")

    entry=',{
    "info": {
      "test_name": "TPC-H Benchmark Runtime Total",
      "args": {}
    },
    "metrics": [
    {
      "name": "total_duration_s",
      "type": "SUM",
      "value": '"$total_duration_s"'
    }]
  }'
    result_str="$result_str$entry"

    result_str="[$result_str]"
    echo "$result_str" >processed-tpch-perf-results.json
    cat processed-tpch-perf-results.json

    # Print out the translations
    max_q=$(ls benchmark/${pipeline_dir}/phase/q[0-9]*.yml | sed -E 's/.*q([0-9]+)_.*.yml/\1/' | sort -n | tail -n 1)

    for i in $(seq 1 $max_q); do
        file=$(ls benchmark/${pipeline_dir}/phase/q${i}_*.yml 2>/dev/null || true)
        if [ -f "$file" ]; then
            cat "$file" >>translations.log
        fi
    done
fi

rm -f ${dataset_filename}
pkill mongod

kill -9 $SAR_MEM $SAR_DISK $SAR_CPU

if [[ "${pipeline_dir}" == *"validation"* ]]; then
    exit 0
fi

plot_data_filename=${pipeline_dir//\//_}_${workload_filename}.log
echo "$query_run_results" >$plot_data_filename
# Generates a plot named: TPCH_query_timings.png
$(dirname "$0")/generate_plot.sh "$plot_data_filename" "${pipeline_dir} $(basename "${workload_filename}" .yml | cut -d'_' -f3)"

tar -czvf tpch.metrics.tgz --ignore-failed-read ./sar-mem.log \
    ./sar-disk.log \
    ./sar-cpu.log \
    "./$plot_data_filename" \
    ./profile.log \
    ./translations.log \
    ./TPCH_query_timings.png

total_duration_s_int=${total_duration_s%.*}

# Check whether duration exceeds configured limit
case $pipeline_dir in
atlas_sql/normalized*)
    case $workload_filename in
    *sf0.01*)
        (($total_duration_s_int > $EVERGREEN_SF0_01_LIMIT)) && echo "ERROR: The query total runtime is greater than $EVERGREEN_SF0_01_LIMIT Seconds." && exit 1
        ;;
    *sf0.1*)
        (($total_duration_s_int > $EVERGREEN_SF0_1_LIMIT)) && echo "ERROR: The query total runtime is greater than $EVERGREEN_SF0_1_LIMIT Seconds." && exit 1
        ;;
    *)
        echo "Workload $workload_filename is not checked for total runtime"
        ;;
    esac
    ;;
*)
    echo "Pipeline directory $pipeline_dir is not checked for total runtime"
    ;;
esac

echo "Script completed successfully."
exit 0
