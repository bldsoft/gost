package testcontainers

import (
	"fmt"
	"os"

	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
)

func init() {
	tclog.SetDefault(tclog.NewNoopLogger())
}

func noop() {}

func terminate(container testcontainers.Container) {
	if err := testcontainers.TerminateContainer(container); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "terminate container: %v\n", err)
	}
}
