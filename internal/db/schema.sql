-- This table is based on the response from the TiDB Cloud API "List all accessible projects."
-- https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Project/operation/ListProjects
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    cluster_count INT NOT NULL DEFAULT 0,
    user_count INT NOT NULL DEFAULT 0,
    create_timestamp BIGINT NOT NULL, -- Use BIGINT to store timestamp in seconds for API compatibility
    aws_cmek_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    fetched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- This table is based on the response from the TiDB Cloud API "Get a cluster by ID."
-- Especially, focusing on the cluster meta data
-- https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Cluster/operation/GetCluster
CREATE TABLE IF NOT EXISTS clusters (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    cluster_type VARCHAR(32) NOT NULL,
    cloud_provider VARCHAR(32) NOT NULL,
    region VARCHAR(32) NOT NULL,
    create_timestamp BIGINT NOT NULL,
    tidb_version VARCHAR(32) NOT NULL,
    cluster_status VARCHAR(32) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE
);
