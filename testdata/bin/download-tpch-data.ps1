wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_small.bson.gz -OutFile ../resources/data/tpch_small.bson.gz
wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_normalized.bson.gz -OutFile ../resources/data/tpch_full_normalized.bson.gz
wget http://noexpire.s3.amazonaws.com/sqlproxy/data/tpch_full_denormalized.bson.gz -OutFile ../resources/data/tpch_full_denormalized.bson.gz

cmd /c mklink tpch_small.bson.gz tpch.bson.gz