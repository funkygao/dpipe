alser
=====

ALS guard

[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

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
    alser should wait for alsparser collectAlarm done before exit
    each error msg has alarm threshold
    auto restart when more files or few files appear
    backoff email alarm
    fsmon
    main wait for all parsers collect alarm done
