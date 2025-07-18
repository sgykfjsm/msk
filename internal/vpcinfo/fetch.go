package vpcinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/sgykfjsm/msk/internal/util"
	"gopkg.in/yaml.v3"
)

// This module provides information about the VPC (Virtual Private Cloud) in which the application is running.
// As of now, following information will be provided using the AWS SDK for Go v2:
// - VPC ID (expected to be passed from the caller)
// - CIDR block
// - Caller's AWS account ID

// Output format is available in:
// - JSON
// - YAML
// - Text as log message

type VPCInfo struct {
	VPCID     string `json:"vpc_id" yaml:"vpc_id"`                 // The ID of the VPC
	CIDRBlock string `json:"cidr_block" yaml:"cidr_block"`         // The CIDR block of the VPC
	AccountID string `json:"aws_account_id" yaml:"aws_account_id"` // The AWS account
}

func (v *VPCInfo) String() string {
	return "VPC ID: " + v.VPCID + ", CIDR Block: " + v.CIDRBlock + ", AWS Account ID: " + v.AccountID
}

func (v *VPCInfo) ToJSON() ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (v *VPCInfo) ToYAML() ([]byte, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func FetchVPCInfo(ctx context.Context, vpcID string) (*VPCInfo, error) {
	vpcInfo := &VPCInfo{
		VPCID:     vpcID,
		CIDRBlock: "",
		AccountID: "", // This would be fetched from the AWS SDK
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	// Fetch the VPC details using the EC2 client
	ec2Client := ec2.NewFromConfig(cfg)
	// Ref: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ec2#Client.DescribeVpcs
	vpcs, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		VpcIds: []string{vpcID},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to describe VPCs: %w", err)
	}

	// Assuming we only want the first VPC returned
	if len(vpcs.Vpcs) == 0 {
		return nil, fmt.Errorf("no VPC found with ID: %s", vpcID)
	}
	targetVPC := vpcs.Vpcs[0]
	vpcInfo.CIDRBlock = *targetVPC.CidrBlock // // We need the primary IPv4 CIDR block for the VPC.

	// Fetch the AWS account ID
	accountID, err := util.GetCallerAccountID(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS account ID: %w", err)
	}
	vpcInfo.AccountID = accountID

	return vpcInfo, nil
}

func (v *VPCInfo) PrintAs(format string, w io.Writer) error {
	switch format {
	case "json":
		data, err := v.ToJSON()
		if err != nil {
			return fmt.Errorf("error converting VPCInfo to JSON: %w", err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("error writing JSON output: %w", err)
		}

	case "yaml":
		data, err := v.ToYAML()
		if err != nil {
			return fmt.Errorf("error converting VPCInfo to YAML: %w", err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("error writing YAML output: %w", err)
		}
	default:
		// Default to text output
		if _, err := fmt.Fprintln(w, v.String()); err != nil {
			return fmt.Errorf("error writing text output: %w", err)
		}
	}

	return nil
}
