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

// Worker Pool Metrics - Gauges
var (
	workerpoolWorkersTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "workerpool_workers_total",
		Help: "Total number of workers in the pool",
	})

	workerpoolWorkersBusy = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "workerpool_workers_busy",
		Help: "Number of workers currently executing tasks",
	})

	workerpoolQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "workerpool_queue_size",
		Help: "Current number of tasks waiting in queue",
	})

	workerpoolQueueCapacity = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "workerpool_queue_capacity",
		Help: "Maximum capacity of the task queue",
	})
)

// Worker Pool Metrics - Counters
var (
	workerpoolTasksSubmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "workerpool_tasks_submitted_total",
		Help: "Total number of tasks submitted to the pool",
	})

	workerpoolTasksCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "workerpool_tasks_completed_total",
		Help: "Total number of tasks completed successfully",
	})

	workerpoolTasksFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "workerpool_tasks_failed_total",
		Help: "Total number of tasks that failed with error",
	})

	workerpoolTasksRejected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "workerpool_tasks_rejected_total",
		Help: "Total number of tasks rejected due to full queue",
	})

	workerpoolTasksTimeout = promauto.NewCounter(prometheus.CounterOpts{
		Name: "workerpool_tasks_timeout_total",
		Help: "Total number of tasks cancelled due to context timeout",
	})
)

// Worker Pool Metrics - Histograms
var (
	workerpoolQueueWait = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "workerpool_queue_wait_seconds",
		Help:    "Time tasks spend waiting in queue before execution",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	})

	workerpoolTaskDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "workerpool_task_duration_seconds",
		Help:    "Duration of task execution",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
	})
)

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

// Worker Pool Metric Functions - Gauges
func SetWorkerpoolWorkersTotal(count float64)  { workerpoolWorkersTotal.Set(count) }
func SetWorkerpoolWorkersBusy(count float64)   { workerpoolWorkersBusy.Set(count) }
func SetWorkerpoolQueueSize(count float64)     { workerpoolQueueSize.Set(count) }
func SetWorkerpoolQueueCapacity(count float64) { workerpoolQueueCapacity.Set(count) }
func IncWorkerpoolWorkersBusy()                { workerpoolWorkersBusy.Inc() }
func DecWorkerpoolWorkersBusy()                { workerpoolWorkersBusy.Dec() }

// Worker Pool Metric Functions - Counters
func IncWorkerpoolTasksSubmitted() { workerpoolTasksSubmitted.Inc() }
func IncWorkerpoolTasksCompleted() { workerpoolTasksCompleted.Inc() }
func IncWorkerpoolTasksFailed()    { workerpoolTasksFailed.Inc() }
func IncWorkerpoolTasksRejected()  { workerpoolTasksRejected.Inc() }
func IncWorkerpoolTasksTimeout()   { workerpoolTasksTimeout.Inc() }

// Worker Pool Metric Functions - Histograms
func ObserveWorkerpoolQueueWait(seconds float64)    { workerpoolQueueWait.Observe(seconds) }
func ObserveWorkerpoolTaskDuration(seconds float64) { workerpoolTaskDuration.Observe(seconds) }
