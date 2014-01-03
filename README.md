funpipe
=======

Performing "in-flight" processing of collected data, real time streaming analysis and alarming, and delivering the results to any number of destinations for further analysis.

[![Build Status](https://travis-ci.org/funkygao/funpipe.png?branch=master)](https://travis-ci.org/funkygao/funpipe)

### Install

    go get github.com/funkygao/funpipe
    funpipe -h # help

### Features

*   convert data from outside sources into a standard internal representation(area,ts,json)
*   perform any required 'in flight' processing
*   deliver collated data to intended destination(s)
*   approximate quantiles over an unbounded data stream(such as MAU)
*   sliding-window events alarming
*   distributed cluster ready

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
                                    | input/decode/clean/filter/output
                                    |
                                   +-------------------------------------------------------+
                                   |                   |           |           |           |
                              realtime analysis     indexer     archive    BehaviorDB      S3
                                   |                   |           |           |           |
            +----------------------|                   |           |           |           |
            |                      |                   |           |           |           |
       +----------+       +-----------------+          |           |           |           |
       |          |       |    |     |      |   ElasticSearch    HDFS      LevelDB/sky   RedShift
     quantile   hyper     |    |   color    |          |           |           |           |
    histogram  loglog   beep email console etc         |           |           |           |
      topN        |       |    |     |      |          |           |           |           |
       |          |       +-----------------+       Kibana3        |           |        tableau
       +----------+                |                   |           |           |           |
            |                      |                   |           |           |           |
          PM/dev                dev/ops               PM          ops         PM          PM


#### Data

*   app performance metrics: statsd/graphite
*   app biz metrics: analytics/MR/dashboard
*   app error/traceback: arecibo/sentry
*   security/anomalous activity events: CEF/CEP/arcsight
*   log file messages: logstash/syslog
*   system events: nagios/zenoss

### BI

* T+1
* T+0

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

### Implementation

#### PipelinePack DataFlow

buffer size of PipelinePack

* EngineConfig
  - PoolSize
* Runner
  - PluginChanSize
* Router
  - PluginChanSize



                        -------<--------+
                        |               |
                        V               | generate pool
       EngineConfig.inputRecycleChan    | recycling
            |           |               |
            | is        +------->-------+
            |
    InputRunner.inChan
            |
            |     +--------------------------------------------------------+
    consume |     |                     Router.inChan                      |
            |     +--------------------------------------------------------+
          Input         ^           |               |                   ^
            |           |           | put           | put               |
            V           |           V               V                   |
            +-----------+  OutputRunner.inChan   FilterRunner.inChan    |
              inject                |               |                   |
                                    | consume       | consume           | inject
                                    V               V                   |
                                 Output           +------------------------+
                                                  |         Filter         |
                                                  +------------------------+


