package clusters

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sgykfjsm/msk/internal/db"
	"github.com/stretchr/testify/require"
)

func TestAPIClusterFetcher_FetchClusters_Success(t *testing.T) {
	ctx := context.Background()

	expectedClusters := Clusters{
		{
			ID:              "1",
			ProjectID:       "1",
			Name:            "Cluster1",
			ClusterType:     "DEDICATED",
			CloudProvider:   "AWS",
			Region:          "us-west-2",
			CreateTimestamp: "1656991448",
			Status: ClusterStatus{
				TidbVersion:   "v1.2.3",
				ClusterStatus: "AVAILABLE",
				NodeMap: NodeMap{
					Tidb: Nodes{
						{NodeName: "tidb-0", AvailabilityZone: "us-west-2a", NodeSize: "8C16G"},
					},
					Tikv: Nodes{
						{NodeName: "tidb-0", AvailabilityZone: "us-west-2a", NodeSize: "8C16G", RAMBytes: "1234567890", StorageSizeGib: 100},
					},
				},
			},
		},
	}

	expectedResponse := &ListClustersResponse{
		Items: expectedClusters,
		Total: len(expectedClusters),
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIClusterFetcher(server.Client(), server.URL)
	clusters, total, err := fetcher.FetchClusters(ctx, "p1", 1, 10)

	require.NoError(t, err)
	require.Equal(t, expectedClusters, clusters)
	require.Equal(t, len(expectedClusters), total)
}

func TestAPIClusterFetcher_FetchClusters_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIClusterFetcher(server.Client(), server.URL)
	_, _, err := fetcher.FetchClusters(context.Background(), "p1", 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestAPIClusterFetcher_FetchClusters_HTTPError(t *testing.T) {
	httpErrorStatus := http.StatusInternalServerError
	fakeResponse := &ListClusterResponseError{
		Code:    httpErrorStatus,
		Message: http.StatusText(httpErrorStatus),
		Details: []string{"internal issue"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErrorStatus)
		_ = json.NewEncoder(w).Encode(fakeResponse)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIClusterFetcher(server.Client(), server.URL)
	_, _, err := fetcher.FetchClusters(context.Background(), "p1", 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "error from TiDB Cloud API")
	require.Contains(t, err.Error(), fmt.Sprintf("code: %d", httpErrorStatus))
	require.Contains(t, err.Error(), "internal issue")
}

func TestAPIClusterFetcher_FetchClusters_DecodeError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIClusterFetcher(server.Client(), server.URL)
	_, _, err := fetcher.FetchClusters(context.Background(), "p1", 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response from TiDB Cloud API")
}

func TestAPIClusterFetcher_FetchClusters_DecodeErrorResponseError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid json"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIClusterFetcher(server.Client(), server.URL)
	_, _, err := fetcher.FetchClusters(context.Background(), "p1", 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode error response from TiDB Cloud API")
}

func TestDBClusterStore_storageSizeGibForNode(t *testing.T) {
	testCases := []struct {
		name          string
		componentType db.ClusterNodesComponentType
		node          Node
		expected      sql.NullInt32
	}{
		{
			name:          "Tidb node should have null storage",
			componentType: db.ClusterNodesComponentTypeTidb,
			node:          Node{StorageSizeGib: 100}, // This value should be ignored
			expected:      sql.NullInt32{Valid: false},
		},
		{
			name:          "Tikv node should have valid storage",
			componentType: db.ClusterNodesComponentTypeTikv,
			node:          Node{StorageSizeGib: 1024},
			expected:      sql.NullInt32{Int32: 1024, Valid: true},
		},
		{
			name:          "Tiflash node should have valid storage",
			componentType: db.ClusterNodesComponentTypeTiflash,
			node:          Node{StorageSizeGib: 2048},
			expected:      sql.NullInt32{Int32: 2048, Valid: true},
		},
		{
			name:          "Unknown component type should have null storage",
			componentType: "some-unknown-type",
			node:          Node{StorageSizeGib: 500},
			expected:      sql.NullInt32{Valid: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := storageSizeGibForNode(tc.componentType, tc.node)
			require.Equal(t, tc.expected, actual)
		})
	}
}
