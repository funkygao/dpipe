alser
=====

ALS guard 

Performing "in-flight" processing of collected data, real time streaming analysis and alarming, and delivering the results to any number of destinations for further analysis.

*   als is aimed to be swiss knife kind tool instead of a complex system.
*   als is different from splunk which is basically a search engine.
*   als is different from opentsdb which is metrics based while als is event based.
*   als is like logstash but can do more data manipulations and has feature of real time analysis.


[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

### Install

    go get github.com/funkygao/alser
    alser -h # help

### Architecture

#### Overview

        +---------+     +---------+     +---------+     +---------+
        | server1 |     | server2 |     | server3 |     | serverN |
        +---------+     +---------+     +---------+     +---------+
            |               |               |               |
             -----------------------------------------------
                                    |
                                    | log events datasources(db | text log)
                                    |
                            +-----------------+
                            |   ALS Server    |
                            |-----------------| 
                            | alser daemon    |
                            +-----------------+
                                    |
                                    | filtering/parsing based on rule engine
                                    |
                                    | output target
                                    |
                            +-------------------------------------------+
                            |                   |           |           |
                       realtime analysis     indexer     archive    BehaviorDB
                            |                   |           |           |
                   +-----------------+          |           |           |
                   |    |     |      |   ElasticSearch    HDFS      LevelDB/sky
                 beep email console etc


#### Parsers

*   data filtering
*   data parsing
*   data chaining
*   data monitoring and alarming

### BI

* T+1

* RealTime/streaming

        +-------+   +-------+   +-------+   +-------+
        | app   |   | app   |   | app   |   | app   |
        |-------|   |-------|   |-------|   |-------|
        | agent |   | agent |   | agent |   | agent |
        +-------+   +-------+   +-------+   +-------+
            |           |           |           |
             -----------------------------------
                                |
                    +-----------------------+
                    | load balance cluster  |
                    | collector distributor |
                    +-----------------------+
                                | routing
                                |
                                |                            persist
                                |-----------------------------------
             RealTimeAnalysis   |                                   |
             -----------------------------------            +-------------------------+
            |           |           |           |           | central logfile cluster |
        +-------+   +-------+   +-------+   +-------+       +-------------------------+
        |secure |   | debug |   |monitor|   |analyse|
        +-------+   +-------+   +-------+   +-------+


### TODO

    priority queue https://github.com/daviddengcn/go-villa
    websocket for web based alarm
    lua integration to parse complex log

    geodbfile   indexer.domain

    add indexing to HostLineParser and RegexCollectorParser

    add dimensions e,g. abtype/paid to log json

