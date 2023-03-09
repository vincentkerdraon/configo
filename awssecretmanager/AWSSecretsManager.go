package awssecretmanager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/vincentkerdraon/configo/awssecretmanager/awssecretmanagerlib/versionstage"
	"github.com/vincentkerdraon/configo/lock"
	"github.com/vincentkerdraon/configo/secretrotation"
)

type (
	Cache interface {
		Add(key, value interface{})
		Get(key interface{}) (value interface{}, ok bool)
	}

	AWSSecretsManager interface {
		//GetSecretValueWithContext grabs the secrets. In case of error, it will retry as per the AWS session configuration.
		GetSecretValueWithContext(ctx context.Context, input *secretsmanager.GetSecretValueInput, opts ...request.Option) (*secretsmanager.GetSecretValueOutput, error)
	}

	Manager interface {
		LoadValueWhenJSON(ctx context.Context, secretName string, secretKey string) (_ *secretrotation.Secret, fromCache bool, _ error)
		LoadValueWhenPlainText(ctx context.Context, secretName string) (_ *secretrotation.Secret, fromCache bool, _ error)
		LoadRotatingSecretWhenJSON(ctx context.Context, secretName string, secretKey string) (_ *secretrotation.RotatingSecret, fromCache bool, _ error)
		LoadRotatingSecretWhenPlainText(ctx context.Context, secretName string) (_ *secretrotation.RotatingSecret, fromCache bool, _ error)
	}

	impl struct {
		cache            Cache
		svcSecretManager AWSSecretsManager
		lock             lock.Locker
	}
)

// impl implements Manager
var _ Manager = (*impl)(nil)

// New creates a manager.
//
// svcSecretManager is the AWS service.
//
// cache=nil means no cache. A cache with TTL is recommended to increase speed and reduce cost.
// See cachelruttl.
func New(svcSecretManager AWSSecretsManager, cache Cache) *impl {
	return &impl{
		svcSecretManager: svcSecretManager,
		cache:            cache,
		lock:             lock.New(),
	}
}

// LoadValue is a helper to load a non-rotating value from the secret manager.
//
// SecretKey is the JSON key (a secret can store multiple values, see AWS doc)
func (sm *impl) LoadValueWhenJSON(ctx context.Context, secretName string, secretKey string) (_ *secretrotation.Secret, fromCache bool, _ error) {

	decode := func(val *secretrotation.Secret) (*secretrotation.Secret, error) {
		return sm.decodeJSONValue(*val, secretKey)
	}
	res, fromCache, err := loadValue(ctx, secretName, sm.loadSecretSimpleValue, decode, sm.cache, sm.lock, secretName)
	if err != nil {
		return nil, fromCache, fmt.Errorf("for secretName=%q, secretKey=%q, %w", secretName, secretKey, err)
	}
	return res, fromCache, nil
}

func (sm *impl) LoadValueWhenPlainText(ctx context.Context, secretName string) (_ *secretrotation.Secret, fromCache bool, _ error) {
	decode := func(val *secretrotation.Secret) (*secretrotation.Secret, error) {
		return val, nil
	}
	res, fromCache, err := loadValue(ctx, secretName, sm.loadSecretSimpleValue, decode, sm.cache, sm.lock, secretName)
	if err != nil {
		return nil, fromCache, fmt.Errorf("for secretName=%q, %w", secretName, err)
	}
	return res, fromCache, nil
}

func (sm *impl) LoadRotatingSecretWhenJSON(ctx context.Context, secretName string, secretKey string) (_ *secretrotation.RotatingSecret, fromCache bool, _ error) {
	decode := func(val *secretrotation.RotatingSecret) (*secretrotation.RotatingSecret, error) {
		var rs secretrotation.RotatingSecret
		res, err := sm.decodeJSONValue(val.Previous, secretKey)
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, nil
		}
		rs.Previous = *res

		res, err = sm.decodeJSONValue(val.Current, secretKey)
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, nil
		}
		rs.Current = *res

		res, err = sm.decodeJSONValue(val.Pending, secretKey)
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, nil
		}
		rs.Pending = *res

		if err := rs.Validate(); err != nil {
			return nil, err
		}
		return &rs, nil
	}
	res, fromCache, err := loadValue(ctx, secretName, sm.loadSecretVersionStage, decode, sm.cache, sm.lock, secretName)
	if err != nil {
		return nil, fromCache, fmt.Errorf("for secretName=%q, secretKey=%q, %w", secretName, secretKey, err)
	}
	return res, fromCache, nil
}

