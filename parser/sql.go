package parser

const (
	// error log
	ERRLOG_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS error (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	cls VARCHAR(50),
    level CHAR(20),
    msg VARCHAR(200) NULL,
    flash INT
);
`
	ERRLOG_CREATE_INDEX = `
CREATE INDEX IF NOT EXISTS err_idx ON error(ts, cls);
`
	ERRLOG_INSERT = "INSERT INTO error(area, ts, cls, level, msg, flash, host) VALUES(?,?,?,?,?,?,?)"

	// payment log
	PAYMENT_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS payment (
	area CHAR(10),
	host CHAR(20),
	ts INT,
	type VARCHAR(50),
    uid INT(10) NULL,
    level INT,
    amount INT,
    ref VARCHAR(50) NULL,
    item VARCHAR(40),
    currency VARCHAR(20)
);
`
	PAYMENT_CREATE_INDEX = `
CREATE INDEX IF NOT EXISTS pay_idx ON payment(ts, type, area);
`
	PAYMENT_INSERT = "INSERT INTO payment(area, host, ts, type, uid, level, amount, ref, item, currency) VALUES(?,?,?,?,?,?,?,?,?,?)"

	// slowresp log
	SLOWRESP_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS slowresp (
	area CHAR(10),
	host CHAR(20),
	uri VARCHAR(50),
	ts INT,
	req_t INT,
	db_t INT	
);
`
	SLOWRESP_INSERT = `INSERT INTO slowresp(area, host, uri, ts, req_t, db_t) VALUES(?,?,?,?,?,?)`

	// level up log
	LEVELUP_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS levelup (
	area CHAR(10),	
	ts INT,
	fromlevel INT
);
`
	LEVELUP_INSERT = `INSERT INTO levelup(area, ts, fromlevel) VALUES(?,?,?)`

	PHPERROR_CREATE_TABLE = `
CREATE TABLE IF NOT EXISTS phperror (
	area CHAR(10),	
	ts INT,
	host CHAR(25),
	level CHAR(15),
	src_file VARCHAR(80),
	msg VARCHAR(100)
);
`
	PHPERROR_INSERT = `INSERT INTO phperror(area, ts, host, level, src_file, msg) VALUES(?,?,?,?,?,?)`
)
