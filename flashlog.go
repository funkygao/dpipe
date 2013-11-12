package main

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

func flashlogDataSource() {
    db, err := sql.Open("mysql", "flashlog:flashlog@unix(/var/run/mysqld/mysqld.sock)/flashlog?charset=utf8")
    if err != nil {
        panic(err)
        return
    }

    rows, err := db.Query("select * from log_us WHERE type=299 ORDER BY ID limit 10")
    if err != nil {
        panic(err)
    }

    for rows.Next() {
        fmt.Println(rows)
    }
}
