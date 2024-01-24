package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	res "google.golang.org/api/cloudresourcemanager/v1"
)

const MetricName = "gcp_service_accounts_expired_keys"

type Collector struct {
	refreshInterval time.Duration
	ctx             context.Context
	metric          *prometheus.GaugeVec
	projects        []*res.Project
	deltaDays       int
}

func New(refreshInterval time.Duration, deltaDays int, projects []*res.Project) *Collector {
	// Metric definitions
	metric := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricName,
			Help: "Expired keys",
		},
		// The label names by which to split the metric.
		[]string{"key"},
	)

	return &Collector{
		ctx:             context.TODO(), // Define something else to handle cancellation
		metric:          metric,
		projects:        projects,
		deltaDays:       deltaDays,
		refreshInterval: refreshInterval,
	}
}

func (c *Collector) Run() {
	// Registration metric
	prometheus.MustRegister(c.metric)

	// Update earliest
	c.update()

	ticker := time.NewTicker(c.refreshInterval)
	for range ticker.C {
		c.update()
	}
}

func (c *Collector) update() {
	log.Infof("Update metrics in %s...", c.refreshInterval)

	for _, project := range c.projects {
		go func(project *res.Project) {
			l := log.WithField("project", project.Name)
			l.Info("Update has been started")

			// Get keys for project
			keys, err := GetExpiredKeys(c.ctx, project.ProjectId, c.deltaDays)
			if err != nil {
				log.Fatalln("Can't get keys from google api: ", err)
			}

			// Render keys via prometheus metrics
			for key, days := range keys {
				l.WithFields(log.Fields{
					"key":  key.Name,
					"days": days,
				}).Debugf("Rendering metric")
				c.metric.WithLabelValues(key.Name).Set(float64(days))
			}
		}(project)
	}
}
