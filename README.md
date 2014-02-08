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
    # install dependency: geoip go lib
    #========================================
    go get github.com/abh/geoip

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
*   cardinality statistics(for MAU alike counters where storing the data for statistics is prohibitive)
    
    In fact, if the data is stored only for the purpose of statistical calculations, incremental updates make storage unnecessary.
*   write decorated events to ElasticSearch(geoip, level range, del fields, auto sharding, currency convert)
*   ElasticSearch buffering(lessen uneccessary load of ES, e,g. dau, pv, hits)
*   behaviour db, dimensional funnel analysis(user based action series)
*   batch processing of historical logs
*   self monitoring
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


#### PipelinePack

Main pipeline data structure containing a AlsMessage and other metadata

##### buffer size of PipelinePack

* EngineConfig
  - PoolSize
* Runner
  - PluginChanSize
* Router
  - PluginChanSize


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
