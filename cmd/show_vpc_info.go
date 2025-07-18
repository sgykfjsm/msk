package cmd

import (
	"context"
	"fmt"

	"github.com/sgykfjsm/msk/internal/vpcinfo"
	"github.com/urfave/cli/v3"
)

var ShowVPCInfoCmd = &cli.Command{
	Name:  "show-vpc-info",
	Usage: "Show CIDR block and AWS account ID of the VPC",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "vpc-id",
			Usage:    "The ID of the VPC to show information for",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "Output format (json, yaml, text)",
			Value: "text",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		vpcID := c.String("vpc-id")
		if vpcID == "" {
			return fmt.Errorf("vpc-id is not allowed to be empty")
		}

		outputFormat := c.String("output")

		// We expect AWS related variables to be set in the environment
		vpcInfo, err := vpcinfo.FetchVPCInfo(ctx, vpcID)
		if err != nil {
			return fmt.Errorf("failed to fetch VPC info: %w", err)
		}

		return vpcInfo.PrintAs(outputFormat, c.Root().Writer)
	},
}
