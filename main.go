package main

import (
	"context"
	"log/slog"
	"os"

	mskcmd "github.com/sgykfjsm/msk/cmd"
	"github.com/urfave/cli/v3"
)

var Version = "v0.1.0"

func main() {
	cmd := &cli.Command{
		Name:    "msk",
		Usage:   "Manage TiDB Cloud clusters, Support daily operations, and Keep your workloads efficient",
		Version: Version,

		Commands: []*cli.Command{
			mskcmd.FetchProjectsCmd,
			{
				Name:  "fetch-clusters",
				Usage: "Get information about TiDB Clusters and save it to a database",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					slog.Info("Collecting cluster information...")
					return nil
				},
			},
			{
				Name:  "generate-notice",
				Usage: "Generate a notice from the information collected by fetch-clusters and save it to S3",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					slog.Info("Generating notice from collected information...")
					return nil
				},
			},
			{
				Name:  "notify",
				Usage: "Notify via messaging service using the generated notice",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					slog.Info("Sending notification with the generated notice...")
					return nil
				},
			},
			mskcmd.ShowVPCInfoCmd,
			mskcmd.AcceptPeeringCmd,
			mskcmd.UpdateRoutesCmd,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error("Error running command", "error", err)
		os.Exit(1)
	}
}
