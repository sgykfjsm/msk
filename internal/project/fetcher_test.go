package project

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIProjectFetcher_FetchProjects_Success(t *testing.T) {
	ctx := context.Background()

	expectedProjects := Projects{
		{ID: "1", OrgID: "org1", Name: "Project1", ClusterCount: 2, UserCount: 5, CreateTimestamp: "1622547800", AwsCmekEnabled: true},
		{ID: "2", OrgID: "org2", Name: "Project2", ClusterCount: 3, UserCount: 10, CreateTimestamp: "1622547801", AwsCmekEnabled: false},
	}

	expectedResponse := &ListProjectResponse{
		Items: expectedProjects,
		Total: 2,
	}

	// Fake handler to simulate API response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIProjectFetcher("dummy-token", server.URL, server.Client())
	projects, total, err := fetcher.FetchProjects(ctx, 1, 10)

	require.NoError(t, err)
	require.Equal(t, expectedProjects, projects)
	require.Equal(t, len(expectedProjects), total)
}

func TestAPIProjectFetcher_FetchProjects_Unauthorized(t *testing.T) {
	// Fake handler to simulate unauthorized access
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIProjectFetcher("invalid-token", server.URL, server.Client())
	_, _, err := fetcher.FetchProjects(context.Background(), 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unauthorized")
}

func TestAPIProjectFetcher_FetchProjects_HTTPError(t *testing.T) {
	httpErrorStatus := http.StatusInternalServerError
	fakeResponse := &ListProjectResponseError{
		Code:    httpErrorStatus,
		Message: http.StatusText(httpErrorStatus),
		Details: []string{"An unexpected error occurred."},
	}

	// Fake handler to simulate an internal server error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpErrorStatus)
		_ = json.NewEncoder(w).Encode(fakeResponse)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIProjectFetcher("dummy-token", server.URL, server.Client())
	_, _, err := fetcher.FetchProjects(context.Background(), 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "error from TiDB Cloud API")
	require.Contains(t, err.Error(), fmt.Sprintf("code: %d", httpErrorStatus))
	require.Contains(t, err.Error(), "details: [An unexpected error occurred.]")
}

func TestAPIProjectFetcher_FetchProjects_DecodeError(t *testing.T) {
	// Fake handler to simulate a decoding error
	fakeResponse := []byte("invalid json")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fakeResponse)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIProjectFetcher("dummy-token", server.URL, server.Client())
	_, _, err := fetcher.FetchProjects(context.Background(), 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response from TiDB Cloud API")
}

func TestAPIProjectFetcher_FetchProjects_DecodeErrorResponseError(t *testing.T) {
	// Fake handler to simulate a decoding error for the error response
	fakeResponse := []byte("invalid json")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(fakeResponse)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	fetcher := NewAPIProjectFetcher("dummy-token", server.URL, server.Client())
	_, _, err := fetcher.FetchProjects(context.Background(), 1, 10)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode error response from TiDB Cloud API")
}
