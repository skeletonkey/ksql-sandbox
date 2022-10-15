# KSQL Sandbox

A place to experiment with KSQL.

This repo provides a data source (data/*) and how to start working with KSQL using Confluence's all-in-one repo.

## Dependencies

* [Confluent All In One Repo](https://github.com/confluentinc/cp-all-in-one)
  * Make sure to up your Docker memory to greater than 6G - see docs
* Mac only: `brew install kafka`
  * this is all the kafka tool CLIs that you will need
  * Windows/Linux users will need to figure out how to properly install CLI tools (please provide pull request with instructions)
* [ksql_sandbox_data.tgz](https://www.dropbox.com/s/1y1ijfz5dwp4dqm/ksql_sandbox_data.tgz?dl=0)
  * unpack file into the `producer/data` directory

## Steps

* Steps done in Confluent's cp-all-in-one repo
  * `cd cp-all-in-one`
  * `docker-compose up -d`
    * if this doesn't work try `docker-compose-v1 up -d`
    * This step takes a while (even after everything is downloaded).  Check [WebUI](http://localhost:9021) to make sure it's up. 
  * `kafka-topics --create --topic explore-main --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1`
    * creates the topic `explore-main`
  * Run the data producer in another terminal (same directory)
    * NOTE: `throttle` sets the 'pace' of sending data - adjust it to meet your needs
    * `cd prducer`
    * `go run main.go`
  * Log into the KSQL 
    * `docker exec -it ksqldb-cli /bin/bash`
      * `ksql http://ksqldb-server:8088`
        * This gets you into the KSQL CLI
      * Start playing - see KSQL Commands for examples

## KSQL Commands

These commands are meant to be run in order:

`SHOW TOPCIS` - see available topics

`PRINT 'explore-main' limit 5;`
This will print 5 events from `explore-main`.  This is from the end of the topic - so if you have set a large threshold in the producer it may take a while.

`create stream explore_stream (station varchar, name varchar, record_date varchar, record_value int) with (kafka_topic='explore-main', value_format='JSON');`
Create a stream.  Streams are required to be able to run queries.

`create stream explore_stream_filtered as select * from explore_stream where record_value != 0 and record_value > -9999;`
Query from the `explore_stream` stream to form a new topic

## Kcat commands

Once the `KSQL Commands` have been run you should be able to view the stream with the following commands in individual terminals:

`kcat -b localhost:9092 -t explore-main -C`
View the original stream of data from the producer

`kcat -C -b localhost:9092 -t EXPLORE_STREAM_FILTERED`
View the new stream from the KSQL query

## Work in progress

### MongoDB

Trying to get a MongoDB connector to work.

Add MongoDB to `cp-all-in-one` docker compose:
```
  mongo:
    image: mongo
    hostname: mongodb
    container_name: mongodb
    restart: always
    environment:
      MONGO_INITDB_ROOT_USERNAME: root
      MONGO_INITDB_ROOT_PASSWORD: example

  mongo-express:
    image: mongo-express
    restart: always
    ports:
      - 8089:8081
    environment:
      ME_CONFIG_MONGODB_ADMINUSERNAME: root
      ME_CONFIG_MONGODB_ADMINPASSWORD: example
      ME_CONFIG_MONGODB_URL: mongodb://root:example@mongo:27017/
```

Need to figure out how to install the MongoDB connector, but the restart doesn't work.

Connect to the connector: `docker exec -it connect /bin/bash`
Install the connector: `confluent-hub install mongodb/kafka-connect-mongodb:1.2.0`
Restart the connector: `docker container restart connect`
Unfortunately, now the connector is no longer accessible.

Create a sink:
```ksql
CREATE SINK CONNECTOR `mongodb-test-sink-connector` WITH (
   "connector.class"='com.mongodb.kafka.connect.MongoSinkConnector',
   "key.converter"='org.apache.kafka.connect.json.JsonConverter',
   "value.converter"='org.apache.kafka.connect.json.JsonConverter',
   "key.converter.schemas.enable"='false',
   "value.converter.schemas.enable"='false',
   "tasks.max"='1',
   "connection.uri"='mongodb://mongodb:27017/admin?readPreference=primary&appname=ksqldbConnect&ssl=false',
   "database"='erik_filtered',
   "collection"='mongodb-connect',
   "topics"='ERIK_STREAM_FILTERED'
);
```

[Web UI](http://localhost:8089/)

## Reference

* [Confluent Local Quickstart](https://docs.confluent.io/5.4.3/ksql/docs/tutorials/basics-local.html#ksql-quickstart-local)
* [Confluent cp-all-in-one](https://github.com/confluentinc/cp-all-in-one)
  * `cd cp-all-in-one`
  * `docker-compose-v1 up -d`
  * `docker-compose-v1 down`