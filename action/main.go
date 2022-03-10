package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog"
)

const (
	GithubRequestTTLDefault  = "5s"
	GithubDownloadTTLDefault = "5m"

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
	EnvGithubDownloadTTL     = "GITHUB_DOWNLOAD_TTL"

	EnvTFAddress           = "TF_ADDRESS"
	EnvTFToken             = "TF_TOKEN"
	EnvTFGPGKeyID          = "TF_GPG_KEY_ID"
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
	GithubDownloadTTL     string

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

	githubRequestTTL  time.Duration
	githubDownloadTTL time.Duration

	tfProviderPlatforms []string
	tfRequestTTL        time.Duration
	tfUploadTTL         time.Duration
}

func (c Config) providerVersion() string {
	return strings.TrimPrefix(c.GithubRefName, "v")
}

func (c Config) githubRepository() string {
	return strings.Replace(c.GithubRepository, fmt.Sprintf("%s/", c.GithubRepositoryOwner), "", 1)
}

func (c Config) ghRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.githubRequestTTL)
}

func (c Config) ghDownloadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.githubDownloadTTL)
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
		GithubDownloadTTL:   GithubDownloadTTLDefault,
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
		EnvGithubDownloadTTL:     &cfg.GithubDownloadTTL,

		EnvTFAddress:           &cfg.TFAddress,
		EnvTFToken:             &cfg.TFToken,
		EnvTFGPGKeyID:          &cfg.TFGPGKeyID,
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
				err = multierror.Append(err, fmt.Errorf("missing required environment variable: %q", envName))
			}
		} else {
			*vPtr = v
		}
	}

	if me, ok := err.(*multierror.Error); ok && me.Len() > 0 {
		for _, e := range me.Errors {
			fmt.Println(e.Error())
		}
		os.Exit(1)
	}

	err = nil

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
			log.Error().Err(err).Msg("Error occurred during execution")
			fmt.Println(err.Error())
			exitCode = 1
		}
	case <-ctx.Done():
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error().Err(err).Msg("Execution terminated")
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

	fileData := TFFileUploadRequest{
		File:        bytes.NewBuffer(rc.Shasum.Bytes),
		Destination: pv.Data.Links.ShasumsUpload,
		ContentType: binaryOctetStream,
		Filename:    rc.Shasum.Filename,
	}

	log.Debug().Msgf("Attempting to upload %q file to %q...", rc.Shasum.Filename, pv.Data.Links.ShasumsUpload)
	ctx, cancel = cfg.tfUploadContext(ctx)
	defer cancel()
	if err = tfc.UploadsClient().UploadFile(ctx, fileData); err != nil {
		err = fmt.Errorf("error uploading %s file: %w", rc.Shasum.Filename, err)
		return
	}

	fileData = TFFileUploadRequest{
		File:        bytes.NewBuffer(rc.ShasumSig.Bytes),
		Destination: pv.Data.Links.ShasumsSigUpload,
		ContentType: binaryOctetStream,
		Filename:    rc.ShasumSig.Filename,
	}

	log.Debug().Msgf("Attempting to upload %q file to %q...", rc.ShasumSig.Filename, pv.Data.Links.ShasumsSigUpload)
	ctx, cancel = cfg.tfUploadContext(ctx)
	defer cancel()
	if err = tfc.UploadsClient().UploadFile(ctx, fileData); err != nil {
		err = fmt.Errorf("error uploading %q file: %w", rc.ShasumSig.Filename, err)
		return
	}

	log.Info().Msgf("Files %q and %q uploaded successfully", rc.Shasum.Filename, rc.ShasumSig.Filename)
	log.Info().Msgf("Preparing %d binary uploads...", len(rc.ProviderArtifacts))

	// todo: if i were a smart man, i could do this with just the channel.
	wg := new(sync.WaitGroup)
	wg.Add(len(rc.ProviderArtifacts))
	errc := make(chan error, len(rc.ProviderArtifacts))
	defer close(errc)

	for _, pa := range rc.ProviderArtifacts {
		log := log.With().Str("provider-artifact", *pa.Asset.Name).Logger()
		go uploadProviderBinary(ctx, log, tfc, ghc, pa, cfg, wg, errc)
	}

	wg.Wait()

	for uploadErr := range errc {
		if uploadErr != nil {
			log.Error().Err(err).Msg("Error during binary upload")
			err = multierror.Append(err, uploadErr)
		}
	}
}

func uploadProviderBinary(
	ctx context.Context,
	log zerolog.Logger,
	tfc *TFClient,
	ghc *github.Client,
	pa ProviderArtifact,
	cfg *Config,
	wg *sync.WaitGroup,
	errc chan<- error,
) {

	var err error

	// queue up cleanup
	defer func() {
		wg.Done()
		errc <- err
	}()

	log.Info().Msg("Creating provider version platform...")

	pvfc := NewTFCreateProviderVersionPlatformRequest(
		pa.ShasumFileEntry.OS,
		pa.ShasumFileEntry.Arch,
		pa.ShasumFileEntry.Shasum,
		pa.ShasumFileEntry.Filename,
	)
	ctx, cancel := cfg.tfRequestContext(ctx)
	defer cancel()
	pvf, err := tfc.ProviderClient().CreateProviderVersionPlatform(
		ctx,
		cfg.TFOrganizationName,
		cfg.TFRegistryName,
		cfg.TFNamespace,
		cfg.TFProviderName,
		pa.ShasumFileEntry.Version,
		pvfc,
	)
	if err != nil {
		err = fmt.Errorf("error creating provider version platform: %w", err)
		return
	}

	log.Info().Msg("Preparing to upload provider binary...")

	ctx, cancel = cfg.ghDownloadContext(ctx)
	defer cancel()
	rdr, _, err := ghc.Repositories.DownloadReleaseAsset(ctx, cfg.GithubRepositoryOwner, cfg.githubRepository(), pa.Asset.GetID(), cleanhttp.DefaultClient())
	if rdr != nil {
		defer drainReader(rdr)
	}
	if err != nil {
		err = fmt.Errorf("error initiating download of release asset %q: %w", pa.ShasumFileEntry.Filename, err)
		return
	}

	fileData := TFFileUploadRequest{
		File:        rdr,
		Destination: pvf.Data.Links.ProviderBinaryUpload,
		ContentType: binaryOctetStream,
		Filename:    pa.ShasumFileEntry.Filename,
	}
	ctx, cancel = cfg.tfUploadContext(ctx)
	defer cancel()
	if err = tfc.UploadsClient().UploadFile(ctx, fileData); err != nil {
		err = fmt.Errorf("error uploading provider binary %q: %w", pa.ShasumFileEntry.Filename, err)
		return
	}

	log.Info().Msg("Provider binary successfully uploaded!")
}
