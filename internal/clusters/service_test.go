package clusters

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClusterService_FetchAndStoreClusters(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	pageSize := 2

	clustersPage1 := Clusters{
		{ID: "cluster-1", ProjectID: projectID},
		{ID: "cluster-2", ProjectID: projectID},
	}
	clustersPage2 := Clusters{
		{ID: "cluster-3", ProjectID: projectID},
	}

	tests := []struct {
		name                      string
		projectIDs                []string
		setupMocks                func(fetcher *MockClusterFetcher, store *MockClusterStore)
		expectedProcessedProjects int
		expectedProcessedClusters int
		expectedDeletedClusters   int
		expectErr                 bool
		expectedErrMsg            string
	}{
		{
			name:       "Success with pagination",
			projectIDs: []string{projectID},
			setupMocks: func(fetcher *MockClusterFetcher, store *MockClusterStore) {
				// Page 1
				fetcher.EXPECT().FetchClusters(ctx, projectID, 1, pageSize).Return(clustersPage1, 3, nil).Once()
				store.EXPECT().StoreClusters(ctx, clustersPage1).Return(nil).Once()
				// Page 2
				fetcher.EXPECT().FetchClusters(ctx, projectID, 2, pageSize).Return(clustersPage2, 3, nil).Once()
				store.EXPECT().StoreClusters(ctx, clustersPage2).Return(nil).Once()
				store.EXPECT().MarkStaleClustersAsDeleted(ctx, projectID, mock.AnythingOfType("time.Time")).Return(int64(1), nil).Once()
			},
			expectedProcessedProjects: 1,
			expectedProcessedClusters: 3,
			expectedDeletedClusters:   1,
			expectErr:                 false,
		},
		{
			name:       "Success with single page",
			projectIDs: []string{projectID},
			setupMocks: func(fetcher *MockClusterFetcher, store *MockClusterStore) {
				fetcher.EXPECT().FetchClusters(ctx, projectID, 1, pageSize).Return(clustersPage1, 2, nil).Once()
				store.EXPECT().StoreClusters(ctx, clustersPage1).Return(nil).Once()
				store.EXPECT().MarkStaleClustersAsDeleted(ctx, projectID, mock.AnythingOfType("time.Time")).Return(int64(0), nil).Once()
			},
			expectedProcessedProjects: 1,
			expectedProcessedClusters: 2,
			expectedDeletedClusters:   0,
			expectErr:                 false,
		},
		{
			name:       "Fetcher fails on first call",
			projectIDs: []string{projectID},
			setupMocks: func(fetcher *MockClusterFetcher, store *MockClusterStore) {
				fetcher.EXPECT().FetchClusters(ctx, projectID, 1, pageSize).Return(nil, 0, errors.New("API error")).Once()
			},
			expectedProcessedProjects: 0,
			expectedProcessedClusters: 0,
			expectedDeletedClusters:   0,
			expectErr:                 true,
			expectedErrMsg:            "failed to fetch clusters for project project-1",
		},
		{
			name:       "Store fails",
			projectIDs: []string{projectID},
			setupMocks: func(fetcher *MockClusterFetcher, store *MockClusterStore) {
				fetcher.EXPECT().FetchClusters(ctx, projectID, 1, pageSize).Return(clustersPage1, 2, nil).Once()
				store.EXPECT().StoreClusters(ctx, clustersPage1).Return(errors.New("DB error")).Once()
			},
			expectedProcessedProjects: 0,
			expectedProcessedClusters: 0,
			expectedDeletedClusters:   0,
			expectErr:                 true,
			expectedErrMsg:            "failed to store 2 clusters for project project-1",
		},
		{
			name:       "MarkStaleClustersAsDeleted fails",
			projectIDs: []string{projectID},
			setupMocks: func(fetcher *MockClusterFetcher, store *MockClusterStore) {
				fetcher.EXPECT().FetchClusters(ctx, projectID, 1, pageSize).Return(clustersPage1, 2, nil).Once()
				store.EXPECT().StoreClusters(ctx, clustersPage1).Return(nil).Once()
				store.EXPECT().MarkStaleClustersAsDeleted(ctx, projectID, mock.AnythingOfType("time.Time")).Return(int64(0), errors.New("mark stale error")).Once()
			},
			expectedProcessedProjects: 0,
			expectedProcessedClusters: 0,
			expectedDeletedClusters:   0,
			expectErr:                 true,
			expectedErrMsg:            "failed to mark stale clusters as deleted for project project-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockFetcher := NewMockClusterFetcher(t)
			mockStore := NewMockClusterStore(t)
			tt.setupMocks(mockFetcher, mockStore)

			service := NewClusterService(mockFetcher, mockStore)

			// Act
			processedProjects, processedClusters, deletedClusters, err := service.FetchAndStoreClusters(ctx, tt.projectIDs, pageSize)

			// Assert
			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.expectedProcessedProjects, processedProjects)
			require.Equal(t, tt.expectedProcessedClusters, processedClusters)
			require.Equal(t, tt.expectedDeletedClusters, deletedClusters)
		})
	}
}
