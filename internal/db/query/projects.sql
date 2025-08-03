-- name: UpsertProject :exec
INSERT INTO
    projects (
        id,
        org_id,
        name,
        cluster_count,
        user_count,
        create_timestamp,
        aws_cmek_enabled
    )
VALUES
    (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    org_id = VALUES(org_id),
    name = VALUES(name),
    cluster_count = VALUES(cluster_count),
    user_count = VALUES(user_count),
    create_timestamp = VALUES(create_timestamp),
    aws_cmek_enabled = VALUES(aws_cmek_enabled);

-- name: ListProjectIDsWithClusters :many
SELECT
    id
FROM
    projects
WHERE
    cluster_count > 0
    AND DATE(fetched_at) = (
        SELECT
            DATE(MAX(fetched_at))
        FROM
            projects
    );
