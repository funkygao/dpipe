alser
=====

ALS guard

Distributed log collector push data to ALS center server where alser will run.

alser keeps eyes on events from logs(defined in conf file) and has rule based engine
to send alarms via beep/email/IRC/etc.

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


#### Process

          alser main()
              |
          LoadConfig
              |
              |<-------------------------------
              |                                |
              | goN(wait group)                |
              | each log a worker         -----------------
              V                          | alarm collector |
              |                          |    watchdog     |
        -----------------------           -----------------
       |       |       |       |               |
      log1    log2    ...     logN             |
      worker  worker  ...     worker           |
       |       |       |       |               | send alarm
        -----------------------                | 
              |                                |
              | feed lines                     | TODO(backoff alarms)
              V                                | 
        -----------------------                ^
      parser is shared among logs              |
        -----------------------                |
       |       |       |       |               |
     parser1 parser2  ...   parserM            |
       |       |       |       |               |
        -----------------------                |
              |                                |
          --------------------                 |
         |                    |                |
      handle alarm            ---->------------


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


### TODO

    alarm according to history behavior
    backoff email alarm/priority queue https://github.com/daviddengcn/go-villa
    IRC alarm channel
    flashlog alarm, data source

    mongo unit test
