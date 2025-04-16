package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var codeRunCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ca_run_count",
	Help: "The number of code executions thare is started by a request",
}, []string{"project_uuid", "code_id"})

var codeRunElapsed = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "ca_run_elapsed",
	Help: "The time a run execution request took to complete",
}, []string{"project_uuid", "code_id"})

var codeCreatedCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ca_code_created_count",
	Help: "The number of code creations",
}, []string{"project_uuid", "code_id"})

var codeUpdatetdCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ca_code_updated_count",
	Help: "The number of code updates",
}, []string{"project_uuid", "code_id"})

func AddCodeRunCount(projectUUID string, codeID string, count float64) {
	codeRunCount.WithLabelValues(
		projectUUID, codeID,
	).Add(count)
}

func CodeRunElapsed(projectUUID string, codeID string, elapsed float64) {
	codeRunElapsed.WithLabelValues(
		projectUUID, codeID,
	).Observe(elapsed)
}

func AddCodeCreatedCount(projectUUID string, codeID string, count float64) {
	codeCreatedCount.WithLabelValues(
		projectUUID, codeID,
	).Add(count)
}

func AddCodeUpdatedCount(projectUUID string, codeID string, count float64) {
	codeUpdatetdCount.WithLabelValues(
		projectUUID, codeID,
	).Add(count)
}
