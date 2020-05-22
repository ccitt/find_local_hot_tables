# Introduction
find_local_hot_tables is a local database instance hotspot update table statistic sort tool.Used to quickly locate master-slave replication delays for local mysql instances. And high I/O load of database system.

## Installation
tar -xzvf find_local_hot_tables.linux32/64.tar

chmod 750 find_local_hot_tables

## Use Examples
//Analyze the local 3306 instance hotspot update table from 2020-05-22 09:50:00 to 2020-05-22 09:55:00

./find_local_hot_tables -h 127.0.0.1 -P 3306 -u username -p password -start-datetime "2020-05-22 09:50:00" -stop-datetime "2020-05-22 09:55:00"

//Analyze the local 3307 instance hotspot update table last 10 Minutes

./find_local_hot_tables -h 127.0.0.1 -P 3307 -u username -p password

## Note
* Run only on 32/64-bit linux operating systems.
* Requires Root permission.
* Requires mysqlbing to be installed.
* The binlog format only supports row mode

## Authors
[@ccitt](https://github.com/ccitt)

## Email
ccitt@tom.com

## License
find_local_hot_tables is under the Apache 2.0 license. 
