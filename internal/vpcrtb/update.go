package vpcrtb

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sgykfjsm/msk/internal/util"
)

// UpdateRoutes updates the routes for a given VPC with the specified CIDR and peer ID.
func UpdateRoutes(ctx context.Context, vpcID, cidr, peerID string, dryRun bool, w io.Writer) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// 1. Find target route tables for specified VPC with ID and CIDR
	// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeRouteTables
	param := &ec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"),
				Values: []string{vpcID}},
		},
	}
	rtOutput, err := ec2Client.DescribeRouteTables(ctx, param)
	if err != nil {
		return fmt.Errorf("failed to describe route tables for VPC %s: %w", vpcID, err)
	}
	if len(rtOutput.RouteTables) == 0 {
		fmt.Fprintf(w, "No route tables found for VPC %s\n", vpcID)
		return nil
	}
	fmt.Fprintf(w, "Found %d route tables for VPC %s\n", len(rtOutput.RouteTables), vpcID)

	// 2. Check the route tables one by one. If dryRun is true, only simulate the update without making changes.
	// 2-1. If the route table already has a route for the CIDR
	// 		2-1-1. If the peer ID is same, skip it
	// 		2-1-2. If the peer ID is different, update it
	// 2-2. If the route table does not have a route for the CIDR, add a new route with the peer ID
	for _, rtb := range rtOutput.RouteTables {
		rtbID := aws.ToString(rtb.RouteTableId)

		var found, updateNeeded bool
		var currentPeerID string
		for _, rt := range rtb.Routes {
			if aws.ToString(rt.DestinationCidrBlock) == cidr {
				found = true
				if rt.VpcPeeringConnectionId == nil {
					updateNeeded = true
					currentPeerID = "unset"
				} else if aws.ToString(rt.VpcPeeringConnectionId) != peerID {
					updateNeeded = true
					currentPeerID = aws.ToString(rt.VpcPeeringConnectionId)
				}
				break
			}
		}

		rtbName := util.GetNameFromTags(rtb.Tags)
		switch {
		case !found: // No route found for the CIDR
			if dryRun {
				fmt.Fprintf(w, "[DRY RUN] Would add route for CIDR %s to route table %q (ID: %s) with peer ID %s\n",
					cidr, rtbName, rtbID, peerID)
				continue
			}

			params := &ec2.CreateRouteInput{
				RouteTableId:           aws.String(rtbID),
				DestinationCidrBlock:   aws.String(cidr),
				VpcPeeringConnectionId: aws.String(peerID),
			}

			// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.CreateRoute
			if _, err := ec2Client.CreateRoute(ctx, params); err != nil {
				return fmt.Errorf("failed to create route for CIDR %s in route table %q (ID: %s): %w",
					cidr, rtbName, rtbID, err)
			}
			fmt.Fprintf(w, "[SUCCESS] Route for CIDR %s added to route table %q (ID: %s) with peer ID %s\n",
				cidr, rtbName, rtbID, peerID)

		case updateNeeded: // Route found but needs update
			if dryRun {
				fmt.Fprintf(w, "[DRY RUN] Would update route for CIDR %s in route table %q (ID: %s) from peer ID %s to peer ID %s\n",
					cidr, rtbName, rtbID, currentPeerID, peerID)
				continue
			}
			// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.ReplaceRoute
			param := &ec2.ReplaceRouteInput{
				RouteTableId:           aws.String(rtbID),
				DestinationCidrBlock:   aws.String(cidr),
				VpcPeeringConnectionId: aws.String(peerID),
			}
			if _, err := ec2Client.ReplaceRoute(ctx, param); err != nil {
				return fmt.Errorf("failed to replace route for CIDR %s in route table %q (ID: %s): %w",
					cidr, rtbName, rtbID, err)
			}
			fmt.Fprintf(w, "[SUCCESS] Route for CIDR %s updated in route table %q (ID: %s) from peer ID %s to peer ID %s\n",
				cidr, rtbName, rtbID, currentPeerID, peerID)

		default: // Route found and no update needed
			fmt.Fprintf(w, "Route for CIDR %s already exists in route table %q (ID: %s) with peer ID %s, skipping\n",
				cidr, rtbName, rtbID, peerID)
		}
	}

	return nil
}
