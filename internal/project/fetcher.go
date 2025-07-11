package project

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/sgykfjsm/msk/internal/db"
)

// Project represents a TiDB Cloud project with relevant metadata.
type Project struct {
	ID              string `json:"id,omitempty"`
	OrgID           string `json:"org_id,omitempty"`
	Name            string `json:"name,omitempty"`
	ClusterCount    int    `json:"cluster_count,omitempty"`
	UserCount       int    `json:"user_count,omitempty"`
	CreateTimestamp string `json:"create_timestamp,omitempty"`
	AwsCmekEnabled  bool   `json:"aws_cmek_enabled,omitempty"`
}

// Projects is a slice of Project, used for API responses.
type Projects []Project

// ListProjectResponse represents the res structure for listing projects.
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

const defaultAPIEndpoint = "https://api.tidbcloud.com/v1beta/projects"

// In this module, I will implement the following:
// 1. Fetch projects from TiDB Cloud API
//	- Endpoint: https://api.tidbcloud.com/v1beta/projects
//	- Method: GET
//	- Response: ListProjectResponse (defined above)
//	- Ref: https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Project/operation/ListProjects
//	- Note: This endpoint requires an API Key for authentication, which should be passed from the caller.
// 	1.1 Initialize HTTP client with the API key passed from the caller
// 	1.2 Make a GET request to the API endpoint
// 	1.3 Parse the res into ListProjectResponse struct
// 	1.4 Handle errors appropriately (e.g., log and return error)
// 	1.5 Return the parsed Projects slice to the caller
// 2. Map the API res to internal Project struct
// 3. Store the projects in a MySQL-compatible database (TiDB)
// That's all!

// ProjectFetcher defines the interface for fetching projects from TiDB Cloud.
// It abstracts the logic of retrieving project data, allowing for easy testing and mocking.
type ProjectFetcher interface {
	// FetchProjects retrieves a list of projects from TiDB Cloud.
	// It takes pagination parameters (page and pageSize) to control the number of projects returned
	// It returns a slice of Project and an error if any.
	FetchProjects(ctx context.Context, page int, pageSize int) (Projects, error)
}

// APIProjectFetcher implements the ProjectFetcher interface.
// It uses an HTTP client to make requests to the TiDB Cloud API to fetch project data
type APIProjectFetcher struct {
	APIKey   string
	Endpoint string
	Client   *http.Client
}

// NewAPIProjectFetcher creates a new instance of APIProjectFetcher.
func NewAPIProjectFetcher(apiKey string, endpoint string, client *http.Client) *APIProjectFetcher {
	if endpoint == "" {
		endpoint = defaultAPIEndpoint // Use default endpoint if not provided
	}

	return &APIProjectFetcher{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Client:   client,
	}
}

// FetchProjects retrieves the list of projects from TiDB Cloud using the API key and endpoint.
// It returns a slice of Project and the total number of projects, or an error if the request fails or the response cannot be parsed.
// The second return value represents the total number of accessible projects.
// If the value is greater than the number of returned projects, it indicates there are more projects available to fetch.
// You can use the page and pageSize parameters to fetch the next page of projects.
func (f *APIProjectFetcher) FetchProjects(ctx context.Context, page int, pageSize int) (Projects, int, error) {
	// Set the API key in the request header to the HTTP Client
	req, err := http.NewRequest("GET", f.Endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+f.APIKey)
	q := req.URL.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", pageSize))
	req.URL.RawQuery = q.Encode()

	// Request to the TiDB Cloud API
	res, err := f.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusUnauthorized {
			return nil, 0, fmt.Errorf("unauthorized: check your API key")
		}

		// Decode the error response
		var apiError ListProjectResponseError
		if err := json.NewDecoder(res.Body).Decode(&apiError); err != nil {
			return nil, 0, fmt.Errorf("failed to decode error response from TiDB Cloud API: %w", err)
		}
		return nil, 0, fmt.Errorf("error from TiDB Cloud API: %s (code: %d, details: %v)", apiError.Message, apiError.Code, apiError.Details)
	}

	// Parse the res body into ListProjectResponse
	var listProjectResponse ListProjectResponse
	if err := json.NewDecoder(res.Body).Decode(&listProjectResponse); err != nil {
		return nil, 0, fmt.Errorf("failed to decode res from TiDB Cloud API: %w", err)
	}

	return listProjectResponse.Items, listProjectResponse.Total, nil
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

func NewDBProjectStore(config mysql.Config) (*DBProjectStore, error) {
	conn, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings if needed
	conn.SetMaxOpenConns(10)    // TODO: Set the value from config or environment variable
	conn.SetMaxIdleConns(5)     // TODO: Set the value from config or environment variable
	conn.SetConnMaxLifetime(60) // TODO: Set the value from config or environment variable

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
			// Stop the operation immediately if any error occurs
			tx.Rollback() // Rollback the transaction on error
			return fmt.Errorf("failed to upsert project %s: %w", project.ID, err)
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
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}
	return nil
}

