package testcontainers

import (
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

func init() {
	tclog.SetDefault(tclog.NewNoopLogger())

	// testcontainers are slow to boot and sometimes fail on default timeout
	if os.Getenv("RYUK_RECONNECTION_TIMEOUT") == "" && os.Getenv("TESTCONTAINERS_RYUK_RECONNECTION_TIMEOUT") == "" {
		_ = os.Setenv("RYUK_RECONNECTION_TIMEOUT", "5m")
	}
}

func noop() {}

func terminate(container testcontainers.Container) {
	if err := testcontainers.TerminateContainer(container); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "terminate container: %v\n", err)
	}
}
