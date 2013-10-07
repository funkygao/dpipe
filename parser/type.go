package parser

type Parser interface {
	parseLine(line string)
	getStat(duration int32)
}
