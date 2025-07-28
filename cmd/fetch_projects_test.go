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
			args:  &fetchProjectArgs{APIKey: "test-token", APISecret: "test-secret"},
			isErr: false,
		}, {
			name:  "api-key is empty",
			args:  &fetchProjectArgs{APIKey: "", APISecret: "test-secret"},
			isErr: true,
		}, {
			name:  "api-secret is empty",
			args:  &fetchProjectArgs{APIKey: "test-token", APISecret: ""},
			isErr: true,
		}, {
			name:  "all empty",
			args:  &fetchProjectArgs{APIKey: "", APISecret: ""},
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
