package clusters

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/sgykfjsm/msk/internal/db"
)

// Cluster represents the response structure for listing clusters.
// Ref: https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Cluster/operation/ListClustersOfProject
// Note that this instance holds only required attributes for this module, and omit some properties.
type Cluster struct {
	ID              string        `json:"id,omitempty"`
	ProjectID       string        `json:"project_id,omitempty"`
	Name            string        `json:"name,omitempty"`
	ClusterType     string        `json:"cluster_type,omitempty"`
	CloudProvider   string        `json:"cloud_provider,omitempty"`
	Region          string        `json:"region,omitempty"`
	CreateTimestamp string        `json:"create_timestamp,omitempty"`
	Status          ClusterStatus `json:"status"`
}

// Clusters is a slice of Cluster.
type Clusters []Cluster

// ClusterStatus holds runtime status information of a cluster, including version and node mapping.
type ClusterStatus struct {
	TidbVersion   string  `json:"tidb_version,omitempty"`
	ClusterStatus string  `json:"cluster_status,omitempty"`
	NodeMap       NodeMap `json:"node_map"`
}

// NodeMap maps cluster component types (tidb, tikv, tiflash) to their respective node lists.
type NodeMap struct {
	Tidb    Nodes `json:"tidb,omitempty"`
	Tikv    Nodes `json:"tikv,omitempty"`
	Tiflash Nodes `json:"tiflash,omitempty"`
}

