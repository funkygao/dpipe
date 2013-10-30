alser
=====

ALS guard

[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

### Architecture
    

          alser main()
              |
          loadJsonConfig
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


### Parsers

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
    each error msg has alarm threshold
    backoff email alarm/priority queue https://github.com/daviddengcn/go-villa
    IRC alarm
    flashlog alarm, data source

    each parser is controlled by configuration without any biz logic

