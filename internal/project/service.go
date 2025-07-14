package project

import "context"

// ProjectService orchestrates the fetching and storing of project data.
// It uses the ProjectFetcher to retrieve projects and then processes them as needed.
type ProjectService struct {
	fetcher ProjectFetcher
	store   ProjectStore // This will be used in future iterations to store projects in a database.
}

// NewProjectService creates a new ProjectService with the given ProjectFetcher.
func NewProjectService(fetcher ProjectFetcher, store ProjectStore) *ProjectService {
	return &ProjectService{
		fetcher: fetcher,
		store:   store,
	}
}

// FetchAndStoreProjects fetches projects using the ProjectFetcher and processes them.
// It returns an error if the fetching or processing fails.
func (s ProjectService) FetchAndStoreProjects(ctx context.Context, page int, pageSize int) error {
	var processedProjectsNum int

	for {
		projects, totalProjectNum, err := s.fetcher.FetchProjects(ctx, page, pageSize)
		if err != nil {
			return err
		}

		if err := s.store.StoreProjects(ctx, projects); err != nil {
			return err
		}

		processedProjectsNum += len(projects)
		if processedProjectsNum >= totalProjectNum {
			break
		}
		page++ // Increment the page number to fetch the next set of projects
	}

	return nil
}
