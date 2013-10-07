package parser

import (
	//"reflect"
)

func Dispatch(parserName, line string) {
	allParsers[parserName].ParseLine(line)
}
