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

### Core Problems

*   acquire some generated data
*   process/collate said data for audience consumption
*   deliver processed data to appropriate destination

### Features

*   convert data from outside sources into a standard internal representation(area,ts,json)
*   perform any required 'in flight' processing
*   deliver collated data to intended destination(s)
*   approximate quantiles over an unbounded data stream(such as MAU)
*   sliding-window events alarming

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
                            +-------------------------------------------------------+
                            |                   |           |           |           |
                       realtime analysis     indexer     archive    BehaviorDB      S3
                            |                   |           |           |           |
                   +-----------------+          |           |           |           |
                   |    |     |      |   ElasticSearch    HDFS      LevelDB/sky   RedShift
                 beep email console etc         |           |           |           |
                   |    |     |      |          |           |           |           |
                   +-----------------+       Kibana3        |           |        tableau
                            |                   |           |           |           |
                         dev/ops               PM                                  PM

#### Parsers

*   data filtering
*   data parsing
*   data chaining
*   data monitoring and alarming

#### Data

*   app performance metrics: statsd/graphite
*   app biz metrics: analytics/MR/dashboard
*   app error/traceback: arecibo/sentry
*   security/anomalous activity events: CEF/CEP/arcsight
*   log file messages: logstash/syslog
*   system events: nagios/zenoss

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

