package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/helmwave/logrus-emoji-formatter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/zhilyaev/gcp-serviceaccounts-exporter/pkg/collector"
	res "google.golang.org/api/cloudresourcemanager/v1"
)

var (
	flagAddr            string
	flagDeltaDays       int
	flagLogFormat       string
	flagLogLevel        string
	flagParentID        string
	flagProjectID       string
	flagRefreshInterval time.Duration
)

func init() {
	flag.StringVar(&flagAddr, "address", ":8080", "Listen address")
	flag.StringVar(&flagLogFormat, "log-format", "emoji", "emoji | pad | json | text ")
	flag.StringVar(&flagLogLevel, "log-level", "debug", "debug | info | warn | trace ")
	flag.StringVar(&flagProjectID, "project-id", "", "GCP project ID")
	flag.StringVar(&flagParentID, "parent-id", "", "Fetching projects by parent ID")
	flag.IntVar(&flagDeltaDays, "days", 90, "Expired after N days")
	flag.DurationVar(&flagRefreshInterval, "refresh-interval", 30*time.Second, "refresh interval")

	flag.Parse()

	switch flagLogFormat {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			PrettyPrint: true,
		})
	case "pad":
		log.SetFormatter(&log.TextFormatter{
			PadLevelText: true,
			ForceColors:  false,
		})
	case "emoji":
		log.SetFormatter(&formatter.Config{
			Color: true,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			ForceColors: false,
		})
	}

	lvl, err := log.ParseLevel(flagLogLevel)
	if err != nil {
		log.Error(err)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(lvl)
	}

	if flagProjectID == "" && flagParentID == "" {
		log.Fatal("You must specify project-id or parent-id")
	} else if flagProjectID != "" && flagParentID != "" {
		log.Fatal("You must chose only project-id or parent-id")
	}

}

func main() {
	projects, err := GetProjects(flagProjectID, flagParentID)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Projects: %d", len(projects))

	// Create Collector
	c := collector.New(flagRefreshInterval, flagDeltaDays, projects)
	go c.Run()
	// TODO: implement pretty cancellation for Collector

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Server is listening on %s", flagAddr)
	log.Fatalln(http.ListenAndServe(flagAddr, nil))
}

// GetProjects is shortcut around []*res.Project
func GetProjects(projectID, parentID string) (projects []*res.Project, err error) {
	if projectID != "" {
		projects = append(projects, &res.Project{ProjectId: projectID})
	} else if parentID != "" {
		projects, err = collector.GetProjects(context.Background(), parentID)
		if err != nil {
			return nil, err
		}
	}

	return projects, err
}

// GetProjectsIDs not used anymore
//func GetProjectsIDs(projectID, parentID string) (ids []string, err error) {
//	if projectID != "" {
//		ids = append(ids, projectID)
//	} else if parentID != "" {
//		projects, err := collector.GetProjects(context.Background(), parentID)
//		if err != nil {
//			return nil, err
//		}
//
//		for _, project := range projects {
//			ids = append(ids, project.ProjectId)
//		}
//	}
//
//	return nil, err
//}
