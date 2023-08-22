# TPC-H Helper Scripts  
This document contains instructions for setting up and running TPC-H queries on a newly spawned 
evergreen `ubuntu2204-large` host [here](https://spruce.mongodb.com/spawn/host).

After creating the host, ssh to it and clone the mongosql-rs repository:
```
git clone git@github.com:10gen/mongosql-rs.git
cd mongosql-rs/benchmark/scripts
```

## Setup  
- Clone the mongosql-rs repository to your host.
- Run the `setup.sh` script.
  - By default loads a Scale Factor 0.001 dataset
  - Takes an optional scale factor parameter: `0.01`, `0.1`, `1`
```
./setup.sh

# With optional parameter
./setup.sh 0.1
```

## Run Query  
Use the `query.sh` script to run the TPC-H pipelines.  
Query times run through this script are a little over a second longer than queries run through the Genny tool.
```
# Takes full query file name as input
./query.sh q1_normalized.yml
./query.sh q2_denormalized.yml
```

## Check Scale Factor  
This helper script will output the scale factor
```
$ ./scale_factor 
Scale factor 0.01
```

