-- name: UpsertCluster :exec
INSERT INTO clusters (
        id,
        project_id,
        name,
        cluster_type,
        cloud_provider,
        region,
        create_timestamp,
        tidb_version,
        cluster_status
    )
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY
UPDATE
    project_id = VALUES(project_id),
    name = VALUES(name),
    cluster_type = VALUES(cluster_type),
    cloud_provider = VALUES(cloud_provider),
    region = VALUES(region),
    create_timestamp = VALUES(create_timestamp),
    tidb_version = VALUES(tidb_version),
    cluster_status = VALUES(cluster_status);

-- name: MarkStaleClustersAsDeleted :execresult
-- MarkStaleClustersAsDeleted marks clusters as deleted if they have not been updated since the given timestamp.
UPDATE clusters
SET is_deleted = TRUE,
    deleted_at = CURRENT_TIMESTAMP
WHERE project_id = sqlc.arg('project_id')
    AND updated_at < sqlc.arg('synced_at')
    AND is_deleted = FALSE;
