package main

import (
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"github.com/zhilyaev/gcp-serviceaccounts-exporter/pkg/version"
	"net/http"
	"os"
	"time"

	"github.com/helmwave/logrus-emoji-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/zhilyaev/gcp-serviceaccounts-exporter/pkg/collector"
	res "google.golang.org/api/cloudresourcemanager/v1"
)

const RootEnvVar = "GCP_SA_EXPORTER"

var (
	flagAddr            string
	flagDeltaDays       int
	flagLogFormat       string
	flagLogLevel        string
	flagParentID        string
	flagProjectID       string
	flagRefreshInterval time.Duration
)

var ctl = &cli.App{
	Name:        "gcp-serviceaccounts-exporter",
	Description: "GCP service accounts exporter",
	Version:     version.Version,
	Before: func(c *cli.Context) error {
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
			log.Error(err, "set debug log format")
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(lvl)
		}

		return nil
	},
	Commands: []*cli.Command{
		{
			Name:    "run",
			Aliases: []string{"start"},
			Usage:   "run exporter",
			Before: func(c *cli.Context) error {
				// Check flags
				if flagProjectID == "" && flagParentID == "" {
					return errors.New("you must specify project-id or parent-id")
				} else if flagProjectID != "" && flagParentID != "" {
					return errors.New("you must chose only project-id or parent-id")
				}

				return nil

			},
			Action: func(c *cli.Context) error {
				return run()
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "address",
					Value:       ":8080",
					Usage:       "Listen address",
					EnvVars:     []string{RootEnvVar + "_ADDR"},
					Destination: &flagAddr,
				},
				&cli.StringFlag{
					Name:        "project-id",
					Value:       "",
					Usage:       "GCP project ID",
					EnvVars:     []string{RootEnvVar + "_PROJECT_ID"},
					Destination: &flagProjectID,
				},
				&cli.StringFlag{
					Name:        "parent-id",
					Value:       "",
					Usage:       "Fetching projects by parent ID",
					EnvVars:     []string{RootEnvVar + "_PARENT_ID"},
					Destination: &flagParentID,
				},
				&cli.IntFlag{
					Name:        "days",
					Value:       90,
					Usage:       "Expired after N days",
					EnvVars:     []string{RootEnvVar + "_DELTA_DAYS"},
					Destination: &flagDeltaDays,
				},
				&cli.DurationFlag{
					Name:        "refresh-interval",
					Value:       30 * time.Second,
					Usage:       "Refresh interval",
					EnvVars:     []string{RootEnvVar + "_DELTA_DAYS"},
					Destination: &flagRefreshInterval,
				},
			},
		},
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Value:       "debug",
			Usage:       "debug | info | warn | trace",
			EnvVars:     []string{RootEnvVar + "_LOG_LVL"},
			Destination: &flagLogLevel,
		},
		&cli.StringFlag{
			Name:        "log-format",
			Value:       "emoji",
			Usage:       "emoji | pad | json | text",
			EnvVars:     []string{RootEnvVar + "_LOG_FORMAT"},
			Destination: &flagLogFormat,
		},
	},
}

func main() {
	err := ctl.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	projects, err := GetProjects(flagProjectID, flagParentID)
	if err != nil {
		return err
	}

	log.Debugf("Found projects: %d", len(projects))

	// Create Collector
	c := collector.New(flagRefreshInterval, flagDeltaDays, projects)
	go c.Run()

	// Expose /metrics HTTP endpoint using the created custom registry.
	http.Handle("/metrics", promhttp.Handler())
	log.Infof("Server is listening on %s", flagAddr)
	return http.ListenAndServe(flagAddr, nil)
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
