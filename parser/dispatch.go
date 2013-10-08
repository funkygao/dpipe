package parser

// Dispatch a line of log entry to target parser by name
func Dispatch(parserName, line string) {
	GetParser(parserName).ParseLine(line)
}

// Get a parser instance by name
func GetParser(parserName string) Parser {
    return allParsers[parserName]
}
