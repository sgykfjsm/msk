package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sgykfjsm/msk/internal/util"
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
			Usage: "Output format (json, yaml, text) case-insensitive, defaults to text",
			Value: "text",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "If set, the command will not perform any actions but will show given parameters",
			Value: false,
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		vpcID := c.String("vpc-id")
		if vpcID == "" {
			return fmt.Errorf("vpc-id is not allowed to be empty")
		} else if !strings.HasPrefix(vpcID, "vpc-") {
			return fmt.Errorf("vpc-id should start with 'vpc-' prefix, please provide the actual VPC ID with the prefix")
		}

		outputFormat := c.String("output")
		allowedFormats := map[string]bool{"json": true, "yaml": true, "text": true}
		if !allowedFormats[strings.ToLower(outputFormat)] {
			return fmt.Errorf("invalid output format: %s, allowed formats are: %v", outputFormat, allowedFormats)
		}

		if c.Bool("dry-run") {
			fmt.Printf("Dry run: would fetch VPC info for VPC ID %q with output format %q\n", vpcID, outputFormat)
			util.PrintAWSVariables(ctx, c.Root().Writer)
			return nil
		}

		// We expect AWS related variables to be set in the environment
		vpcInfo, err := vpcinfo.FetchVPCInfo(ctx, vpcID)
		if err != nil {
			return fmt.Errorf("failed to fetch VPC info: %w", err)
		}

		return vpcInfo.PrintAs(outputFormat, c.Root().Writer)
	},
}
