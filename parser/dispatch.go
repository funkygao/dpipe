package parser

func Dispatch(parserName, line string) {
    allParsers[parserName].ParseLine(line)
}

func GetParser(parserName string) Parser {
    return allParsers[parserName]
}
