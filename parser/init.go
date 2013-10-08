package parser

func init() {
    allParsers = make(map[string]Parser)

    allParsers["MemcacheFailParser"] = newMemcacheFailParser()
	allParsers["ErrorLogParser"] = newErrorLogParser()
	allParsers["PaymentParser"] = newPaymentParser()
}
