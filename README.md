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


### Dependencies

    go get github.com/bmizerany/assert
    go get github.com/bitly/go-simplejson
    go get github.com/daviddengcn/go-ljson-conf
    go get github.com/mattn/go-sqlite3
    go get github.com/funkygao/tail
    go get github.com/funkygao/gofmt
    go get github.com/funkygao/gotime
    go get -u github.com/funkygao/alser

### TODO

    alarm according to history behavior
    each error msg has alarm threshold
    backoff email alarm
    IRC alarm
    when only 1 parser, wait never return
    [15310]2013/10/27 07:43:13 all workers finished
    [15310]2013/10/27 07:43:13 stopping all parsers...
    [15310]2013/10/27 07:43:13 waiting all parsers...
