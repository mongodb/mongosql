#!/bin/bash
profiler=$1

rss_peak_max=0
heap_peak_max=0
mem_leak_max=0
rss_peak_accum=0
heap_peak_accum=0
query_count=0

for file in benchmark/profiler/src/config_loader/queries/*.yml; do
  name=$(basename "$file" .yml)
  output=$(heaptrack $profiler $name)
  analyze_cmd=$(echo "$output" | grep '\-\-analyze')
  heapstack_summary=$(eval $analyze_cmd | tail -5)
  echo "Query: $name"
  echo "$heapstack_summary"

  rss_peak=$(echo "$heapstack_summary" | grep "peak RSS" | sed 's/.*: \([0-9.]*\)M.*/\1/')
  heap_peak=$(echo "$heapstack_summary" | grep "peak heap memory consumption" | sed 's/.*: \([0-9.]*\)M.*/\1/')
  mem_leak=$(echo "$heapstack_summary" | grep "total memory leaked" | sed 's/.*: \([0-9.]*\)M.*/\1/')

  if (( $(bc <<< "$rss_peak > $rss_peak_max") )); then
    rss_peak_max=$rss_peak
  fi
  if (( $(bc <<< "$heap_peak > $heap_peak_max") )); then
    heap_peak_max=$heap_peak
  fi
  if (( $(bc <<< "$mem_leak > $mem_leak_max") )); then
    mem_leak_max=$mem_leak
  fi
  rss_peak_accum=$(bc <<< "$rss_peak_accum + $rss_peak")
  heap_peak_accum=$(bc <<< "$heap_peak_accum + $heap_peak")

  (( query_count++ ))
done

rss_peak_avg=$(bc <<< "scale=2; $rss_peak_accum/$query_count")
heap_peak_avg=$(bc <<< "scale=2; $heap_peak_accum/$query_count")
echo "-------------------------------------------------"
printf "Average RSS peak:  %10.2fMB\n" "${rss_peak_avg}"
printf "Average heap peak: %10.2fMB\n" "${heap_peak_avg}"
printf "Memory leaked max: %10.2fMB\n" "${mem_leak_max}"
printf "RSS peak max:      %10.2fMB\n" "${rss_peak_max}"
printf "Heap peak max:     %10.2fMB\n" "${heap_peak_max}"
