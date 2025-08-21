package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/icholy/digest"
	"github.com/sgykfjsm/msk/internal/clusters"
	"github.com/sgykfjsm/msk/internal/project"
	"github.com/sgykfjsm/msk/internal/util"
	"github.com/urfave/cli/v3"
)

var FetchClustersCmd = &cli.Command{
	Name:  "fetch-clusters",
	Usage: "Fetch and store clusters from the TiDB Cloud API",
	UsageText: `MSK_API_KEY=... MSK_API_SECRET=... msk fetch-clusters --all --page-size 50
msk fetch-clusters --all
msk fetch-clusters --project-id 123 --project-id 456 --page-size 20
`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "api-endpoint-base",
			Usage: "TiDB Cloud API endpoint base",
			Value: "https://api.tidbcloud.com/api/v1beta",
		},
		&cli.StringFlag{
			Name:    "api-key",
			Usage:   "API key for authentication with TiDB Cloud API",
			Sources: cli.EnvVars("MSK_API_KEY"),
			Hidden:  true, // accept only from environment variable
		},
		&cli.StringFlag{
			Name:    "api-secret",
			Usage:   "API secret for authentication with TiDB Cloud API",
			Sources: cli.EnvVars("MSK_API_SECRET"),
			Hidden:  true, // accept only from environment variable
		},
		&cli.StringSliceFlag{
			Name:  "project-id",
			Usage: "Target project ID(s) (can be specified multiple times)",
		},
		&cli.IntFlag{
			Name:  "page-size",
			Usage: "Number of clusters per page",
			Value: 20,
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "Fetch all clusters for all active projects in the database",
		},
		&cli.StringFlag{
			Name:  "db-host",
			Usage: "Database host for storing clusters (and reading active projects)",
			Value: "127.0.0.1",
		},
		&cli.StringFlag{
			Name:  "db-user",
			Usage: "Database user for storing clusters (and reading active projects)",
			Value: "root",
		},
		&cli.StringFlag{
			Name:  "db-name",
			Usage: "Database name for storing clusters (and reading active projects)",
			Value: "test",
		},
		&cli.IntFlag{
			Name:  "db-port",
			Usage: "Database port for storing clusters (and reading active projects)",
			Value: 4000,
		},
		&cli.StringFlag{
			Name:    "db-password",
			Usage:   "Database password for storing clusters (and reading active projects)",
			Sources: cli.EnvVars("MSK_DB_PASSWORD"),
			Value:   "",
			Hidden:  true, // accept only from environment variable
		},
		&cli.DurationFlag{
			Name:  "http-timeout",
			Usage: "Timeout for the HTTP request to the TiDB Cloud API. (duration, e.g. 30s, 1m)",
			Value: 30 * time.Second, // Default to 30 seconds
		},
		&cli.DurationFlag{
			Name:  "job-timeout",
			Usage: "Timeout for the entire job. (duration, e.g. 180s, 5m)",
			Value: 180 * time.Second, // Default to 3 minutes
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return runFetchClustersCmd(ctx, c)
	},
}

type fetchClustersArgs struct {
	APIKey          string
	APISecret       string
	APIEndpointBase string
	ProjectIDs      []string
	PageSize        int
	All             bool
	DBHost          string
	DBUser          string
	DBName          string
	DBPort          int
	DBPassword      string
	HTTPTimeout     time.Duration // time.Second
	JobTimeout      time.Duration // time.Second
}

func parseFetchClustersArgs(c *cli.Command) *fetchClustersArgs {
	return &fetchClustersArgs{
		APIKey:          c.String("api-key"),
		APISecret:       c.String("api-secret"),
		APIEndpointBase: c.String("api-endpoint-base"),
		ProjectIDs:      c.StringSlice("project-id"),
		PageSize:        c.Int("page-size"),
		All:             c.Bool("all"),
		DBHost:          c.String("db-host"),
		DBUser:          c.String("db-user"),
		DBName:          c.String("db-name"),
		DBPort:          c.Int("db-port"),
		DBPassword:      c.String("db-password"),
		HTTPTimeout:     c.Duration("http-timeout"),
		JobTimeout:      c.Duration("job-timeout"),
	}
}

func validateFetchClustersArgs(v *fetchClustersArgs) error {
	if v.APIKey == "" {
		return fmt.Errorf("api key is not allowed to be empty")
	}

	if v.APISecret == "" {
		return fmt.Errorf("api secret is not allowed to be empty")
	}

	// fool proofing for API endpoint base
	if v.APIEndpointBase == "" {
		v.APIEndpointBase = "https://api.tidbcloud.com/api/v1beta"
	}

	if v.PageSize <= 0 || v.PageSize > 100 {
		return fmt.Errorf("page-size must be a positive integer less than or equal to 100")
	}

	if v.All && len(v.ProjectIDs) > 0 {
		return fmt.Errorf("--all cannot be used together with --project-id")
	}

	if !v.All && len(v.ProjectIDs) == 0 {
		return fmt.Errorf("project-id is not allowed to be empty if --all is not set")
	}

	if v.DBPort <= 0 || v.DBPort > 65535 {
		return fmt.Errorf("db-port must be a valid TCP port (1–65535)")
	}

	if v.HTTPTimeout <= 0 {
		return fmt.Errorf("http-timeout must be a positive duration")
	}

	if v.JobTimeout <= 0 {
		return fmt.Errorf("job-timeout must be a positive duration")
	}

	return nil
}

func runFetchClustersCmd(ctx context.Context, c *cli.Command) error {
	// Parse command line arguments
	args := parseFetchClustersArgs(c)
	if err := validateFetchClustersArgs(args); err != nil {
		return fmt.Errorf("failed to parse fetch clusters arguments: %w", err)
	}

	// Set the context with a timeout for the entire job
	ctx, cancel := context.WithTimeout(ctx, args.JobTimeout)
	defer cancel()

	// Generate database connection string
	dbDSN := util.GetDBConnectionString(args.DBHost, args.DBUser, args.DBPassword, args.DBName, args.DBPort)

	transport := &digest.Transport{
		Username: args.APIKey,
		Password: args.APISecret,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   args.HTTPTimeout,
	}
	defer client.CloseIdleConnections()

	// Initialize fetcher with the API endpoint and client
	fetcher := clusters.NewAPIClusterFetcher(client, args.APIEndpointBase)

	// Initialize the store service with the database connection string
	store, err := clusters.NewDBClusterStore(dbDSN, nil)
	if err != nil {
		return fmt.Errorf("failed to create cluster store: %w", err)
	}
	defer store.Close()

	projectIDs := args.ProjectIDs
	// If --all is set, fetch all active projects from the database
	if args.All {
		projectStore, err := project.NewDBProjectStore(dbDSN, nil)
		if err != nil {
			return fmt.Errorf("failed to create project store: %w", err)
		}
		defer projectStore.Close()

		activeProjectIDs, err := projectStore.ListActiveProjects(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch active project IDs: %w", err)
		}

		if len(activeProjectIDs) == 0 {
			return fmt.Errorf("no active projects (cluster_count > 0 on latest fetched_at) found in the database")
		}
		projectIDs = activeProjectIDs
	}

	svc := clusters.NewClusterService(fetcher, store)
	if projectNum, clusterNum, deletedClusterNum, err := svc.FetchAndStoreClusters(ctx, projectIDs, args.PageSize); err != nil {
		return err
	} else {
		fmt.Fprintf(c.Root().Writer, "Clusters fetched and stored successfully. Projects: %d, Clusters: %d, Deleted clusters: %d\n", projectNum, clusterNum, deletedClusterNum)
	}

	return nil
}
