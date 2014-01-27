package main

import (
	"github.com/funkygao/GoStats"
	"github.com/funkygao/bayesian"
)

const (
	PayUser  bayesian.Class = "pay"
	FreeUser bayesian.Class = "free"
)

func classify() {
	classifier := bayesian.NewClassifier(PayUser, FreeUser)
	classifier.Learn(document, which)
	classifier.LogScores(document)

}
