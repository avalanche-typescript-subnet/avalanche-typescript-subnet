// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vm

import (
	"github.com/ava-labs/avalanchego/utils/metric"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	unitsVerified             prometheus.Counter
	unitsAccepted             prometheus.Counter
	txsSubmitted              prometheus.Counter // includes gossip
	txsVerified               prometheus.Counter
	txsAccepted               prometheus.Counter
	decisionsRPCConnections   prometheus.Gauge
	blocksRPCConnections      prometheus.Gauge
	rootCalculationWait       metric.Averager
	signatureVerificationWait metric.Averager
}

func newMetrics() (*prometheus.Registry, *Metrics, error) {
	r := prometheus.NewRegistry()

	rootCalculations, err := metric.NewAverager(
		"chain",
		"root_calculations",
		"time spent on root calculations in verify",
		r,
	)
	if err != nil {
		return nil, nil, err
	}
	signaturesVerified, err := metric.NewAverager(
		"chain",
		"signature_verification_wait",
		"time spent waiting for signature verification in verify",
		r,
	)
	if err != nil {
		return nil, nil, err
	}

	m := &Metrics{
		unitsVerified: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "chain",
			Name:      "units_verified",
			Help:      "amount of units verified",
		}),
		unitsAccepted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "chain",
			Name:      "units_accepted",
			Help:      "amount of units accepted",
		}),
		txsSubmitted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "vm",
			Name:      "txs_submitted",
			Help:      "number of txs submitted to vm",
		}),
		txsVerified: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "vm",
			Name:      "txs_verified",
			Help:      "number of txs verified by vm",
		}),
		txsAccepted: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "vm",
			Name:      "txs_accepted",
			Help:      "number of txs accepted by vm",
		}),
		decisionsRPCConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "vm",
			Name:      "decisions_rpc_connections",
			Help:      "number of open decisions connections",
		}),
		blocksRPCConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "vm",
			Name:      "blocks_rpc_connections",
			Help:      "number of open blocks connections",
		}),
		rootCalculationWait:       rootCalculations,
		signatureVerificationWait: signaturesVerified,
	}
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.unitsVerified),
		r.Register(m.unitsAccepted),
		r.Register(m.txsSubmitted),
		r.Register(m.txsVerified),
		r.Register(m.txsAccepted),
		r.Register(m.decisionsRPCConnections),
		r.Register(m.blocksRPCConnections),
	)
	return r, m, errs.Err
}
