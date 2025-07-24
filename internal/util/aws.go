package util

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// PrintAWSVariables prints AWS related variables to the provided writer.
func PrintAWSVariables(ctx context.Context, w io.Writer) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	accountID, err := GetCallerAccountID(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID: %w", err)
	}

	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve AWS credentials: %w", err)
	}

	fmt.Fprintln(w, "AWS Context variables")
	fmt.Fprintf(w, "    Account ID: %s\n", accountID)
	fmt.Fprintf(w, "    Region    : %s\n", cfg.Region)
	fmt.Fprintf(w, "    Credential: %s\n", credentials.Source)

	return nil
}

// GetCallerAccountID retrieves the AWS account ID of the caller.
func GetCallerAccountID(ctx context.Context, cfg aws.Config) (string, error) {
	stsClient := sts.NewFromConfig(cfg)
	output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("unable to get caller identity: %w", err)
	}

	return aws.ToString(output.Account), nil
}

func GetNameFromTags(tags []ec2types.Tag) string {
	if tags == nil {
		return ""
	}

	for _, tag := range tags {
		if strings.ToLower(aws.ToString(tag.Key)) == "name" {
			return aws.ToString(tag.Value)
		}
	}

	return ""
}
