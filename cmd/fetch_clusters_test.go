// cmd/fetch_clusters_test.go
package cmd

import (
	"testing"
	"time"
)

func TestValidateFetchClustersArgs_SuccessWithProjectIDs(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:          "k",
		APISecret:       "s",
		APIEndpointBase: "https://api.tidbcloud.com/api/v1beta",
		ProjectIDs:      []string{"1", "2"},
		PageSize:        50,
		All:             false,
		DBHost:          "127.0.0.1",
		DBUser:          "root",
		DBName:          "test",
		DBPort:          4000,
		DBPassword:      "",
		HTTPTimeout:     30 * time.Second,
		JobTimeout:      3 * time.Minute,
	}
	if err := validateFetchClustersArgs(arg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFetchClustersArgs_EmptyAPIKey(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:      "",
		APISecret:   "s",
		PageSize:    10,
		ProjectIDs:  []string{"1"},
		DBPort:      4000,
		HTTPTimeout: time.Second,
		JobTimeout:  time.Second,
	}
	if err := validateFetchClustersArgs(arg); err == nil {
		t.Fatalf("expected error for empty api key, got nil")
	}
}

func TestValidateFetchClustersArgs_EmptyAPISecret(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:      "k",
		APISecret:   "",
		PageSize:    10,
		ProjectIDs:  []string{"1"},
		DBPort:      4000,
		HTTPTimeout: time.Second,
		JobTimeout:  time.Second,
	}
	if err := validateFetchClustersArgs(arg); err == nil {
		t.Fatalf("expected error for empty api secret, got nil")
	}
}

func TestValidateFetchClustersArgs_DefaultAPIEndpointBaseWhenEmpty(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:          "k",
		APISecret:       "s",
		APIEndpointBase: "",
		PageSize:        10,
		ProjectIDs:      []string{"1"},
		DBPort:          4000,
		HTTPTimeout:     time.Second,
		JobTimeout:      time.Second,
	}
	if err := validateFetchClustersArgs(arg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := arg.APIEndpointBase, "https://api.tidbcloud.com/api/v1beta"; got != want {
		t.Fatalf("APIEndpointBase not defaulted, got %q want %q", got, want)
	}
}

func TestValidateFetchClustersArgs_PageSizeBounds(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		wantErr  bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"tooLarge", 101, true},
		{"minOK", 1, false},
		{"maxOK", 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := &fetchClustersArgs{
				APIKey:      "k",
				APISecret:   "s",
				PageSize:    tt.pageSize,
				ProjectIDs:  []string{"1"},
				DBPort:      4000,
				HTTPTimeout: time.Second,
				JobTimeout:  time.Second,
			}
			err := validateFetchClustersArgs(arg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFetchClustersArgs_AllAndProjectIDsConflict(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:      "k",
		APISecret:   "s",
		PageSize:    10,
		ProjectIDs:  []string{"1"},
		All:         true,
		DBPort:      4000,
		HTTPTimeout: time.Second,
		JobTimeout:  time.Second,
	}
	if err := validateFetchClustersArgs(arg); err == nil {
		t.Fatalf("expected error when --all is used with --project-id")
	}
}

func TestValidateFetchClustersArgs_ProjectIDsRequiredIfNotAll(t *testing.T) {
	arg := &fetchClustersArgs{
		APIKey:      "k",
		APISecret:   "s",
		PageSize:    10,
		All:         false,
		DBPort:      4000,
		HTTPTimeout: time.Second,
		JobTimeout:  time.Second,
	}
	if err := validateFetchClustersArgs(arg); err == nil {
		t.Fatalf("expected error when --all is false and no project IDs provided")
	}
}

func TestValidateFetchClustersArgs_DBPortBounds(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"zero", 0, true},
		{"neg", -1, true},
		{"tooLarge", 70000, true},
		{"minOK", 1, false},
		{"maxOK", 65535, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := &fetchClustersArgs{
				APIKey:      "k",
				APISecret:   "s",
				PageSize:    10,
				ProjectIDs:  []string{"p"},
				DBPort:      tt.port,
				HTTPTimeout: time.Second,
				JobTimeout:  time.Second,
			}
			err := validateFetchClustersArgs(arg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFetchClustersArgs_TimeoutsMustBePositive(t *testing.T) {
	tests := []struct {
		name        string
		httpTO      time.Duration
		jobTO       time.Duration
		expectError bool
	}{
		{"bothOK", time.Second, time.Second, false},
		{"httpZero", 0, time.Second, true},
		{"jobZero", time.Second, 0, true},
		{"httpNeg", -time.Second, time.Second, true},
		{"jobNeg", time.Second, -time.Second, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := &fetchClustersArgs{
				APIKey:      "k",
				APISecret:   "s",
				PageSize:    10,
				ProjectIDs:  []string{"p"},
				DBPort:      4000,
				HTTPTimeout: tt.httpTO,
				JobTimeout:  tt.jobTO,
			}
			err := validateFetchClustersArgs(arg)
			if (err != nil) != tt.expectError {
				t.Fatalf("err=%v expectError=%v", err, tt.expectError)
			}
		})
	}
}
