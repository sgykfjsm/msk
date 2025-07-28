package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/icholy/digest"
	"github.com/sgykfjsm/msk/internal/db"
)

// In this module, I will implement the following:
// 1. Fetch projects from TiDB Cloud API
//	- Endpoint: https://api.tidbcloud.com/v1beta/projects
//	- Method: GET
//	- Response: ListProjectResponse (defined above)
//	- Ref: https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Project/operation/ListProjects
//	- Note: This endpoint requires an API Key for authentication, which should be passed from the caller.
// 	1.1 Initialize HTTP client with the API key passed from the caller
// 	1.2 Make a GET request to the API endpoint
// 	1.3 Parse the resp into ListProjectResponse struct
// 	1.4 Handle errors appropriately (e.g., log and return error)
// 	1.5 Return the parsed Projects slice to the caller
// 2. Map the API resp to internal Project struct
// 3. Store the projects in a MySQL-compatible database (TiDB)
// That's all!

// Project represents a TiDB Cloud project with relevant metadata.
type Project struct {
	ID              string `json:"id,omitempty"`
	OrgID           string `json:"orgId,omitempty"`
	Name            string `json:"name,omitempty"`
	ClusterCount    int    `json:"clusterCount,omitempty"`
	UserCount       int    `json:"userCount,omitempty"`
	CreateTimestamp string `json:"createTimestamp,omitempty"`
	AwsCmekEnabled  bool   `json:"awsCmekEnabled,omitempty"`
}

type Projects []Project // Projects represents a list of projects returned by the TiDB Cloud API.

// ListProjectResponse represents the response structure for listing projects.
// Ref: https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Project/operation/ListProjects
type ListProjectResponse struct {
	Items Projects `json:"items,omitempty"`
	Total int      `json:"total,omitempty"`
}

// ListProjectResponseError represents the error structure for the ListProjects API response.
type ListProjectResponseError struct {
	Message string   `json:"message,omitempty"`
	Code    int      `json:"code,omitempty"`
	Details []string `json:"details,omitempty"`
}

// ProjectFetcher defines the interface for fetching projects from TiDB Cloud.
// It abstracts the logic of retrieving project data, allowing for easy testing and mocking.
type ProjectFetcher interface {
	// FetchProjects retrieves a list of projects from TiDB Cloud.
	// It takes pagination parameters (page and pageSize) to control the number of projects returned
	// It returns a slice of Project and an error if any.
	FetchProjects(ctx context.Context, page int, pageSize int) (Projects, int, error)
}

// APIProjectFetcher implements the ProjectFetcher interface.
// Fetcher is expected to use HTTP client generated from the library github.com/icholy/digest
type APIProjectFetcher struct {
	APIKey    string // Public Key
	APISecret string // Private Key
	Endpoint  string
}

const defaultAPIEndpoint = "https://api.tidbcloud.com/v1beta/projects"

// NewAPIProjectFetcher creates a new instance of APIProjectFetcher.
func NewAPIProjectFetcher(apiKey, apiSecret string, endpoint string) *APIProjectFetcher {
	if endpoint == "" {
		endpoint = defaultAPIEndpoint // Use default endpoint if not provided
	}

	return &APIProjectFetcher{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Endpoint:  endpoint,
	}
}

