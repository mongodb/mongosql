#!/bin/bash
# This script processes pairs of log files and their corresponding titles to extract and plot duration
# data for query run times. It requires an even number of arguments representing log files and their titles,
# and generates a histogram plot using gnuplot

if (( $# % 2 != 0 )); then
    echo "Usage: $0 <log1> <title1> <log2> <title2> ... <logN> <titleN>"
    exit 1
fi

declare -A phase_durations
declare -A phase_names

extract_durations() {
    local log_file=$1
    local index=$2
    echo "Processing $log_file"
    while read -r line; do
        phase_name=$(echo "$line" | awk '{print $1}' | sed 's/_/-/g')
        duration=$(echo "$line" | awk '{print $2}')
        if [[ -n "$phase_name" && -n "$duration" ]]; then
            phase_durations["$index,$phase_name"]=$duration
            phase_names["$phase_name"]=1
        else
            echo "Warning: Skipping invalid line: $line" >&2
        fi
    done < "$log_file"
}

num_logs=$(( $# / 2 ))
log_files=()
titles=()
num_args=$#
for (( i=1; i<=num_args; i+=2 )); do
    log_files+=("$1")
    titles+=("$2")
    extract_durations "$1" $((i/2))
    shift 2
done

{
    sorted_phase_names=($(printf "%s\n" "${!phase_names[@]}" | sort -t'-' -k1.2n))
    for phase_name in "${sorted_phase_names[@]}"; do
        echo -n "$phase_name"
        for (( i=0; i<num_logs; i++ )); do
            echo -n " ${phase_durations["$i,$phase_name"]:-0}"
        done
        echo
    done
} > data.txt

{
    echo "set terminal pngcairo size 800,600"
    echo "set output 'TPCH_query_timings.png'"
    echo "set title 'TPC-H Query Timings'"
    echo "set ylabel 'Duration (ms)'"
    echo "set xtics rotate by -45"
    echo "set grid"
    echo "set style data histograms"
    echo "set style histogram clustered gap 1"
    echo "set style fill solid 1.00 border -1"
    echo -n "plot "
    for (( i=0; i<num_logs; i++ )); do
        if (( i > 0 )); then
            echo -n ", "
        fi
        echo -n "'data.txt' using $((i + 2)):xtic(1) title '${titles[$i]}' with histograms"
    done
} > plot.gp

gnuplot plot.gp
echo "plot.gp created"

