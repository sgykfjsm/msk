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

-- name: MarkClustersAsDeleted :exec
UPDATE clusters
SET is_deleted = TRUE,
    deleted_at = CURRENT_TIMESTAMP
WHERE project_id = ?
    AND id NOT IN (sqlc.slice('id'))
    AND is_deleted = FALSE;
