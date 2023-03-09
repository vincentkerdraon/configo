package lambdaconf

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/lambdaconf/constraint"
)

func TestPrepareNewSecretFormatted(t *testing.T) {
	now, err := time.Parse(time.DateTime, "2023-03-08 15:04:05")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		now        time.Time
		lambdaConf LambdaConf
		secretARN  string
		secretOld  string
	}
	tests := []struct {
		name          string
		args          args
		wantSecretNew string
	}{
		{
			name: "ok with TimeAnd8AlphaNum",
			args: args{
				now: now,
				lambdaConf: LambdaConf{
					Secrets: map[string]JSONSecretRotationConf{
						"secretARN1": {
							Keys: map[string]JSONSecretRotationKeyConf{
								"key2": {
									Constraint:     constraint.ConstraintAlphaNum,
									Prefix:         "pre",
									WithTime:       true,
									AlphaNumLength: 16,
								},
							},
						},
					},
				},
				secretARN: "secretARN1",
				secretOld: `{"key1":"val1","key2":"val2"}`,
			},
			wantSecretNew: `{"key1":"val1","key2":"pre-20230308150405-TGJV7qbsxJjpyYGJ"}`,
		},
		{
			name: "ok when ignore",
			args: args{
				now: now,
				lambdaConf: LambdaConf{
					Secrets: map[string]JSONSecretRotationConf{
						"secretARN1": {
							Keys: map[string]JSONSecretRotationKeyConf{},
						},
					},
				},
				secretARN: "secretARN1",
				secretOld: `{"key1":"val1","key2":"val2"}`,
			},
			wantSecretNew: `{"key1":"val1","key2":"val2"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rand.Seed(tt.args.now.UnixNano())
			gotF := PrepareNewSecretFormatted(tt.args.now, tt.args.lambdaConf)
			secretNew, err := gotF(context.Background(), tt.args.secretARN, tt.args.secretOld)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(secretNew, tt.wantSecretNew) {
				t.Errorf("PrepareNewSecretFormatted() = %v, want %v", secretNew, tt.wantSecretNew)
			}
		})
	}
}
