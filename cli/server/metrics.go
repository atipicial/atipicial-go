package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

var atipicialgoVersion = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Help:      "AtipicialGo version",
		Name:      "version",
		Namespace: "atipicialgo",
	},
	[]string{"version"})

func setAtipicialGoVersion(nodeVer string) {
	atipicialgoVersion.WithLabelValues(nodeVer).Add(1)
}

func init() {
	prometheus.MustRegister(
		atipicialgoVersion,
	)
}
