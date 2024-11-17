package targets

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func TestGetTargets(t *testing.T) {
	// TODO mock and test CloudWatchLogsTarget
	
	tests := []struct {
		name          string
		envVars       map[string]string
		targetsConfig string
		expectErr     bool
		expectTargets int
	}{
		{
			name:          "Single valid target - stdout",
			targetsConfig: TargetStdout,
			expectErr:     false,
			expectTargets: 1,
		},
		{
			name:          "Unsupported target type",
			targetsConfig: "unsupported",
			expectErr:     true,
			expectTargets: 0,
		},
		{
			name:          "Mixed valid and invalid targets",
			targetsConfig: TargetStdout + ",unsupported",
			expectErr:     false,
			expectTargets: 1,
		},
		{
			name:          "Empty target configuration",
			targetsConfig: "",
			expectErr:     true,
			expectTargets: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			for key, value := range test.envVars {
				t.Setenv(key, value)
			}

			mockSession := &session.Session{}

			targets, err := GetTargets(test.targetsConfig, mockSession)

			if test.expectErr {
				assert.Error(t, err)
				assert.Nil(t, targets)
			} else {
				assert.NoError(t, err)
				assert.Len(t, targets, test.expectTargets)
			}
		})
	}
}
