funpipe
=======

Performing "in-flight" processing of collected data, real time streaming analysis and alarming, and delivering the results to any number of destinations for further analysis.

*   aimed to be swiss knife kind tool instead of a complex system.
*   different from splunk which is basically a search engine.
*   different from opentsdb which is metrics based while funpipe is event based.
*   like logstash but can do more data manipulations and has feature of real time analysis.


[![Build Status](https://travis-ci.org/funkygao/funpipe.png?branch=master)](https://travis-ci.org/funkygao/funpipe)

### Install

    go get github.com/funkygao/funpipe
    funpipe -h # help

### Architecture

#### Overview

        +---------+     +---------+     +---------+     +---------+
        | server1 |     | server2 |     | server3 |     | serverN |
        |---------|     |---------|     |---------|     |---------|
        |syslog-ng|     |syslog-ng|     |syslog-ng|     |syslog-ng|
        |---------|     |---------|     |---------|     |---------|
        |collector|     |collector|     |collector|     |collector|
        +---------+     +---------+     +---------+     +---------+
            |               |               |               |
             -----------------------------------------------
                                    |
                                    | HTTP POST
                                    |
                            +-----------------+
                            |   ALS Server    |
                            |-----------------| 
                            | funpipe daemon  |
                            +-----------------+
                                    |
                                    | clean/filter/parse/transform based on rule engine
                                    |
                                    | output target
                                    |
                            +-------------------------------------------+
                            |                   |           |           |
                       realtime analysis     indexer     archive    BehaviorDB
                            |                   |           |           |
                   +-----------------+          |           |           |
                   |    |     |      |   ElasticSearch    HDFS      LevelDB/sky
                 beep email console etc         |
                                                |
                                                |
                                             Kibana3
                                                |
                                                |
                                               .-.
                                              (e.e)
                                               (m)
                                             .-="=-.  W
                                            // =T= \\,/
                                           () ==|== ()
                                            \  =V=
                                             M(oVo)
                                              // \\
                                             //   \\
                                            ()     ()
                                             \\    ||
                                          PM  \'   '|
                                            =="     "==

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

