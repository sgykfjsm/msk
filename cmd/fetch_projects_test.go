package cmd

import "testing"

func TestFetchProjects_validateFetchProjectArgs(t *testing.T) {
	tests := []struct {
		name  string
		args  *fetchProjectArgs
		isErr bool
	}{
		{
			name:  "valid args",
			args:  &fetchProjectArgs{OrgID: "testOrg123", APIToken: "test-token"},
			isErr: false,
		}, {
			name:  "org-id is empty",
			args:  &fetchProjectArgs{OrgID: "", APIToken: "test-token"},
			isErr: true,
		}, {
			name:  "api-token is empty",
			args:  &fetchProjectArgs{OrgID: "testOrg123", APIToken: ""},
			isErr: true,
		}, {
			name:  "both empty",
			args:  &fetchProjectArgs{OrgID: "", APIToken: ""},
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateFetchProjectArgs(tt.args); (err != nil) != tt.isErr {
				t.Errorf("validateFetchProjectArgs error=%v, isErr=%t", err, tt.isErr)
			}
		})
	}
}
