package clusters

import (
	"context"
	"fmt"
	"time"
)

type ClusterService interface {
	FetchAndStoreClusters(ctx context.Context, projectIDs []string, pageSize int) (int, int, int, error)
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
	var totalProcessedProjectNum, totalProcessedClusterNum, totalDeletedClusterCount int
	for _, projectID := range projectIDs {
		// 0. Record the start time of synchronization for this project in UTC.
		syncStartTime := time.Now().UTC()
		var processedClusterNum int
		page := 1

		for {
			// 1. Fetch cluster info using projectID
			clusters, total, err := c.fetcher.FetchClusters(ctx, projectID, page, pageSize)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("failed to fetch clusters for project %s with pageSize %d: %w", projectID, pageSize, err)
			}

			// 2. Upsert cluster info. Existing clusters will have their updated_at timestamp refreshed.
			if err := c.store.StoreClusters(ctx, clusters); err != nil {
				return 0, 0, 0, fmt.Errorf("failed to store %d clusters for project %s with pageSize %d: %w", len(clusters), projectID, pageSize, err)
			}

			// 3. If the number of processed cluster is less than total, repeat the process.
			processedClusterNum += len(clusters)
			if processedClusterNum >= total {
				break
			}

			// 4. Increment the page number to fetch the next set of clusters
			page++
		}

		// Mark clusters as deleted if they were not updated during this sync cycle.
		// We identify them by checking if their `updated_at` is older than `syncStartTime`.
		deletedCount, err := c.store.MarkStaleClustersAsDeleted(ctx, projectID, syncStartTime)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to mark stale clusters as deleted for project %s: %w", projectID, err)
		}

		totalProcessedProjectNum++
		totalProcessedClusterNum += processedClusterNum
		totalDeletedClusterCount += int(deletedCount)
	}

	return totalProcessedProjectNum, totalProcessedClusterNum, totalDeletedClusterCount, nil
}
