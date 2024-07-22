package utils

import (
	"fmt"
	"runtime"
)

func GetUserAgent(version string) string {
	return fmt.Sprintf("terraform-provider-wundergraph/%s Go/%s (%s; %s)",
		version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
