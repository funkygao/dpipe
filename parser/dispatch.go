package parser

func Dispatch(parserName, line string) {
	allParsers[parserName].ParseLine(line)
}
