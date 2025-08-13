package clusters

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterService_FetchAndStoreClusters_Success(t *testing.T) {
	ctx := context.Background()

	// Create a mock mockFetcher and store
	mockFetcher := NewMockClusterFetcher(t)
	mockStore := NewMockClusterStore(t)
	svc := NewClusterService(mockFetcher, mockStore)

	projectID := "1"
	projectIDs := []string{projectID}
	mockClusters := Clusters([]Cluster{{ID: "10", ProjectID: projectID}})
	pageSize := 1

	// Test scenario: Target project has only a single cluster in itself.
	page := 1
	mockFetcher.EXPECT().
		FetchClusters(ctx, projectID, page, pageSize).
		Return(mockClusters, 1, nil).
		Times(1)

	mockStore.EXPECT().
		StoreClusters(ctx, mockClusters).
		Return(nil).
		Times(1)

	actualProjectNum, actualClusterNum, err := svc.FetchAndStoreClusters(ctx, projectIDs, pageSize)
	require.NoError(t, err)
	require.Equal(t, len(projectIDs), actualProjectNum)
	require.Equal(t, len(mockClusters), actualClusterNum)
}

func TestClusterService_FetchAndStoreClusters_Paging_Success(t *testing.T) {
	ctx := context.Background()

	mockFetcher := NewMockClusterFetcher(t)
	mockStore := NewMockClusterStore(t)
	svc := NewClusterService(mockFetcher, mockStore)

	projectID := "1"
	projectIDs := []string{projectID}
	mockClusters1 := Clusters([]Cluster{{ID: "10", ProjectID: projectID}})
	mockClusters2 := Clusters([]Cluster{{ID: "11", ProjectID: projectID}})
	pageSize := 1

	// Test scenario: TiDB Cloud API returns multiple pages of clusters
	// so we need to fetch store them all using pagination

	// Page1: Fetcher should return 1 cluster, total 2
	page := 1
	mockFetcher.EXPECT().
		FetchClusters(ctx, projectID, page, pageSize).
		Return(mockClusters1, 2, nil).
		Times(1)

	mockStore.EXPECT().
		StoreClusters(ctx, mockClusters1).
		Return(nil).
		Times(1)

	// Page2: Fetcher should return 2 clusters, total 2
	page = 2
	mockFetcher.EXPECT().
		FetchClusters(ctx, projectID, page, pageSize).
		Return(mockClusters2, 2, nil).
		Times(1)

	mockStore.EXPECT().
		StoreClusters(ctx, mockClusters2).
		Return(nil).
		Times(1)

	actualProjectNum, actualClusterNum, err := svc.FetchAndStoreClusters(ctx, projectIDs, pageSize)
	require.NoError(t, err)
	require.Equal(t, len(projectIDs), actualProjectNum)
	require.Equal(t, len(mockClusters1)+len(mockClusters2), actualClusterNum)
}

func TestClusterService_FetchAndStoreClusters_FetchError(t *testing.T) {
	ctx := context.Background()

	mockFetcher := NewMockClusterFetcher(t)
	mockStore := NewMockClusterStore(t)
	svc := NewClusterService(mockFetcher, mockStore)

	projectID := "1"
	projectIDs := []string{projectID}
	pageSize := 1

	// Test scenario: Failed to fetch clusters
	page := 1
	mockFetcher.EXPECT().
		FetchClusters(ctx, projectID, page, pageSize).
		Return(nil, 0, errors.New("failed to fetch clusters")).
		Times(1)

	actualProjectNum, actualClusterNum, err := svc.FetchAndStoreClusters(ctx, projectIDs, pageSize)
	require.Error(t, err)
	require.Equal(t, 0, actualProjectNum)
	require.Equal(t, 0, actualClusterNum)
}

func TestClusterService_FetchAndStoreClusters_StoreError(t *testing.T) {
	ctx := context.Background()

	mockFetcher := NewMockClusterFetcher(t)
	mockStore := NewMockClusterStore(t)
	svc := NewClusterService(mockFetcher, mockStore)

	projectID := "1"
	projectIDs := []string{projectID}
	mockClusters1 := Clusters([]Cluster{{ID: "10", ProjectID: projectID}})
	pageSize := 1

	// Test scenario: Failed to store clusters
	page := 1
	mockFetcher.EXPECT().
		FetchClusters(ctx, projectID, page, pageSize).
		Return(mockClusters1, 1, nil).
		Times(1)

	mockStore.EXPECT().
		StoreClusters(ctx, mockClusters1).
		Return(errors.New("failed to store clusters")).
		Times(1)

	actualProjectNum, actualClusterNum, err := svc.FetchAndStoreClusters(ctx, projectIDs, pageSize)
	require.Error(t, err)
	require.Equal(t, 0, actualProjectNum)
	require.Equal(t, 0, actualClusterNum)
}
