-- This table is based on the response from the TiDB Cloud API "List all accessible projects."
-- https://docs.pingcap.com/tidbcloud/api/v1beta/#tag/Project/operation/ListProjects
CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    cluster_count INT,
    user_count INT,
    create_timestamp BIGINT,
    aws_cmek_enabled BOOLEAN,
    fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
