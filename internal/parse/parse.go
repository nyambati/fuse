package parse

import (
	"github.com/nyambati/fuse/internal/am"
	"github.com/nyambati/fuse/internal/diag"
	"github.com/nyambati/fuse/internal/dsl"
	"github.com/nyambati/fuse/internal/secrets"
)

// ToAlertmanager translates a loaded Fuse project into an Alertmanager config.
//
// MVP steps:
//  1. Build Receivers from channels
//  2. Build Routes from flows (attached under root route)
//  3. Build TimeIntervals from silence_windows
//  4. Inhibit rules are passed through as-is from DSL (v0.1 simple copy)
func ToAlertmanager(proj dsl.Project, prov secrets.Provider) (am.Config, []diag.Diagnostic) {
	var (
		cfg   am.Config
		diags []diag.Diagnostic
	)

	// Receivers
	recvs, rDiags := BuildReceivers(proj)
	if len(rDiags) > 0 {
		diags = append(diags, rDiags...)
	}

	cfg.Receivers = recvs

	// // Routes (flows)
	rootRoute, fDiags := BuildFlowRoutes(proj)
	if len(fDiags) > 0 {
		diags = append(diags, fDiags...)
	}

	cfg.Route = rootRoute

	// TimeIntervals (silence windows)
	intervals, tDiags := BuildTimeIntervals(proj)
	if len(tDiags) > 0 {
		diags = append(diags, tDiags...)
	}
	cfg.TimeIntervals = intervals

	// InhibitRules — copy directly from DSL
	for _, ir := range proj.Inhibitors {
		sourceMatchers, _ := ToMatchers(ir.If)
		targetMatchers, _ := ToMatchers(ir.Suppress)
		cfg.InhibitRules = append(cfg.InhibitRules, am.InhibitRule{
			SourceMatchers: sourceMatchers,
			TargetMatchers: targetMatchers,
			Equal:          ir.When,
		})
	}

	// Global config (from DSL global section) — MVP: straight copy
	cfg.Global = proj.Global

	// TODO (later): secrets resolution via prov
	_ = prov // currently unused until secrets resolution is added

	return cfg, diags
}
