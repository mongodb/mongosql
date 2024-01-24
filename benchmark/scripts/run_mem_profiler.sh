#!/bin/bash
profiler=$1
file=$2

# Create variables to track file names.
#  - heaptrack_results_file is the output file for full heaptrack results
#  - mem_usage_results_file is the output file that stores the heaptrack result summary
#  - mem_usage_cedar_file is the file that stores metrics tracked in evergreen
heaptrack_results_file=heaptrack.profiler.gz
mem_usage_results_file=memory-usage-results.log
mem_usage_cedar_file=memory-usage-cedar-data.json

# Get basename of test file
name=$(basename "$file" .yml)

# Initialize JSON structure for metrics-tracking with the test name
json_template="{
  \"info\": {
    \"test_name\": \"Memory Usage Profiler - $name\",
    \"args\": {}
  },
  \"metrics\": []
}"
json=$(echo $json_template | jq '.')

add_metric_to_json() {
  local metric_name="$1"
  local metric_value="$2"

  json=$(echo $json | jq --arg name "$metric_name (MB)" --argjson value "$metric_value" '.metrics += [{"name": $name, "value": $value}]')
}

# Run heaptrack
output=$(heaptrack $profiler $name)

# Extract the analyze command from heaptrack output
analyze_cmd=$(echo "$output" | grep '\-\-analyze')

# Extract the summary of results from heaptrack output
heaptrack_summary=$(eval $analyze_cmd | tail -5)

# Write data into the memory_usage_results_file
{
  echo "Query: $name"
  echo "$heaptrack_summary"
} >> $mem_usage_results_file

# Rename the full heaptrack results to the heaptrack_results_file
heaptrack_tar=$(echo $analyze_cmd | cut -d\" -f2)
mv -f $heaptrack_tar $heaptrack_results_file

# Extract relevant metrics
rss_peak=$(echo "$heaptrack_summary" | grep "peak RSS" | sed 's/.*: \([0-9.]*\)M.*/\1/')
heap_peak=$(echo "$heaptrack_summary" | grep "peak heap memory consumption" | sed 's/.*: \([0-9.]*\)M.*/\1/')
mem_leak=$(echo "$heaptrack_summary" | grep "total memory leaked" | sed 's/.*: \([0-9.]*\)M.*/\1/')

# Add metrics data to metrics file
add_metric_to_json "Heap peak" "$heap_peak"
add_metric_to_json "RSS peak" "$rss_peak"
add_metric_to_json "Memory leaked" "$mem_leak"

# Write the metrics to the mem_usage_cedar_file
echo "[$(echo "$json" | jq '.')]" > $mem_usage_cedar_file
