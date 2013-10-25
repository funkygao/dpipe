package parser

// Dispatch a line of log entry to target parser by name
func Dispatch(parserName, line string) {
    p, ok := GetParser(parserName)
    if !ok {
        return
    }

    p.ParseLine(line)
}

// Get a parser instance by name
func GetParser(parserName string) (p Parser, ok bool) {
    p, ok = allParsers[parserName]
    return
}
