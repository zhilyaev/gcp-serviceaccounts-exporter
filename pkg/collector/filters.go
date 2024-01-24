package collector

import (
	"context"
	"fmt"
	"time"

	res "google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
)

// GetExpiredKeys will return expired keys
func GetExpiredKeys(ctx context.Context, projectID string, delta int) (expired map[*iam.ServiceAccountKey]int, err error) {
	expired = make(map[*iam.ServiceAccountKey]int)

	iamService, err := iam.NewService(ctx)
	if err != nil {
		return nil, err
	}

	// Get projects via http request to api
	list, err := iamService.Projects.ServiceAccounts.List("projects/" + projectID).Do()
	if err != nil {
		return nil, err
	}

	for _, sa := range list.Accounts {
		keyListUrl := fmt.Sprintf("projects/%s/serviceAccounts/%s", projectID, sa.UniqueId)
		// // Get serviceAccounts via http request to api
		response, err := iamService.Projects.ServiceAccounts.Keys.List(keyListUrl).Do()
		if err != nil {
			return nil, err
		}

		for _, key := range response.Keys {
			now := time.Now()
			suspected, _ := time.Parse(time.RFC3339, key.ValidAfterTime)
			diff := now.Sub(suspected)
			days := int(diff.Hours() / 24)

			// Filter older than delta
			if days > delta {
				expired[key] = days
			}

		}
	}

	return expired, nil
}

// GetProjects returns a list of projects filtered by the labels
func GetProjects(ctx context.Context, parentID string) (projects []*res.Project, err error) {
	rms, err := res.NewService(ctx)
	if err != nil {
		return nil, err
	}

	list, err := rms.Projects.List().Do()
	if err != nil {
		return nil, err
	}

	for _, project := range list.Projects {
		if project.Parent.Id == parentID {
			projects = append(projects, project)
		}
	}

	return projects, nil
}
