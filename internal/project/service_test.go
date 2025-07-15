package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectService_FetchAndStoreProjects_Success(t *testing.T) {
	ctx := context.Background()

	// Create a mock fetcher
	mockFetcher := NewMockProjectFetcher(t)
	mockStore := NewMockProjectStore(t)
	svc := NewProjectService(mockFetcher, mockStore)

	// Define test data
	projectPage1 := Projects{
		{ID: "1", OrgID: "org1", Name: "Project1", ClusterCount: 2, UserCount: 5, CreateTimestamp: "1622547800", AwsCmekEnabled: true},
		{ID: "2", OrgID: "org2", Name: "Project2", ClusterCount: 3, UserCount: 10, CreateTimestamp: "1622547801", AwsCmekEnabled: false},
	}
	projectPage2 := Projects{
		{ID: "3", OrgID: "org3", Name: "Project3", ClusterCount: 1, UserCount: 2, CreateTimestamp: "1622547802", AwsCmekEnabled: true},
	}

	// Test scenario: TiDB Cloud API returns multiple pages of projects
	// so we need to fetch and store them all using pagination.

	// Page1: Fetcher should return 2 projects, total 3
	mockFetcher.EXPECT().
		FetchProjects(ctx, 1, 2).
		Return(projectPage1, 3, nil).
		Times(1)

	// Store the first page of projects
	mockStore.EXPECT().
		StoreProjects(ctx, projectPage1).
		Return(nil).
		Times(1)

	// Page2: Fetcher should return 1 project, total 3
	mockFetcher.EXPECT().
		FetchProjects(ctx, 2, 2).
		Return(projectPage2, 3, nil).
		Times(1)

	// Store the second page of projects
	mockStore.EXPECT().
		StoreProjects(ctx, projectPage2).
		Return(nil).
		Times(1)

	err := svc.FetchAndStoreProjects(ctx, 1, 2)
	require.NoError(t, err)
}
