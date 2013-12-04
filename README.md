alser
=====

ALS guard

Distributed log collector push data to ALS center server where alser will run.

alser keeps eyes on events from logs(defined in conf file) and has rule based engine
to send alarms via beep/email/IRC/etc.

*   als is aimed to be swiss knife kind tool instead of a complex system.
*   als is different from splunk which is basically a search engine.
*   als is different from opentsdb which is metrics based while als is event based.


[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

### Install

    go get github.com/funkygao/alser
    alser -h # help

### Architecture

#### Deployment

        +---------+     +---------+     +---------+     +---------+
        | server1 |     | server2 |     | server3 |     | serverN |
        +---------+     +---------+     +---------+     +---------+
            |               |               |               |
             -----------------------------------------------
                                    |
                                    | push log events
                                    |
                            +-----------------+
                            |   ALS Server    |
                            |-----------------| 
                            | alser daemon    |
                            +-----------------+
                                    |
                                    | send alarm
                                    |
                   +------------------------------------+
                   |            |           |           |
                 beep         email       IRC          etc


#### Overview

          alser main()
              |
          LoadConfig
              |
              |-----------------------------------------------------------------
              |                                                                 |
              | goN(wait group)                                                 |
              | each datasource a worker                                -----------------
              V                                                        | alarm collector |
          DataSource                                                   |    watchdog     |
              |                                                         ----------------- 
           ------------------------------------------                           |
          |                                          |                          |
          | logfile                                  | db                       |
          |                                          |                          |
        -----------------------         -----------------------                 |
       |       |       |       |       |       |        |      |                |
      log1    log2    ...     logN    table1  table2   ...  tableN              |
      worker  worker  ...     worker   |       |        |      |                |
       |       |       |       |       |       |        |      |                |
        -----------------------         -----------------------                 |
              |                                   |                             |
               -----------------------------------                              | feed
                    |                                                           | alarms
                    | feed lines                                                ^
                    V                                                           |
        -----------------------                                                 |
       |       |       |       |                                                |
     parser1 parser2  ...   parserM                                             |
       |       |       |       |                                                |
        -----------------------                                                 |
                    |                                                           |
          --------------------                                                  |
         |                    |                                                 |
      handle alarm            ---->---------------------------------------------
         |
        ------------
       |     |      |
     email console IRC

#### Parsers

        log1  log2  ...  logN
         |     |          |
          ----------------
                |
                | log content line by line
                |
              parser
                |
            --------------------------------
           |                                |
           |                        go collectAlarms
           |                                |                every N seconds
       parse line into columns       +--------------------------------------+
           |                         | got checkpoint period                |
           |                         | group by(chkpoint period) order desc |
       insert into sqlite3(columns)  | delete from db(chkpoint period)      |
           |                         | send alarm(console/email/irc)        |
           |                         +--------------------------------------+
           |                                ^
           V                                |
            --------------------------------
                          |
                    +----------------+
                    | sqlite DB file |
                    +----------------+


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
                                |                           文件落地
                                |-----------------------------------
             RealTimeAnalysis   |                                   |
             -----------------------------------            +-------------------------+
            |           |           |           |           | central logfile cluster |
        +-------+   +-------+   +-------+   +-------+       +-------------------------+
        |secure |   | debug |   |monitor|   |analyse|
        +-------+   +-------+   +-------+   +-------+


### TODO

    priority queue https://github.com/daviddengcn/go-villa
    abnormal change LRU in case of OOM
    websocket for web based alarm
    lua integration to parse complex log

    confirm session/newaccount file format
    geodbfile   indexer.domain

    es grant access rights