// Node represents a single node in a cluster, such as TiDB, TiKV, or TiFlash node.
type Node struct {
	NodeName         string `json:"node_name,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	NodeSize         string `json:"node_size,omitempty"`
	VcpuNum          int    `json:"vcpu_num,omitempty"`
	RAMBytes         string `json:"ram_bytes,omitempty"` // defined as a string in the TiDB Cloud API response schema.
	StorageSizeGib   int    `json:"storage_size_gib,omitempty"`
	Status           string `json:"status,omitempty"`
}

// Nodes is a slice of Node.
type Nodes []Node

// ListClustersResponse represents the successful response structure from the TiDB Cloud ListClustersOfProject API.
type ListClustersResponse struct {
	Items Clusters `json:"items,omitempty"`
	Total int      `json:"total,omitempty"`
}

// ListClusterResponseError represents the error response from the TiDB Cloud API.
type ListClusterResponseError struct {
	Message string   `json:"message,omitempty"`
	Code    int      `json:"code,omitempty"`
	Details []string `json:"details,omitempty"`
}

// ClusterFetcher defines an interface for fetching cluster metadata from a remote source.
type ClusterFetcher interface {
	FetchClusters(ctx context.Context, projectID string, page, pageSize int) (Clusters, int, error)
}

// APIClusterFetcher implements the ClusterFetcher interface using the TiDB Cloud API.
type APIClusterFetcher struct {
	Client       *http.Client
	EndpointBase string
}

const defaultAPIEndpointBase = "https://api.tidbcloud.com/api/v1beta"

// NewAPIClusterFetcher returns a new APIClusterFetcher with the given HTTP client and API endpoint base URL.
// This function allows for injection of custom clients for testing and tracing.
func NewAPIClusterFetcher(client *http.Client, endpointBase string) *APIClusterFetcher {
	return &APIClusterFetcher{
		Client:       client,
		EndpointBase: endpointBase,
	}
}

// FetchClusters retrieves the list of clusters for a specific project from the TiDB Cloud API.
// It returns a slice of Cluster, the total number of clusters, or an error if the request fails or the response cannot be parsed.
// The second return value is the total number of clusters in the project.
// If this value is greater than the number of returned clusters, it indicates more clusters can be fetched on subsequent pages.
// The page and pageSize parameters can be used for pagination.
func (f *APIClusterFetcher) FetchClusters(ctx context.Context, projectID string, page, pageSize int) (Clusters, int, error) {
	apiEndpoint, err := url.JoinPath(f.EndpointBase, "projects", projectID, "clusters")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to construct API endpoint URL started with %s wih projectID %s: %w", f.EndpointBase, projectID, err)
	}

	req, err := http.NewRequest("GET", apiEndpoint, nil)
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

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var listClustersResponse ListClustersResponse
		if err := json.NewDecoder(resp.Body).Decode(&listClustersResponse); err != nil {
			return nil, 0, fmt.Errorf("succeeded to request, but failed to decode response from TiDB Cloud API: %w", err)
		}

		return listClustersResponse.Items, listClustersResponse.Total, nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, 0, errors.New("unauthorized: check your API key")
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return nil, 0, errors.New("rate limit: retry after some minutes. See https://docs.pingcap.com/tidbcloud/api/v1beta/#section/Rate-Limiting")
	}

	var apiError ListClusterResponseError
	if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
		return nil, 0, fmt.Errorf("failed to decode error response from TiDB Cloud API: %w, status: %s", err, resp.Status)
	}

	return nil, 0, fmt.Errorf("error from TiDB Cloud API: %s (code: %d, details: %v, endpoint: %s), status: %s", apiError.Message, apiError.Code, apiError.Details, apiEndpoint, resp.Status)
}

// ClusterStore defines an interface for storing cluster metadata.
type ClusterStore interface {
	StoreClusters(ctx context.Context, clusters Clusters) error
	MarkStaleClustersAsDeleted(ctx context.Context, projectID string, syncedAt time.Time) (int64, error)
}

// DBClusterStore represents a database-backed implementation for persisting cluster metadata.
// It wraps a SQL database connection and provides access to pre-defined SQL queries.
type DBClusterStore struct {
	conn    *sql.DB
	Queries *db.Queries
}

// NewDBClusterStore initializes a new DBClusterStore using the given DSN and optional connection pool settings.
// It opens a connection to the database, verifies connectivity, and prepares SQLC-generated query methods.
// Returns an error if the connection fails or cannot be verified.
func NewDBClusterStore(dsn string, poolConfig *db.PoolConfig) (*DBClusterStore, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if poolConfig == nil {
		poolConfig = db.NewPoolConfig()
	}
	conn.SetMaxOpenConns(poolConfig.MaxOpenConns)
	conn.SetMaxIdleConns(poolConfig.MaxIdleConns)
	conn.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := conn.PingContext(timeoutCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DBClusterStore{
		Queries: db.New(conn),
		conn:    conn,
	}, nil
}

// StoreClusters inserts or updates the given list of clusters into the database within a transaction scope.
// If any operation fails, the transaction will be rolled back and the error returned.
// This method is a no-op if the input slice is empty.
func (s *DBClusterStore) StoreClusters(ctx context.Context, clusters Clusters) error {
	if len(clusters) == 0 {
		return nil // No clusters to store, nothing to do
	}

	// Map clusters to SQL upsert statements
	tx, err := s.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	qtx := s.Queries.WithTx(tx)
	for _, cluster := range clusters {
		createTimestamp, err := strconv.ParseInt(cluster.CreateTimestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid create_timestamp format for cluster %s(%s): %w", cluster.Name, cluster.ID, err)
		}
		values := db.UpsertClusterParams{
			ID:              cluster.ID,
			ProjectID:       cluster.ProjectID,
			Name:            cluster.Name,
			ClusterType:     cluster.ClusterType,
			CloudProvider:   cluster.CloudProvider,
			Region:          cluster.Region,
			CreateTimestamp: createTimestamp,
			ClusterStatus:   cluster.Status.ClusterStatus,
			TidbVersion:     cluster.Status.TidbVersion,
		}
		if err := qtx.UpsertCluster(ctx, values); err != nil {
			upsertErr := fmt.Errorf("failed to upsert cluster %s: %w", cluster.ID, err)
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

func (s *DBClusterStore) MarkStaleClustersAsDeleted(ctx context.Context, projectID string, syncedAt time.Time) (rowsAffected int64, err error) {
	tx, err := s.conn.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("failed to rollback transaction after error: %w", rbErr)
			}
		}
	}()

	qtx := s.Queries.WithTx(tx)
	values := db.MarkStaleClustersAsDeletedParams{
		ProjectID: projectID,
		SyncedAt:  syncedAt,
	}

	res, err := qtx.MarkStaleClustersAsDeleted(ctx, values)
	if err != nil {
		return 0, fmt.Errorf("failed to execute mark stale clusters as deleted for project %s: %w", projectID, err)
	}

	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected after marking stale clusters: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return rowsAffected, nil
}

// Close closes the underlying database connection held by the DBClusterStore.
// Returns an error if the connection could not be closed properly.
func (s *DBClusterStore) Close() error {
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			s.conn = nil // Set to nil to avoid double close
			return fmt.Errorf("failed to close database connection: %w", err)
		}
	}

	return nil
}
