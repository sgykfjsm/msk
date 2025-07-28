package cmd

import (
	"context"
	"fmt"

	"github.com/sgykfjsm/msk/internal/project"
	"github.com/sgykfjsm/msk/internal/util"
	"github.com/urfave/cli/v3"
)

var FetchProjectsCmd = &cli.Command{
	Name:  "fetch-projects",
	Usage: "Fetch and store projects from the TiDB Cloud API",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "api-endpoint",
			Usage: "TiDB Cloud API endpoint",
			Value: "https://api.tidbcloud.com/v1beta/projects",
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
		&cli.IntFlag{
			Name:  "page",
			Usage: "Page number for pagination",
			Value: 1,
		},
		&cli.IntFlag{
			Name:  "page-size",
			Usage: "Number of projects per page",
			Value: 20,
		},
		&cli.StringFlag{
			Name:  "db-host",
			Usage: "Database host for storing projects",
			Value: "127.0.0.1",
		},
		&cli.StringFlag{
			Name:  "db-user",
			Usage: "Database user for storing projects",
			Value: "root",
		},
		&cli.StringFlag{
			Name:  "db-name",
			Usage: "Database name for storing projects",
			Value: "test",
		},
		&cli.IntFlag{
			Name:  "db-port",
			Usage: "Database port for storing projects",
			Value: 4000,
		},
		&cli.StringFlag{
			Name:    "db-password",
			Usage:   "Database password for storing projects",
			Sources: cli.EnvVars("MSK_DB_PASSWORD"),
			Value:   "",
			Hidden:  true, // accept only from environment variable
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return runFetchProjects(ctx, c)
	},
}

func runFetchProjects(ctx context.Context, c *cli.Command) error {
	// Parse command line arguments
	args := parseFetchProjectArgs(c)
	if err := validateFetchProjectArgs(args); err != nil {
		return fmt.Errorf("failed to parse fetch project arguments: %w", err)
	}

	// Generate database connection string
	dbDSN := util.GetDBConnectionString(args.DBHost, args.DBUser, args.DBPassword, args.DBName, args.DBPort)

	// Initialize fetcher with the API endpoint and token
	fetcher := project.NewAPIProjectFetcher(args.APIKey, args.APISecret, args.APIEndpoint)

	// Initialize the store service with the database connection string
	store, err := project.NewDBProjectStore(dbDSN, nil) // Assuming no special pool config needed
	if err != nil {
		return fmt.Errorf("failed to create project store: %w", err)
	}

	if err := runFetchAndStoreProjectsService(ctx, fetcher, store, args.Page, args.PageSize); err != nil {
		return fmt.Errorf("failed to fetch and store projects: %w", err)
	}

	fmt.Fprintln(c.Root().Writer, "Projects fetched and stored successfully.")
	return nil
}

func runFetchAndStoreProjectsService(ctx context.Context, fetcher project.ProjectFetcher, store project.ProjectStore, page int, pageSize int) error {
	if err := project.NewProjectService(fetcher, store).FetchAndStoreProjects(ctx, page, pageSize); err != nil {
		return fmt.Errorf("failed to fetch and store projects: %w", err)
	}

	return nil
}

type fetchProjectArgs struct {
	APIKey      string
	APISecret   string
	APIEndpoint string
	Page        int
	PageSize    int
	DBHost      string
	DBUser      string
	DBName      string
	DBPort      int
	DBPassword  string
}

func parseFetchProjectArgs(c *cli.Command) *fetchProjectArgs {
	return &fetchProjectArgs{
		APIKey:      c.String("api-key"),
		APISecret:   c.String("api-secret"),
		APIEndpoint: c.String("api-endpoint"),
		Page:        c.Int("page"),
		PageSize:    c.Int("page-size"),
		DBHost:      c.String("db-host"),
		DBUser:      c.String("db-user"),
		DBName:      c.String("db-name"),
		DBPort:      c.Int("db-port"),
		DBPassword:  c.String("db-password"),
	}
}

func validateFetchProjectArgs(v *fetchProjectArgs) error {
	if v.APIKey == "" {
		return fmt.Errorf("API key is not allowed to be empty")
	}

	if v.APISecret == "" {
		return fmt.Errorf("API secret is not allowed to be empty")
	}

	return nil
}
