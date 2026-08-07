package main

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

type itemT struct {
	env     string
	profile string
}

// ProfileT -
type ProfileT struct {
	Name        string
	Config      *aws.Config
	ELBv2Client *elasticloadbalancingv2.Client
	envs        *[]EnvT
}

// EnvT -
type EnvT struct {
	Name string
	LBs  *[]LbT
}

// LbT -
type LbT struct {
	ARN    string
	TgARNs *[]string
	mu     *sync.Mutex
}

// SortLBs -
func (e *EnvT) SortLBs() []LbT {
	if lbs := slices.Clone(*(*e).LBs); len(lbs) > 0 {
		slices.SortFunc(lbs, func(a, b LbT) int {
			return strings.Compare(a.ARN, b.ARN)
		})
		return lbs
	}
	return nil
}

func buildProfiles(prntCtx context.Context, args []string) ([]itemT, []ProfileT, error) {
	items := make([]itemT, 0, len(args))
	exists := make(map[string]bool, cap(items))
	prfs := make([]ProfileT, 0)
	m := make(map[string]int, cap(prfs))
	for i := range args {
		env, profile := envProfile(args[i])
		if env == "" {
			return nil, nil, errors.New("\"" + args[i] + "\": environment name cannot be empty")
		}
		if !exists[env+string(epSep)+profile] {
			exists[env+string(epSep)+profile] = true
			items = append(items, itemT{env, profile})
			if j, ok := m[profile]; ok {
				*prfs[j].envs = append(*prfs[j].envs, EnvT{Name: env})
				continue
			}
			opts := []func(*config.LoadOptions) error{config.WithSharedConfigProfile(profile)}
			if profile == "" {
				opts = nil
			}
			ctx, cancel := context.WithTimeout(prntCtx, timeoutSec*time.Second)
			cfg, err := loadConfig(ctx, opts...)
			cancel()
			if err != nil {
				return nil, nil, err
			}
			m[profile] = len(prfs)
			prfs = append(prfs, ProfileT{profile, &cfg, newElbClient(cfg), new([]EnvT{{Name: env}})})
		}
	}
	if len(items) < 1 {
		return nil, nil, nil
	}
	return items, prfs, nil
}