// FetchProjects retrieves the list of projects from TiDB Cloud using the API key and endpoint.
// It returns a slice of Project and the total number of projects, or an error if the request fails or the response cannot be parsed.
// The second return value represents the total number of accessible projects.
// If the value is greater than the number of returned projects, it indicates there are more projects available to fetch.
// You can use the page and pageSize parameters to fetch the next page of projects.
func (f *APIProjectFetcher) FetchProjects(ctx context.Context, page int, pageSize int) (Projects, int, error) {
	transport := &digest.Transport{
		Username: f.APIKey,
		Password: f.APISecret,
	}
	client := &http.Client{Transport: transport}

	// Set the API key in the request header to the HTTP Client
	req, err := http.NewRequest("GET", f.Endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	req = req.WithContext(ctx)

	q := req.URL.Query()
	if page <= 0 {
		page = 1 // Default to page 1 if not provided or invalid
	}

	if pageSize <= 0 {
		pageSize = 20 // Default to 20 if not provided or invalid
	}

	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	req.URL.RawQuery = q.Encode()

	// Request to the TiDB Cloud API
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var listProjectResponse ListProjectResponse
		if err := json.NewDecoder(resp.Body).Decode(&listProjectResponse); err != nil {
			return nil, 0, fmt.Errorf("succeeded to request, but failed to decode response from TiDB Cloud API: %w", err)
		}

		return listProjectResponse.Items, listProjectResponse.Total, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, errors.New("unauthorized: check your API key")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return nil, 0, errors.New("rate limit: retry after some minutes. See https://docs.pingcap.com/tidbcloud/api/v1beta/#section/Rate-Limiting")
	}

	// Decode the error response
	var apiError ListProjectResponseError
	if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
		return nil, 0, fmt.Errorf("failed to decode error response from TiDB Cloud API: %w, status: %s", err, resp.Status)
	}

	return nil, 0, fmt.Errorf("error from TiDB Cloud API: %s (code: %d, details: %v, endpoint: %s), status: %s", apiError.Message, apiError.Code, apiError.Details, f.Endpoint, resp.Status)

}

// ProjectStore defines the interface for storing projects in a database.
type ProjectStore interface {
	// StoreProjects saves the provided projects to the database.
	// It returns an error if the operation fails.
	StoreProjects(ctx context.Context, projects Projects) error
}

// DBProjectStore implements the ProjectStore interface.
// It uses a SQL database connection to store project data.
// It assumes the database is compatible with MySQL (e.g., TiDB).
type DBProjectStore struct {
	conn    *sql.DB
	Queries *db.Queries // Assuming db.Queries is generated by sqlc for database operations
}

// NewDBProjectStore creates a new instance of DBProjectStore.
// It takes a DSN (Data Source Name) for the database connection and an optional pool configuration.
func NewDBProjectStore(dsn string, poolConfig *db.PoolConfig) (*DBProjectStore, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if poolConfig == nil {
		poolConfig = db.NewPoolConfig() // Use default pool config if none provided
	}

	conn.SetMaxOpenConns(poolConfig.MaxOpenConns)
	conn.SetMaxIdleConns(poolConfig.MaxIdleConns)
	conn.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := conn.PingContext(timeoutCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBProjectStore{
		Queries: db.New(conn),
		conn:    conn,
	}, nil
}

func (s *DBProjectStore) StoreProjects(ctx context.Context, projects Projects) error {
	if len(projects) == 0 {
		return nil // No projects to store, nothing to do
	}

	// Map projects to SQL upsert statements
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	qtx := s.Queries.WithTx(tx)
	for _, project := range projects {
		ct, err := strconv.ParseInt(project.CreateTimestamp, 10, 64)
		if err != nil {
			ct = 0 // Default to 0 if parsing fails, assuming CreateTimestamp is not critical
		}
		values := db.UpsertProjectParams{
			ID:              project.ID,
			OrgID:           project.OrgID,
			Name:            project.Name,
			ClusterCount:    int32(project.ClusterCount),
			UserCount:       int32(project.UserCount),
			CreateTimestamp: ct,
			AwsCmekEnabled:  project.AwsCmekEnabled,
		}

		if err := qtx.UpsertProject(ctx, values); err != nil {
			upsertErr := fmt.Errorf("failed to upsert project %s: %w", project.ID, err)
			if err := tx.Rollback(); err != nil {
				return errors.Join(upsertErr, fmt.Errorf("failed to rollback transaction: %w", err))
			}

			return upsertErr
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (s *DBProjectStore) Close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			s.conn = nil // Set to nil to avoid double close
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}
