#!/bin/bash

# Create JSON structure for new tasks. Note that we must also include the
# buildvariant to which we want to add these tasks.
json_template='{
  "tasks": [],
  "buildvariants": [
    {
      "name": "mem-usage",
      "tasks": []
    }
  ]
}'
json=$(echo $json_template | jq '.')

# For each file in the test directory, create a task using the file's basename
# to uniquely assign a task name and using the file's path as an argument to the
# test function. Also add the task to the buildvariant.
for file in benchmark/profiler/src/config_loader/queries/*.yml; do
  name=$(basename "$file" .yml)
  json=$(echo $json | jq --arg task_name "mem-usage-profiler-$name" --arg test_file_basename "$name" --arg test_file "$file" '.tasks += [{
    "name": $task_name,
    "commands": [
      { "func": "install rust toolchain" },
      { "func": "install heaptrack" },
      {
        "func": "run memory usage profiler",
        "vars": {
          "test_file": $test_file,
          "test_file_basename": $test_file_basename
        }
      }
    ]
  }] | .buildvariants[0].tasks += [{ "name": $task_name }]')
done

echo "$json" > generated_profiler_tasks.json
