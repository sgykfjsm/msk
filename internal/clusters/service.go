package clusters

import (
	"context"
	"fmt"
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

// FetchAndStoreClusters fetches clusters for the given project IDs and stores them in the database.
func (c *clusterService) FetchAndStoreClusters(ctx context.Context, projectIDs []string, pageSize int) (int, int, int, error) {
	var totalProcessedProjectNum, totalProcessedClusterNum, totalProcessedNodeNum int
	for _, projectID := range projectIDs {
		// 0. Reset the check counter every iteration
		processedClusterNum := 0
		page := 1

		for {
			// 1. Fetch cluster info using projectID
			clusters, total, err := c.fetcher.FetchClusters(ctx, projectID, page, pageSize)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("failed to fetch clusters for project %s with pageSize %d: %w", projectID, pageSize, err)
			}

			// 2. Store cluster info retrieving from TiDB Cloud API
			if err := c.store.StoreClusters(ctx, clusters); err != nil {
				return 0, 0, 0, fmt.Errorf("failed to store %d clusters for project %s with pageSize %d: %w", len(clusters), projectID, pageSize, err)
			}

			// 3. Store node info from clusters
			if processedNodeNum, err := c.store.StoreClusterNodes(ctx, clusters); err != nil {
				return 0, 0, 0, fmt.Errorf("failed to store cluster nodes for project %s with pageSize %d: %w", projectID, pageSize, err)
			} else {
				totalProcessedNodeNum += processedNodeNum
			}

			// 4. If the number of processed cluster is less than total, repeat the process.
			processedClusterNum += len(clusters)
			if processedClusterNum >= total {
				break
			}

			// 4. Increment the page number to fetch the next set of clusters
			page++
		}
		totalProcessedProjectNum++
		totalProcessedClusterNum += processedClusterNum
	}

	return totalProcessedProjectNum, totalProcessedClusterNum, totalProcessedNodeNum, nil
}
