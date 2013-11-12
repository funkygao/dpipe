package main

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

/*
+-------------+---------------------+------+-----+---------+----------------+
| Field       | Type                | Null | Key | Default | Extra          |
+-------------+---------------------+------+-----+---------+----------------+
| id          | bigint(20) unsigned | NO   | PRI | NULL    | auto_increment |
| uid         | bigint(20) unsigned | NO   | MUL | NULL    |                |
| type        | int(10) unsigned    | NO   | MUL | NULL    |                |
| data        | blob                | NO   |     | NULL    |                |
| ip          | bigint(20)          | NO   | MUL | NULL    |                |
| ua          | int(10) unsigned    | NO   | MUL | NULL    |                |
| date_create | int(10) unsigned    | NO   | MUL | NULL    |                |
+-------------+---------------------+------+-----+---------+----------------+
*/
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
