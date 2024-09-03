package settings

import (
	"fmt"

	"observer/pkg/version"
)

func Version() string {
	return fmt.Sprintf("Version: pinger-%s\n", version.Get())
}
