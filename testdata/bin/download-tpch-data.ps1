wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_small.bson.archive.gz -OutFile ../resources/data/tpch_small.bson.archive.gz
wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_normalized.bson.archive.gz -OutFile ../resources/data/tpch_full_normalized.bson.archive.gz
wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_denormalized.bson.archive.gz -OutFile ../resources/data/tpch_full_denormalized.bson.archive.gz

cmd /c mklink tpch_small.bson.archive.gz tpch.bson.archive.gz