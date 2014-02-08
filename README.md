dpipe - Distributed Data Pipeline
=================================
It's sentry+logstash+flunted+splunk.

         _       _           
        | |     (_)           
      __| |_ __  _ _ __   ___  
     / _  | '_ \| | '_ \ / _ \
    | (_| | |_) | | |_) |  __/ 
     \__,_| .__/|_| .__/ \___|
          | |     | |
          |_|     |_|
            
[![Build Status](https://travis-ci.org/funkygao/dpipe.png?branch=master)](https://travis-ci.org/funkygao/dpipe)

### Features

*   engine + plugins, driven by configuration
*   multi-tenant
*   data(log events) pipeline by design, meets most data processing requirement
*   because of the engine design, it's very easy to create a new plugin to meet more requirements
*   visualization of data pipeline path
    - it's easy make complex of your config file, so visualization is a great help
*   implementation highlights
    - high(universal) abstraction of data processing as input -> codec -> filter -> output
    - plugin design for extenstion
    - reference counter based recyle channel buffer to lessen golang GC
    - shared memory with copy on write mechanism
    - rich self monitoring/diagnostic interface
    - high performance routing
    - thanks to golang channel, self-healing when input/output speed doesn't match without message queue
    - most key checkpoint was under benchmark test and unit test

### Install

    #========================================
    # install dependency: geoip c lib
    #========================================
    git clone https://github.com/maxmind/geoip-api-c.git
    cd geoip-api-c/
    ./bootstrap
    ./configure
    make
    make install

    #========================================
    # install dpipe
    #========================================
    go get github.com/funkygao/dpipe
    cd $GOPATH/src/github.com/funkygao/dpipe
    go get
    ./build.sh
    ./cmd/dpiped/dpiped -conf etc/engine.als.cf

### Currently Supported Plugins

*   slide window based streaming biz alarm
    - console beep(let you know instantly)
    - colored log(different color represent different kind of event you specified) 
    - alert email(aggregation, it is basically a priority queue)
*   cardinality statistics(for MAU alike counters where storing the data for statistics is prohibitive)

    In fact, if the data is stored only for the purpose of statistical calculations, incremental updates makes storage unnecessary.
*   ElasticSearch feeding
    - feed decorated events to ElasticSearch
        - auto ES sharding by date/week/month
        - geoip filling
        - user level range transformation
        - delete unwanted fields
        - currency convert
        - and more
    - ElasticSearch buffering(lessen uneccessary load of ES, e,g. dau, pv, hits)
*   behaviour db integration for dimensional funnel analysis(user/time based action series)
*   batch processing of historical logs
    - some data does not need instant(latency within a second) processing
    - just want to have snapshot
*   tcp receiver/sender for hierarchy deployment
*   to be more...

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
                                    V
                            +-----------------+
                            |   ALS Server    |
                            |-----------------| 
                            |   log files     |
                            |-----------------| 
                            |     dpiped      |
                            +-----------------+
                                        |
                                        |<----------------------------------------------------------------------+
                                        |                                                                       |
                                        | Input-Decode-Filter-Output                                            |
                                        V                                                                       |
                                   +----------------------------------------------------------------+           |
                                   |                   |           |           |           |        |           |
                              realtime analysis     indexer     archive    BehaviorDB      S3   hierarchy       |
                                   |                   |           |           |           |    deployment      |
            +----------------------|                   |           |           |           |        |           |
            |statistics            |alarming           |           |           |           |        |           |
       +----------+       +-----------------+          |           |           |           |        |           |
       |          |       |    |     |      |   ElasticSearch    HDFS      LevelDB/sky   RedShift  dpipe        |
     quantile   hyper     |    |   color    |          |           |           |           |        |           |
    histogram  loglog   beep email console etc         |           |           |           |        |           |
      topN        |       |    |     |      |          |           |      Dimensional      |        +-----------+
       |          |       +-----------------+       Kibana3        |    FunnelAnalysis   tableau
       +----------+                |                   |           |           |           |
            |                      |                   |           |           |           |
          PM/dev                dev/ops               PM          ops         dev         PM



### Implementation

#### Overview



                                             -- predict ----
                   (slide win)              |               |
    Input -> Filter(transform) -> Output -> |-- store ------| -> visualization
                   (cleaness)               |               | 
                   (decorator)              |-- explore ----|
                   (buffering)              |               |
                   (streaming)               -- alarm ------


##### data flow

                            -------<-------- 
                            |               |
                            V               | generate pool
           EngineConfig.inputRecycleChan    | recycling
                |           |               |
                | is         ------->------- 
                |
        InputRunner.inChan
                |
                |     +--------------------------------------------------------+
        consume |     |                     Router.hub                         |
                V     +--------------------------------------------------------+
              Input         ^           |               |                   ^
                |           |           | put           | put               |
                |           |           |               |                   |
                 ----->-----       Matcher.inChan   Matcher.inChan          |
                   inject               ^               ^                   |
                                        | is            | is                |
                                        V               V                   |
                               OutputRunner.inChan   FilterRunner.inChan    |
                                        |               |                   |
                                        | consume       | consume           | inject
                                        V               V                   |
                                     Output           +------------------------+
                                                      |         Filter         |
                                                      +------------------------+
    
   
##### shutdown


        engine SIGINT
          |
        http.Stop
          |
        all Input.Stop()
          |
        router ----- close filterRunner.inChan --- Filter stopped
          |     |
          |      --- close outRunner.inChan ------ Output stopped
          |
        router ----- wait for FO runner finish --- close router.hub
          |
        done



##### performance

        tail.readLine 5405 ns/op
          |
          |          26242 ns/op
        msg.FromLine ----------- parse line     228 ns/op
          |                 |
          |                  --- jsonize        25179 ns/op
          |
          | chan 109 ns/op
          |
          |         5500 ns/op
        filters ---------------- CamelCaseName  86.4 ns/op
          |                 |
          |                 |--- RegexMatch     1570 ns/op
          |                 |
          |                 |--- IndexName      1029 ns/op
          |                 |
          |                 |--- FieldValue     455  ns/op
          |                 |
          |                 |--- SetField       116  ns/op
          |                 |
          |                  --- PackCopy       3438 ns/op
          |
          |
        output ---------------- UUID            1622 ns/op
                            |
                             --- MarshalPayload 16673 ns/op
