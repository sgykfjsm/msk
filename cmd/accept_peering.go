package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/sgykfjsm/msk/internal/util"
	"github.com/sgykfjsm/msk/internal/vpcpeering"
	"github.com/urfave/cli/v3"
)

var AcceptPeeringCmd = &cli.Command{
	Name:  "accept-peering",
	Usage: "Accept a VPC peering connection request by ID if it is in the 'pending-acceptance' state.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "peering-id",
			Usage:    "ID of the VPC peering connection to accept",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "check-only",
			Usage: "Check the state of the VPC peering. No changes will be made if this flag is set",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Perform a dry run without making any changes",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		peeringID := c.String("peering-id")
		if peeringID == "" {
			return fmt.Errorf("peering-id is required")
		} else if !strings.HasPrefix(peeringID, "pcx-") {
			return fmt.Errorf("peering-id should start with 'pcx-' prefix, please provide the actual peering connection ID with the prefix")
		}

		checkOnly := c.Bool("check-only")
		dryRun := c.Bool("dry-run")
		if checkOnly && dryRun {
			return fmt.Errorf("cannot use both check-only and dry-run flags together")
		}

		if dryRun {
			fmt.Printf("Dry run: would accept VPC peering connection with ID %q\n", peeringID)
			util.PrintAWSVariables(ctx, c.Root().Writer)
			return nil
		}

		return vpcpeering.AcceptVPCPeeringConnection(ctx, peeringID, c.Root().Writer, checkOnly)
	},
}
