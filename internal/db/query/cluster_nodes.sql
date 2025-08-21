-- name: UpsertClusterNode :exec
INSERT INTO
    cluster_nodes (
        cluster_id,
        node_name,
        component_type,
        availability_zone,
        node_size,
        storage_size_gib,
        status
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    component_type = VALUES(component_type),
    availability_zone = VALUES(availability_zone),
    node_size = VALUES(node_size),
    storage_size_gib = VALUES(storage_size_gib),
    status = VALUES(status);

-- name: MarkClusterNodesAsDeleted :exec
UPDATE cluster_nodes
SET
    is_deleted = TRUE,
    deleted_at = CURRENT_TIMESTAMP
where

    cluster_id = ?
    AND node_name NOT IN(sqlc.slice('node_name'))
    AND is_deleted = FALSE;
