package ckb

import "github.com/netdata/go-orchestrator/module"

type (
	Charts = module.Charts
	Dims   = module.Dims
)

const (
	Reorg	= "reorg"
	FreshTransactions = "fresh_transactions"
	FreshUncles = "fresh_uncles"
	Error = "error"
	Warn = "warn"
)

var charts = Charts{
	// TODO What is Fam?
	{
		ID:    "ckb-chain",
		Title: "ckb-chain", Units: "events/s", Fam: "ckb-chain",
		Dims: Dims {
			{ID: Reorg , Name: Reorg, Algo: module.Absolute},
		},
	},
	{
		ID:    "ckb-fresh",
		Title: "ckb-fresh", Units: "events/s", Fam: "ckb-fresh",
		Dims: Dims {
			{ID: FreshTransactions, Name: FreshTransactions, Algo: module.Absolute},
			{ID: FreshUncles, Name: FreshUncles, Algo: module.Absolute},
		},
	},
	{
		ID:    "ckb-notice",
		Title: "ckb-notice", Units: "events/s", Fam: "ckb-notice",
		Dims: Dims {
			{ID: Error, Name: Error, Algo: module.Absolute},
			{ID: Warn, Name: Warn, Algo: module.Absolute},
		},
	},
}
