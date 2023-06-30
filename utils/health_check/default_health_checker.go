package health_check

import "context"

var DefaultHealthChecker HealthChecker

func Check(ctx context.Context, url string) error {
	return DefaultHealthChecker.HealthCheck(ctx, url)
}
