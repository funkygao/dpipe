package main

func init() {
	Parsers["DefaultParser"] = &DefaultParser{name: "DefaultParser"}
}
