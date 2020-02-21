package ckb

import "github.com/netdata/go-orchestrator/module"

type (
	Charts = module.Charts
	Dims   = module.Dims
)

const (
	Reorg             = "reorg"
	FreshTransactions = "fresh_transactions"
	FreshUncles       = "fresh_uncles"
	Error             = "error"
	Warn              = "warning"
)

var charts = Charts{
	// TODO What is Fam?
	{
		ID:    "chain",
		Title: "chain", Units: "events/s", Fam: "chain",
		Dims: Dims{
			{ID: Reorg, Name: Reorg, Algo: module.Absolute},
		},
	},
	{
		ID:    "fresh",
		Title: "fresh", Units: "events/s", Fam: "fresh",
		Dims: Dims{
			{ID: FreshTransactions, Name: FreshTransactions, Algo: module.Absolute},
			{ID: FreshUncles, Name: FreshUncles, Algo: module.Absolute},
		},
	},
	{
		ID:    "notice",
		Title: "notice", Units: "events/s", Fam: "notice",
		Dims: Dims{
			{ID: Error, Name: Error, Algo: module.Absolute},
			{ID: Warn, Name: Warn, Algo: module.Absolute},
		},
	},
}
