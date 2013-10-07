package parser

func init() {
    allParsers = make(map[string]Parser)

    allParsers["DefaultParser"] = DefaultParser{}
    allParsers["MemcacheFailParser"] = MemcacheFailParser{}
}