func (sm *impl) LoadRotatingSecretWhenPlainText(ctx context.Context, secretName string) (_ *secretrotation.RotatingSecret, fromCache bool, _ error) {
	decode := func(val *secretrotation.RotatingSecret) (*secretrotation.RotatingSecret, error) {
		if err := val.Validate(); err != nil {
			return nil, err
		}
		return val, nil
	}
	res, fromCache, err := loadValue(ctx, secretName, sm.loadSecretVersionStage, decode, sm.cache, sm.lock, secretName)
	if err != nil {
		return nil, fromCache, fmt.Errorf("for secretName=%q, %w", secretName, err)
	}
	return res, fromCache, nil
}

func (sm *impl) decodeJSONValue(val secretrotation.Secret, secretKey string) (*secretrotation.Secret, error) {
	var m map[string]secretrotation.Secret
	err := json.Unmarshal([]byte(val), &m)
	if err != nil {
		return nil, err
	}
	res, ok := m[secretKey]
	if !ok {
		return nil, nil
	}
	return &res, nil
}

func (sm *impl) loadSecretSimpleValue(ctx context.Context, secretName string) (*secretrotation.Secret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := sm.svcSecretManager.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	res := secretrotation.Secret(*result.SecretString)
	return &res, nil
}

func (sm *impl) loadSecretVersionStage(ctx context.Context, secretName string) (*secretrotation.RotatingSecret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	loadWithStage := func(stage versionstage.VersionStage) (secretrotation.Secret, error) {
		result, err := sm.svcSecretManager.GetSecretValueWithContext(ctx, input.SetVersionStage(stage.String()))
		if err != nil {
			return "", err
		}
		return secretrotation.Secret(*result.SecretString), nil
	}

	var res secretrotation.RotatingSecret
	var err error
	res.Previous, err = loadWithStage(versionstage.Previous)
	if err != nil {
		return nil, err
	}
	res.Current, err = loadWithStage(versionstage.Current)
	if err != nil {
		return nil, err
	}
	res.Pending, err = loadWithStage(versionstage.Previous)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func loadValue[T *secretrotation.Secret | *secretrotation.RotatingSecret](
	ctx context.Context,
	secretName string,
	loadSecretValue func(_ context.Context, secretName string) (T, error),
	decodeValue func(T) (T, error),
	cache Cache,
	cacheLock lock.Locker,
	cacheKey interface{}, //TODO create ID for the case of reading multiple regions/accounts having the same secretName
) (_ T, fromCache bool, _ error) {
	//using the generic here is not bringing much, this is an experiment

	cacheGet := func() (_ T, ok bool, _ error) {
		if cache == nil {
			return nil, false, nil
		}
		v, ok := cache.Get(cacheKey)
		if !ok {
			return nil, false, nil
		}
		val, err := decodeValue(v.(T))
		if err != nil {
			return nil, true, err
		}
		return val, true, nil
	}

	cacheAdd := func(val T) {
		if cache == nil {
			return
		}
		cache.Add(cacheKey, val)
	}

	val, ok, err := cacheGet()
	if err != nil {
		return nil, true, err
	}
	if ok {
		return val, true, nil
	}

	//prevent un-necessary calls to the secret manager api by locking
	if err := cacheLock.LockWithContext(ctx); err != nil {
		return nil, false, err
	}
	defer cacheLock.Unlock()

	val, ok, err = cacheGet()
	if err != nil {
		return nil, true, err
	}
	if ok {
		return val, true, nil
	}

	v, err := loadSecretValue(ctx, secretName)
	if err != nil {
		return nil, false, err
	}

	cacheAdd(v)

	val, err = decodeValue(v)
	if err != nil {
		return nil, false, err
	}

	return val, false, nil
}
