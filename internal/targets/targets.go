package targets

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jdwit/alb-log-pipe/internal/types"
	"strings"
)

const (
	TargetCloudWatch = "cloudwatch"
	TargetStdout     = "stdout"
)

type Target interface {
	SendLogs(entryChain <-chan types.LogEntry)
}

func GetTargets(targetsConfig string, sess *session.Session) ([]Target, error) {
	targetTypes := strings.Split(targetsConfig, ",")
	var targets []Target

	for _, t := range targetTypes {
		var target Target
		var err error

		switch t {
		case TargetCloudWatch:
			target, err = NewCloudWatchTarget(sess)
		case TargetStdout:
			target = NewStdoutTarget()
		default:
			fmt.Printf("warning: unsupported target type: %s", t)
			continue
		}

		// Skip any targets that fail to initialize due to missing config or other errors
		if err != nil {
			fmt.Printf("warning: could not initialize target %s: %v\n", t, err)
			continue
		}

		targets = append(targets, target)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("error: no valid targets initialized")
	}

	return targets, nil
}
