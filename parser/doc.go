/*
Parsers for any kind of ALS log contents.

Luckily, all ALS log entry has the same format:
	area,timestamp,json

Each parser is dedicated for one purpose of logging guard, which
has 2 major jobs:
	parse line
		one line each time
	metric
		can hold for a period of time
 */
package parser
