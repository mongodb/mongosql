#!/bin/sh

# This script will populate a live MySQL database (a real one) with the Mongo data used for the tableau demo.
# 1. Mongoexport the data from a live server that contains the data set to CSV
# 2. Create new tables with schemas matching that of the output from mongoexport
# 3. Use Mysql's LOAD DATA INTO to load the CSV into the tables

# The real MySQL server can then be used as a reference point for checking the behavior of the SQLProxy server.

set -v
mongoexport --host bi-beta.mongodb.com --type=csv --fields="_id,flight_date,carrier_code,tail_num,flight_number,origin_airport_id,origin_city_market_id,origin_airport_code,origin_city_name,origin_state,dest_airport_id,dest_city_market_id,dest_airport_code,dest_city_name,dest_state,dep_time,diff_from_dep_time,dep_delay,arr_time,diff_from_arr_time,flight_time,airline,cancelled,arr_delay,carrier_delay" -d tableau -c flights201406 > flights201406.csv;
mongoexport --host bi-beta.mongodb.com --type=csv --fields="_id,loc.coordinates.0,loc.coordinates.1,airport_city,airport_code,airport_id,city,city_market_description,city_market_id,country,state,zip" -d tableau -c attendees > attendees.csv;
mysql -u root -e 'DROP DATABASE if exists test';
mysql -u root -e 'CREATE DATABASE test CHARACTER SET utf8 COLLATE utf8_bin';
mysql -u root -e 'DROP TABLE IF EXISTS flights201406' test
mysql -u root -e 'DROP TABLE IF EXISTS attendees' test
mysql -u root -e 'CREATE TABLE flights201406  (_id varchar(255), flight_date date, carrier_code varchar(255), tail_num varchar(255), flight_number int, origin_airport_id int, origin_city_market_id int, origin_airport_code varchar(255), origin_city_name varchar(255), origin_state varchar(255), dest_airport_id int, dest_city_market_id int, dest_airport_code varchar(255), dest_city_name varchar(255), dest_state varchar(255), dep_time int, diff_from_dep_time int, dep_delay int, arr_time int, diff_from_arr_time int, flight_time int, airline varchar(255), cancelled int, arr_delay int, carrier_delay int);' test
mysql -u root -e 'CREATE TABLE attendees (_id varchar(255), longitude double, latitude double, airport_city varchar(255),airport_code varchar(255),airport_id int,city varchar(255),city_market_description varchar(255),city_market_id int,country varchar(255),state varchar(255),zip varchar(255))' test
CWD=`pwd`
mysql -u root -e "LOAD DATA INFILE '$CWD/flights201406.csv' INTO TABLE flights201406 FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n' IGNORE 1 LINES" test
mysql -u root -e "LOAD DATA INFILE '$CWD/attendees.csv' INTO TABLE attendees FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n' IGNORE 1 LINES" test
