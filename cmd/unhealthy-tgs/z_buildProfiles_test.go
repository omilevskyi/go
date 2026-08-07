package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/go-jose/go-jose/v4/testutils/assert"
	"github.com/go-openapi/testify/v2/require"
)

func mockDeps(t *testing.T) {
	t.Helper()

	oldLoadConfig := loadConfig
	loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	oldNewElbClient := newElbClient
	newElbClient = func(cfg aws.Config, _ ...func(*elasticloadbalancingv2.Options)) *elasticloadbalancingv2.Client {
		return &elasticloadbalancingv2.Client{}
	}

	t.Cleanup(func() {
		loadConfig = oldLoadConfig
		newElbClient = oldNewElbClient
	})
}

func TestBuildProfiles_EmptyArgs(t *testing.T) {
	items, profiles, err := buildProfiles(context.Background(), nil)

	require.NoError(t, err)
	require.Nil(t, items)
	require.Nil(t, profiles)
}

func TestBuildProfiles_EmptySlice(t *testing.T) {
	items, profiles, err := buildProfiles(context.Background(), []string{})

	require.NoError(t, err)
	require.Nil(t, items)
	require.Nil(t, profiles)
}

func TestBuildProfiles_SingleEnv(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(context.Background(), []string{"devo1"})

	require.NoError(t, err)

	require.Equal(t, []itemT{{"devo1", ""}}, items)

	require.Len(t, profiles, 1)
	require.Equal(t, "", profiles[0].Name)

	require.NotNil(t, profiles[0].Config)
	require.NotNil(t, profiles[0].ELBv2Client)
}

func TestBuildProfiles_SingleProfile(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(context.Background(), []string{"devo1:non-prod"})

	require.NoError(t, err)

	require.Equal(t, []itemT{{"devo1", "non-prod"}}, items)
	require.Len(t, profiles, 1)
	require.Equal(t, "non-prod", profiles[0].Name)
}

func TestBuildProfiles_DuplicateIgnored(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devo1:non-prod",
			"devo1:non-prod",
			"devo1:non-prod",
		},
	)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Len(t, profiles, 1)
}

func TestBuildProfiles_TrimmedDuplicateIgnored(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devo1:non-prod",
			" devo1 : non-prod ",
		},
	)

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Len(t, profiles, 1)
}

func TestBuildProfiles_SameEnvDifferentProfiles(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devo1:non-prod",
			"devo1:digital-non-production",
		},
	)

	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Len(t, profiles, 2)
}

func TestBuildProfiles_DifferentProfiles(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devo1:non-prod",
			"uatx2:digital-non-production",
			"stg:digital-production",
		},
	)

	require.NoError(t, err)

	require.Len(t, items, 3)
	require.Len(t, profiles, 3)
}

func TestBuildProfiles_EmptyEnv(t *testing.T) {
	mockDeps(t)

	_, _, err := buildProfiles(context.Background(), []string{":non-prod"})

	require.EqualError(t, err, `":non-prod": environment name cannot be empty`)
}

func TestBuildProfiles_TrimmedEmptyEnv(t *testing.T) {
	mockDeps(t)

	_, _, err := buildProfiles(context.Background(), []string{" : non-prod "})

	require.EqualError(t, err, `" : non-prod ": environment name cannot be empty`)
}

func TestBuildProfiles_LoadConfigError(t *testing.T) {
	oldLoadConfig := loadConfig
	defer func() {
		loadConfig = oldLoadConfig
	}()

	loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load failure")
	}

	_, _, err := buildProfiles(context.Background(), []string{"devo1:non-prod"})

	require.EqualError(t, err, "load failure")
}

func TestBuildProfiles_OrderPreserved(t *testing.T) {
	mockDeps(t)

	items, _, err := buildProfiles(
		context.Background(),
		[]string{
			"uatx2:digital-non-production",
			"devo1:non-prod",
			"stg:digital-production",
		},
	)

	require.NoError(t, err)

	require.Equal(
		t,
		[]itemT{
			{"uatx2", "digital-non-production"},
			{"devo1", "non-prod"},
			{"stg", "digital-production"},
		},
		items,
	)
}

func TestBuildProfiles_LoadProfileOnlyOnce(t *testing.T) {
	oldLoadConfig := loadConfig
	defer func() {
		loadConfig = oldLoadConfig
	}()

	var calls atomic.Int32

	loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		calls.Add(1)
		return aws.Config{}, nil
	}

	_, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devl1:non-prod",
			"devo1:non-prod",
			"devp1:non-prod",
			"uatx2:non-prod",
		},
	)

	require.NoError(t, err)
	require.Len(t, profiles, 1)

	if got := calls.Load(); got != 1 {
		t.Fatalf("loadConfig calls=%d want 1", got)
	}
}

func TestBuildProfiles_DefaultProfile(t *testing.T) {
	oldLoadConfig := loadConfig
	defer func() {
		loadConfig = oldLoadConfig
	}()

	var calls int32

	loadConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		atomic.AddInt32(&calls, 1)

		if len(optFns) != 0 {
			t.Fatalf("expected no LoadOptions for default profile")
		}

		return aws.Config{}, nil
	}

	_, profiles, err := buildProfiles(context.Background(), []string{"devo1", "uatx2"})
	require.NoError(t, err)
	require.Len(t, profiles, 1)

	if profiles[0].Name != "" {
		t.Fatalf("profile=%q want empty", profiles[0].Name)
	}

	if calls != 1 {
		t.Fatalf("calls=%d want 1", calls)
	}
}

func TestBuildProfiles_AppendEnvToExistingProfile(t *testing.T) {
	mockDeps(t)

	items, profiles, err := buildProfiles(
		context.Background(),
		[]string{
			"devo1:non-prod",
			"uatx2:non-prod",
			"stress:non-prod",
		},
	)

	require.NoError(t, err)

	require.Len(t, items, 3)
	require.Len(t, profiles, 1)

	require.NotNil(t, profiles[0].envs)
	envs := *profiles[0].envs

	require.Len(t, envs, 3)

	assert.Equal(t, "devo1", envs[0].Name)
	assert.Equal(t, "uatx2", envs[1].Name)
	assert.Equal(t, "stress", envs[2].Name)
}
