#!/bin/bash
set -e

echo "Starting TPCH benchmark..."

run_benchmark() {
    local pipeline_dir=$1
    local workload_file=$2
    local archive_file=$3
    local output_log=$4
    local skip=$5
    echo -e "${GREEN}Running benchmark for $pipeline_dir with workload $workload_file ${NC}"
    ./benchmark/scripts/run_tpch.sh "$pipeline_dir" "$workload_file" "$archive_file" $limit $limit $skip
    plot_data_filename=${pipeline_dir//\//_}_${workload_file}.log
    mv "$plot_data_filename" "$output_log"
}

GREEN='\033[0;32m'
NC='\033[0m'
current_plot="current.log"
bic_plot="bic.log"
modified_plot="modified.log"
limit=$((2**63-1))

pushd ../.. > /dev/null

# Run benchmarks
run_benchmark "atlas_sql/normalized" "workload_norm_sf0.001.yml" "tpch-0.001-normalized-indexed.archive.gz" $current_plot
run_benchmark "bic/normalized" "workload.yml" "tpch-0.001-normalized-indexed.archive.gz"  $bic_plot

# Check if a Git commit hash was provided
if [[ -n "$1" ]]; then
    git clone git@github.com:10gen/mongosql-rs.git mongosql-rs-benchmark || true
    pushd mongosql-rs-benchmark > /dev/null
    git checkout "$1"
    cargo run --bin pipeline_generator
    popd > /dev/null
    mv ./mongosql-rs-benchmark/benchmark/atlas_sql/normalized/phase/* benchmark/atlas_sql/normalized/phase/
    run_benchmark "atlas_sql/normalized" "workload_norm_sf0.01.yml" "tpch-0.01-normalized-indexed.archive.gz" $modified_plot "skip"
    # Generate plots including the modified results
    ./benchmark/scripts/generate_plot.sh $current_plot current $bic_plot BIC $modified_plot "from $1"
else
    ./benchmark/scripts/generate_plot.sh $current_plot current $bic_plot BIC
fi

popd > /dev/null

echo "Benchmark completed successfully."
