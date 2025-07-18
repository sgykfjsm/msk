package vpcpeering

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// AcceptVPCPeeringConnection accepts a VPC peering connection request.
// It takes a context, the ID of the peering connection, an io.Writer for output,
// and a boolean to indicate if the operation is a dry run (check only).
// Returns an error if the operation fails.
func AcceptVPCPeeringConnection(ctx context.Context, peeringID string, w io.Writer, checkOnly bool) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}
	ec2Client := ec2.NewFromConfig(cfg)

	// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeVpcPeeringConnections
	param := &ec2.DescribeVpcPeeringConnectionsInput{VpcPeeringConnectionIds: []string{peeringID}}
	output, err := ec2Client.DescribeVpcPeeringConnections(ctx, param)
	if err != nil {
		return fmt.Errorf("failed to describe VPC peering (ID: %s) connection: %w", peeringID, err)
	}
	if len(output.VpcPeeringConnections) == 0 {
		return fmt.Errorf("no VPC peering connection found with ID: %s", peeringID)
	}

	peering := output.VpcPeeringConnections[0]
	state := string(peering.Status.Code)

	switch state {
	case "pending-acceptance":
		if checkOnly {
			fmt.Fprintf(w, "[CHECK] VPC peering connection %s is in %q.\n", peeringID, state)
			return nil
		}
		fmt.Fprintf(w, "Accepting VPC peering connection %s...\n", peeringID)

		paramAccept := &ec2.AcceptVpcPeeringConnectionInput{
			VpcPeeringConnectionId: aws.String(peeringID),
		}
		// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.AcceptVpcPeeringConnection
		if _, err := ec2Client.AcceptVpcPeeringConnection(ctx, paramAccept); err != nil {
			return fmt.Errorf("failed to accept VPC peering connection (ID: %s): %w", peeringID, err)
		}
		fmt.Fprintf(w, "[SUCCESS] VPC peering connection %s accepted.\n", peeringID)

	case "active":
		fmt.Fprintf(w, "[SKIP] VPC peering connection %s is already %q.\n", peeringID, state)

	default:
		fmt.Fprintf(w, "[SKIP] VPC peering connection %s is in the state %q. No action taken.\n", peeringID, state)
	}

	return nil
}
