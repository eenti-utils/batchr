package batchr

type procTrig int

const (
	trgCap procTrig = iota
	trgInt
	trgStop
)
