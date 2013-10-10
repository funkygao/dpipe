alser
=====

ALS guard

[![Build Status](https://travis-ci.org/funkygao/alser.png?branch=master)](https://travis-ci.org/funkygao/alser)

### Dependencies

    go get github.com/bmizerany/assert
    go get github.com/bitly/go-simplejson
    go get github.com/mattn/go-sqlite3
    go get github.com/funkygao/tail
    go get github.com/funkygao/gofmt
    go get github.com/funkygao/alsparser
    go get -u github.com/funkygao/alser

### TODO

    more readable alarm screen output
    parsers can be configured, each parser one conf
    MongoException guard
    each parser should have cleanup, such as close db
    alarm according to history behavior
