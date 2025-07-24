package cmd

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/sgykfjsm/msk/internal/util"
	"github.com/sgykfjsm/msk/internal/vpcpeering"
	"github.com/sgykfjsm/msk/internal/vpcrtb"
	"github.com/urfave/cli/v3"
)

var UpdateRoutesCmd = &cli.Command{
	Name:  "update-routes",
	Usage: "Update routes for a VPC with the specified CIDR and peer ID",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "vpc-id",
			Usage:    "ID of the VPC to update routes for. It should start with 'vpc-' prefix",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "cidr",
			Usage:    "CIDR block to route (e.g., 192.168.1.0/24)",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "peer-id",
			Usage:    "Peer ID for the VPC route update. It should start with 'pcx-' prefix",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "If true, only simulate the update without making changes",
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

		cidr := c.String("cidr")
		if cidr == "" {
			return fmt.Errorf("cidr is not allowed to be empty")
		} else if _, _, err := net.ParseCIDR(cidr); err != nil {
			return fmt.Errorf("cidr %q should be in CIDR notation, e.g., 192.168.1.0/24", cidr)
		}

		peerID := c.String("peer-id")
		if peerID == "" {
			return fmt.Errorf("peer-id is not allowed to be empty")
		} else if !strings.HasPrefix(peerID, "pcx-") {
			return fmt.Errorf("peer-id should start with 'pcx-' prefix, please provide the actual Peer ID with the prefix")
		}

		dryRun := c.Bool("dry-run")
		if dryRun {
			fmt.Fprintf(c.Root().Writer, "Dry run: would update routes for VPC %q with CIDR %q and peer ID %q\n", vpcID, cidr, peerID)
			// Check if the target VPC peering connection is already accepted
			if err := vpcpeering.AcceptVPCPeeringConnection(ctx, peerID, c.Root().Writer, true); err != nil {
				fmt.Fprintf(c.Root().Writer, "failed to check VPC peering connection: %v\n", err)
			}

			util.PrintAWSVariables(ctx, c.Root().Writer)
		}

		if err := vpcrtb.UpdateRoutes(ctx, vpcID, cidr, peerID, dryRun, c.Root().Writer); err != nil {
			return err
		}

		return nil
	},
}
