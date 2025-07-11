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
