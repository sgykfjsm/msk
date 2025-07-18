package vpcinfo

import "testing"

func TestFetch_VPCInfo_ToString(t *testing.T) {
	vpcInfo := &VPCInfo{
		VPCID:     "vpc-12345678",
		CIDRBlock: "10.0.0/16",
		AccountID: "123456789012",
	}
	expected := "VPC ID: vpc-12345678, CIDR Block: 10.0.0/16, AWS Account ID: 123456789012"
	if vpcInfo.String() != expected {
		t.Errorf("Expected %q, got %q", expected, vpcInfo.String())
	}
}

func TestFetch_VPCInfo_ToJSON(t *testing.T) {
	vpcInfo := &VPCInfo{
		VPCID:     "vpc-12345678",
		CIDRBlock: "10.0.0/16",
		AccountID: "123456789012",
	}
	data, err := vpcInfo.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}
	expected := `{"vpc_id":"vpc-12345678","cidr_block":"10.0.0/16","aws_account_id":"123456789012"}`
	if string(data) != expected {
		t.Errorf("Expected %q, got %q", expected, string(data))
	}
}

func TestFetch_VPCInfo_ToYAML(t *testing.T) {
	vpcInfo := &VPCInfo{
		VPCID:     "vpc-12345678",
		CIDRBlock: "10.0.0/16",
		AccountID: "123456789012",
	}
	data, err := vpcInfo.ToYAML()
	if err != nil {
		t.Fatalf("Failed to convert to YAML: %v", err)
	}
	expected := `vpc_id: vpc-12345678
cidr_block: 10.0.0/16
aws_account_id: "123456789012"
`
	if string(data) != expected {
		t.Errorf("Expected %q, got %q", expected, string(data))
	}
}
