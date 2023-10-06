#!/bin/bash
set -e

function download {
    echo "Downloading artifact $1"
    curl -LO $1 \
     --silent \
     --fail \
     --max-time 60 \
     --retry 5 \
     --retry-delay 0
}

pipeline_dir=$1
workload_filename=$2
dataset_filename=$3

echo "downloading mongodb"
download https://fastdl.mongodb.org/linux/$MONGODB_TEST_UBUNTU_VERSION.tgz
tar zxvf $MONGODB_TEST_UBUNTU_VERSION.tgz
mkdir -p data_db

echo "starting mongodb"
$MONGODB_TEST_UBUNTU_VERSION/bin/mongod --dbpath data_db &

echo "setting up Genny"
git clone git@github.com:mongodb/genny ../genny
cd ../genny
# Workaround until issue in `lamplib` requirements.txt is fixed
# EVG-20471
git checkout e50f6d5
./run-genny install -d ubuntu2204
cd -

echo "setting up mongodb-database-tools"
download https://fastdl.mongodb.org/tools/db/mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION.tgz
tar zxvf mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION.tgz

echo "setting up mongosh"
download https://downloads.mongodb.com/compass/mongosh-$MONGOSH_VERSION-linux-x64.tgz
tar zxf mongosh-$MONGOSH_VERSION-linux-x64.tgz

# call pipeline generator to create atlas_sql/(normalized/denormalized) phase files
if [[ "${pipeline_dir}" == *"validation"* ]]; then
  cargo run --bin pipeline_generator --features genny,validation
else
  cargo run --bin pipeline_generator --features genny
fi

# echo "Downloading: https://dsi-donot-remove.s3.us-west-2.amazonaws.com/tpch/${dataset_filename}"
download "https://mongosql-noexpire.s3.us-east-2.amazonaws.com/tpch/${dataset_filename}"
mongodb-database-tools-ubuntu2204-x86_64-$DB_TOOLS_VERSION/bin/mongorestore --drop --numInsertionWorkersPerCollection=8 --bypassDocumentValidation --gzip --archive=${dataset_filename}

echo "Starting SAR resource logging"
# Log resource usage every 5 seconds
SAR_INTERVAL=5
stdbuf -oL sar -r $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' > sar-mem.log &
SAR_MEM=$!
stdbuf -oL sar -d $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' > sar-disk.log &
SAR_DISK=$!
stdbuf -oL sar -P ALL $SAR_INTERVAL | awk '{ print strftime("%Y-%m-%d "), $0; fflush(); }' > sar-cpu.log &
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
../genny/run-genny workload benchmark/${pipeline_dir}/${workload_filename}
mongosh-$MONGOSH_VERSION-linux-x64/bin/mongosh tpch --eval "db.system.profile.find().pretty()" > profile.log

# Convert ftdc to json for perf.send. Only for timing tests, not for validation
if [[ "${pipeline_dir}" != *"validation"* ]]; then
  PERF_RESULTS_FILE=tpch-perf-results.json

  if [[ $pipeline_dir == *"bic"* ]]; then
    type="BIC"
  else
    type="AtlasSQL"
  fi
  curator ftdc export json --input=build/WorkloadOutput/CedarMetrics/NewTest.$type.ftdc \
      --output=$PERF_RESULTS_FILE

  phase=0
  result_str=""
  while IFS= read -r line
  do
    dur_ns=$(echo $line | jq -r '.timers.dur')
    current_dur_s=$(awk "BEGIN {print $dur_ns/1000000000}")
    current_dur_ms=$(echo "$current_dur_s * 1000" | bc | cut -d'.' -f1)
    if [[ "$phase" -eq 0 ]]; then
        elapsed_time=$current_dur_s
    else
        elapsed_time=$(bc <<< $current_dur_s-$previous_dur_s)
    fi

    json=$(jq -n \
        --arg phase "$phase" \
        --argjson elapsed_time "$elapsed_time" \
        '{
          "info": {
            "test_name": ("phase " + $phase),
            "args": {}
          },
          "metrics": [{
            "name": ("duration_phase_" + $phase),
            "type": "SUM",
            "value": $elapsed_time
          }]
        }')
    if [ "$result_str" != "" ]; then
      result_str="$result_str,"
    fi
    result_str="$result_str$json"
    ((phase++)) || true
    previous_dur_s=$current_dur_s
  done < "$PERF_RESULTS_FILE"

  # Add total runtime
  entry=',{
          "info": {
            "test_name": "TPC-H Benchmark Runtime Total",
            "args": {
              "duration_ms": '"$current_dur_ms"'
            }
          },
          "metrics": [{
            "name": "total_duration",
            "type": "SUM",
            "value": '"$current_dur_s"'
          }]
        }'
  result_str="$result_str$entry"
  result_str="[$result_str]"
  echo "$result_str" > processed-tpch-perf-results.json

  # Print out the translations
  max_q=$(ls benchmark/${pipeline_dir}/phase/q[0-9]*.yml | sed -E 's/.*q([0-9]+)_.*.yml/\1/' | sort -n | tail -n 1)

  for i in $(seq 1 $max_q); do
    file=$(ls benchmark/${pipeline_dir}/phase/q${i}_*.yml 2>/dev/null || true)
    if [ -f "$file" ]; then
      cat "$file" >> translations.log
    fi
  done
fi

rm ${dataset_filename}
pkill mongod

kill -9 $SAR_MEM $SAR_DISK $SAR_CPU
tar -czvf sar.log.tgz ./sar-mem.log ./sar-disk.log ./sar-cpu.log

