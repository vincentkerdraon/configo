package awssecretmanager

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/versionstage"
	"github.com/vincentkerdraon/configo/secretrotation"
)

type awsSecretsManagerMock struct {
	res map[versionstage.VersionStage]string
}

func (m *awsSecretsManagerMock) GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	s := m.res[versionstage.VersionStage(*input.VersionStage)]
	return &secretsmanager.GetSecretValueOutput{
		SecretString: &s,
	}, nil
}

type cacheMock struct {
	m map[interface{}]interface{}
}

func (c cacheMock) Add(key, value interface{}) {
	c.m[key] = value
}
func (c cacheMock) Get(key interface{}) (value interface{}, ok bool) {
	value, ok = c.m[key]
	return value, ok
}

func Test_impl_LoadValueWhenPlainText(t *testing.T) {
	type args struct {
		secretName string
	}
	tests := []struct {
		name             string
		svcSecretManager AWSSecretsManager
		args             args
		want             secretrotation.Secret
		wantErr          bool
	}{
		{
			name:             "when ok, value + plain text",
			svcSecretManager: &awsSecretsManagerMock{res: map[int]string{0: "secret"}},
			args:             args{secretName: "secretName"},
			want:             "secret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := cacheMock{m: make(map[interface{}]interface{})}
			sm := New(tt.svcSecretManager, cache)
			got, fromCache, err := sm.LoadValueWhenPlainText(context.Background(), tt.args.secretName)
			if (err != nil) != tt.wantErr {
				t.Errorf("impl.LoadValueWhenPlainText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if *got != tt.want {
				t.Errorf("impl.LoadValueWhenPlainText() = %v, want %v", got, tt.want)
			}
			if fromCache {
				t.Error()
			}

			//second time to use the cache
			got, fromCache, err = sm.LoadValueWhenPlainText(context.Background(), tt.args.secretName)
			if (err != nil) != tt.wantErr {
				t.Errorf("impl.LoadValueWhenPlainText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if *got != tt.want {
				t.Errorf("impl.LoadValueWhenPlainText() = %v, want %v", got, tt.want)
			}
			if !fromCache {
				t.Error()
			}
		})
	}
}

func Test_impl_LoadRotatingSecretWhenJSON(t *testing.T) {
	type args struct {
		secretName string
		secretKey  string
	}
	tests := []struct {
		name             string
		svcSecretManager AWSSecretsManager
		args             args
		want             secretrotation.RotatingSecret
		wantErr          bool
	}{
		{
			name: "when ok, rotating secret + JSON",
			svcSecretManager: &awsSecretsManagerMock{res: map[int]string{
				0: `{"key":"previous"}`,
				1: `{"key":"current"}`,
				2: `{"key":"pending"}`,
			}},
			args: args{secretName: "secretName", secretKey: "key"},
			want: secretrotation.NewRotatingSecret("previous", "current", "pending"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := cacheMock{m: make(map[interface{}]interface{})}
			sm := New(tt.svcSecretManager, cache)
			got, fromCache, err := sm.LoadRotatingSecretWhenJSON(context.Background(), tt.args.secretName, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("impl.LoadRotatingSecretWhenJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("impl.LoadRotatingSecretWhenJSON()\ngot =%q\nwant=%q", got, tt.want)
			}
			if fromCache {
				t.Error()
			}

			//second time to use the cache
			got, fromCache, err = sm.LoadRotatingSecretWhenJSON(context.Background(), tt.args.secretName, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("impl.LoadRotatingSecretWhenJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("impl.LoadRotatingSecretWhenJSON()\ngot =%q\nwant=%q", got, tt.want)
			}
			if !fromCache {
				t.Error()
			}
		})
	}
}

func Test_LoadRotatingSecretWhenJSON_scenario(t *testing.T) {
	cache := cacheMock{m: make(map[interface{}]interface{})}
	svcSecretManager := &awsSecretsManagerMock{res: map[int]string{
		0: `{"key1":"previous1","key2":"previous2","key3":""}`,
		1: `{"key1":"current1","key2":"current2","key3":"current2"}`,
		2: `{"key1":"pending1","key2":"pending2","key3":"pending2"}`,
	}}

	sm := New(svcSecretManager, cache)

	got, fromCache, err := sm.LoadRotatingSecretWhenJSON(context.Background(), "whatever", "key1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Current != "current1" {
		t.Fatal()
	}
	if fromCache {
		t.Fatal()
	}

	//Another JSON key, should be in cache
	got, fromCache, err = sm.LoadRotatingSecretWhenJSON(context.Background(), "whatever", "key2")
	if err != nil {
		t.Fatal(err)
	}
	if got.Current != "current2" {
		t.Fatal()
	}
	if !fromCache {
		t.Fatal()
	}

	//key exists but previous is empty
	got, fromCache, err = sm.LoadRotatingSecretWhenJSON(context.Background(), "whatever", "key3")
	if err == nil {
		t.Fatal()
	}
	if got != nil {
		t.Fatal()
	}
	if !fromCache {
		t.Fatal()
	}

	//Non existant JSON key
	got, fromCache, err = sm.LoadRotatingSecretWhenJSON(context.Background(), "whatever", "keyNonExistant")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal()
	}
	if !fromCache {
		t.Fatal()
	}

}
