alser
=====

ALS guard

[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

### Architecture
    

             main
              |
              |<-------------------------------
              |                                |
              | goN(wait group)           -----------------
              V                          | alarm collector |
        -----------------------           -----------------
       |       |       |       |               |
      log1    log2    ...     logN             |
       |       |       |       |               | alarm
        -----------------------                | chan
              |                                |
              | feed lines                     |
              V                                |
        -----------------------                ^
       |       |       |       |               |
     parser1 parser2  ...   parserM            |
       |       |       |       |               |
        -----------------------                |
              |                                |
               ------------------->------------

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
