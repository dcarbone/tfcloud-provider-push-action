package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/rs/zerolog"
)

const (
	GithubRequestTTLDefault = "5s"

	TFAddressDefault           = "https://app.terraform.io"
	TFRegistryNameDefault      = "private"
	TFProviderPlatformsDefault = "6.0"
	TFRequestTTLDefault        = "5s"
	TFUploadTTLDefault         = "5m"

	EnvGithubToken           = "GITHUB_TOKEN"
	EnvGithubRefName         = "GITHUB_REF_NAME"
	EnvGithubRepository      = "GITHUB_REPOSITORY"
	EnvGithubRepositoryOwner = "GITHUB_REPOSITORY_OWNER"
	EnvGithubRequestTTL      = "GITHUB_REQUEST_TTL"

	EnvTFAddress           = "TF_ADDRESS"
	EnvTFToken             = "TF_TOKEN"
	EnvTFRegistryName      = "TF_REGISTRY_NAME"
	EnvTFOrganizationName  = "TF_ORGANIZATION_NAME"
	EnvTFNamespace         = "TF_NAMESPACE"
	EnvTFProviderName      = "TF_PROVIDER_NAME"
	EnvTFProviderPlatforms = "TF_PROVIDER_PLATFORMS"
	EnvTFRequestTTL        = "TF_REQUEST_TTL"
	EnvTFUploadTTL         = "TF_UPLOAD_TTL"
)

type Config struct {
	GithubToken           string
	GithubRefName         string
	GithubRepository      string
	GithubRepositoryOwner string
	GithubRequestTTL      string

	TFAddress           string
	TFToken             string
	TFGPGKeyID          string
	TFRegistryName      string
	TFOrganizationName  string
	TFNamespace         string
	TFProviderName      string
	TFProviderPlatforms string
	TFRequestTTL        string
	TFUploadTTL         string

	githubRequestTTL time.Duration

	tfProviderPlatforms []string
	tfRequestTTL        time.Duration
	tfUploadTTL         time.Duration
}

func (c Config) providerVersion() string {
	return strings.TrimSuffix(c.GithubRefName, "v")
}

func (c Config) githubRepository() string {
	return strings.Replace(c.GithubRepository, fmt.Sprintf("%s/", c.GithubRepositoryOwner), "", 1)
}

func (c Config) ghRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.githubRequestTTL)
}

func (c Config) tfRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.tfRequestTTL)
}

func (c Config) tfUploadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.tfUploadTTL)
}

func defaultConfig() *Config {
	c := Config{
		GithubRequestTTL:    GithubRequestTTLDefault,
		TFAddress:           TFAddressDefault,
		TFRegistryName:      TFRegistryNameDefault,
		TFProviderPlatforms: TFProviderPlatformsDefault,
		TFRequestTTL:        TFRequestTTLDefault,
		TFUploadTTL:         TFUploadTTLDefault,
	}

	return &c
}

func main() {
	var (
		err error

		cfg = defaultConfig()
	)

	envs := map[string]*string{
		EnvGithubToken:           &cfg.GithubToken,
		EnvGithubRefName:         &cfg.GithubRefName,
		EnvGithubRepository:      &cfg.GithubRepository,
		EnvGithubRepositoryOwner: &cfg.GithubRepositoryOwner,
		EnvGithubRequestTTL:      &cfg.GithubRequestTTL,

		EnvTFAddress:           &cfg.TFAddress,
		EnvTFToken:             &cfg.TFToken,
		EnvTFRegistryName:      &cfg.TFRegistryName,
		EnvTFOrganizationName:  &cfg.TFOrganizationName,
		EnvTFNamespace:         &cfg.TFNamespace,
		EnvTFProviderName:      &cfg.TFProviderName,
		EnvTFProviderPlatforms: &cfg.TFProviderPlatforms,
		EnvTFRequestTTL:        &cfg.TFRequestTTL,
		EnvTFUploadTTL:         &cfg.TFUploadTTL,
	}

	for envName, vPtr := range envs {
		v, ok := os.LookupEnv(envName)
		v = strings.TrimSpace(v)
		if !ok || v == "" {
			if *vPtr == "" {
				panic(fmt.Sprintf("Environment variable %q is required", envName))
			}
			//os.Exit(0)
		} else {
			*vPtr = v
		}
	}

	log := newLogger(cfg)

	// laziness!
	if cfg.githubRequestTTL, err = time.ParseDuration(cfg.GithubRequestTTL); err != nil {
		log.Error().Msgf("Environment variable %q value %q is not parseable as time.Duration: %w", EnvGithubRequestTTL, cfg.GithubRequestTTL, err)
		os.Exit(1)
	}
	if cfg.tfRequestTTL, err = time.ParseDuration(cfg.TFRequestTTL); err != nil {
		log.Error().Msgf("Environment variable %q value %q is not parseable as time.Duration: %w", EnvTFRequestTTL, cfg.TFRequestTTL, err)
		os.Exit(1)
	}
	if cfg.tfUploadTTL, err = time.ParseDuration(cfg.TFUploadTTL); err != nil {
		log.Error().Msgf("Environment variable %q value %q is not parseable as time.Duration: %w", EnvTFUploadTTL, cfg.TFUploadTTL, err)
		os.Exit(1)
	}

	cfg.tfProviderPlatforms = strings.Split(cfg.TFProviderPlatforms, ",")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	errChan := make(chan error, 1)
	exitCode := 0

	go run(ctx, errChan, log, cfg)

	select {
	case err := <-errChan:
		if err != nil {
			fmt.Println(err.Error())
			exitCode = 1
		}
	case <-ctx.Done():
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println(err.Error())
			exitCode = 1
		}
	}

	cancel()

	os.Exit(exitCode)
}

func run(ctx context.Context, done chan<- error, log zerolog.Logger, cfg *Config) {
	var (
		tfc *TFClient
		ghc *github.Client
		err error
	)

	defer func() {
		done <- err
	}()

	if tfc, err = NewTFClient(cfg); err != nil {
		err = fmt.Errorf("error constructing TFClient: %w", err)
		return
	}

	if ghc, err = NewGithubClient(cfg); err != nil {
		err = fmt.Errorf("error constructing github.Client: %w", err)
		return
	}

	rc, err := getReleaseContext(ctx, log, ghc, cfg)
	if err != nil {
		err = fmt.Errorf("error parsing release context: %w", err)
		return
	}

	log.Debug().Msg("Release context parsed")

	pvc := NewTFCreateProviderVersionRequest(cfg.providerVersion(), cfg.TFGPGKeyID, cfg.tfProviderPlatforms)
	ctx, cancel := cfg.tfRequestContext(ctx)
	defer cancel()
	pv, err := tfc.ProviderClient().CreateProviderVersion(ctx, cfg.TFOrganizationName, cfg.TFRegistryName, cfg.TFNamespace, cfg.TFProviderName, pvc)
	if err != nil {
		err = fmt.Errorf("error creating new provider version: %w", err)
		return
	}

	log.Info().Msg("Provider version created")

}
