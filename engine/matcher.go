package engine

type Matcher struct {
	runner  FilterOutputRunner
	matches []string
}

func NewMatchRunner(matches []string, r FilterOutputRunner) *Matcher {
	this := new(Matcher)
	this.matches = matches
	this.runner = r
	return this
}

func (this *Matcher) InChan() chan *PipelinePack {
	return this.runner.InChan()
}

func (this *Matcher) match(pack *PipelinePack) bool {
	if len(this.matches) == 0 {
		// match all
		return true
	}

	for _, match := range this.matches {
		if pack.Ident == match {
			return true
		}
	}

	return false
}
