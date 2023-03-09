package awssecretmanagerrotationlambda

import (
	"context"
)

type (
	Option func(*impl)
)

func WithAWSSecretsManager(svc AWSSecretsManager) func(*impl) {
	return func(r *impl) {
		r.svc = svc
	}
}
func WithLogger(logger LeveledLogger) func(*impl) {
	return func(r *impl) {
		r.logger = logger
	}
}
func WithPrepareSecret(prepareSecret func(ctx context.Context, secretARN string, secretOld string) (secretNew string, _ error)) func(*impl) {
	return func(r *impl) {
		r.prepareSecret = prepareSecret
	}
}
func WithSetSecret(setSecret func(ctx context.Context, secretARN string, versionID string) error) func(*impl) {
	return func(r *impl) {
		r.setSecret = setSecret
	}
}
func WithTestSecret(testSecret func(ctx context.Context, secretARN string, versionID string) error) func(*impl) {
	return func(r *impl) {
		r.testSecret = testSecret
	}
}
