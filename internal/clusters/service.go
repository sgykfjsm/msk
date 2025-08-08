package clusters

import (
	"context"
)

type ClusterService interface {
	FetchAndStoreClusters(ctx context.Context, projectIDs []string, pageSize int) error
}

type clusterService struct {
	fetcher ClusterFetcher
	store   ClusterStore
}

func NewClusterService(fetcher ClusterFetcher, store ClusterStore) *clusterService {
	return &clusterService{
		fetcher: fetcher,
		store:   store,
	}
}

func (c *clusterService) FetchAndStoreClusters(ctx context.Context, projectIDs []string, pageSize int) error {
	for _, projectID := range projectIDs {
		// 0. Reset the check counter every iteration
		processedClusterNum := 0
		page := 1

		for {
			// 1. Fetch cluster info using projectID
			clusters, total, err := c.fetcher.FetchClusters(ctx, projectID, page, pageSize)
			if err != nil {
				return err
			}

			// 2. Store cluster info retrieving from TiDB Cloud API
			if err := c.store.StoreClusters(ctx, clusters); err != nil {
				return err
			}

			// 3. If the number of processed cluster is less than total, repeat the process.
			processedClusterNum += len(clusters)
			if processedClusterNum >= total {
				break
			}

			// 4. Increment the page number to fetch the next set of clusters
			page++
		}
	}

	return nil
}
