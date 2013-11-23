package main

type Indexer struct {
	lineIn chan string
}

func newIndexer() (this *Indexer) {
	this = new(Indexer)
	this.lineIn = make(chan string, 1000)

	go this.mainLoop()

	return
}

func (this *Indexer) mainLoop() {
	for line := range this.lineIn {
		this.doIndex(line)
	}
}

func (this *Indexer) doIndex(line string) {

}

func (this *Indexer) feedLine(line string) {
	this.lineIn <- line
}
