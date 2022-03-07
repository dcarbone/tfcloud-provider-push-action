package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v43/github"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/rs/zerolog"
	"regexp"
	"strings"
)

const (
	shasumSuffix           = "SHA256SUMS"
	shasumSigSuffix        = "SHA256SUMS.sig"
	zipSuffix              = ".zip"
	sourceCodeArtifactName = "Source Code"
)

var (
	ParseShasumLineRe = regexp.MustCompile("([^\\s]+)\\s+(.+_([0-9]+\\.[0-9]+\\.[0-9]+)_([^_]+)_([^.]+)\\.zip)$")
)

func NewGithubClient(cfg *Config) (*github.Client, error) {
	ghc := github.NewClient(cleanhttp.DefaultPooledClient())

	return ghc, nil
}

type ShasumFileEntry struct {
	Shasum   string
	Filename string
	Version  string
	OS       string
	Arch     string
}

func shasumFileEntryFromLine(line []byte) (ShasumFileEntry, error) {

	matches := ParseShasumLineRe.FindSubmatch(line)
	if l := len(matches); l != 6 {
		return ShasumFileEntry{}, fmt.Errorf("expected 6 groups in re match, saw %d: line=%q; matches=%v", l, string(line), matches)
	}

	entry := ShasumFileEntry{
		Shasum:   string(matches[1]),
		Filename: string(matches[2]),
		Version:  string(matches[3]),
		OS:       string(matches[4]),
		Arch:     string(matches[5]),
	}

	return entry, nil
}

type ShasumFile struct {
	Filename string
	Bytes    []byte
	Entries  []ShasumFileEntry
}

type ShasumSigFile struct {
	Filename string
}

type ProviderBinaryArtifact struct {
	Shasum   string
	Filename string
	Asset    *github.ReleaseAsset
}

type GithubReleaseContext struct {
	Shasum       ShasumFile
	ShasumSig    ShasumSigFile
	BinaryAssets []ProviderBinaryArtifact
}

func parseShasumFile(ctx context.Context, log zerolog.Logger, ghc *github.Client, cfg *Config, asset *github.ReleaseAsset) (ShasumFile, error) {
	ctx, cancel := cfg.ghRequestContext(ctx)
	defer cancel()
	rdr, _, err := ghc.Repositories.DownloadReleaseAsset(ctx, cfg.GithubRepositoryOwner, cfg.githubRepository(), *asset.ID, cleanhttp.DefaultClient())
	if rdr != nil {
		defer func() {
			_ = rdr.Close()
		}()
	}
	if err != nil {
		return ShasumFile{}, fmt.Errorf("error downloading shasum file asset: %w", err)
	}

	sumFile := ShasumFile{
		Filename: *asset.Name,
		Bytes:    make([]byte, 0),
		Entries:  make([]ShasumFileEntry, 0),
	}

	scanner := bufio.NewScanner(rdr)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		b := scanner.Bytes()
		log.Debug().Msgf("Parsing sumfile line: %q", string(b))
		entry, err := shasumFileEntryFromLine(b)
		if err != nil {
			return ShasumFile{}, err
		}
		sumFile.Bytes = append(sumFile.Bytes, append(b, '\n')...)
		sumFile.Entries = append(sumFile.Entries, entry)
	}

	return sumFile, nil
}

func getReleaseContext(ctx context.Context, log zerolog.Logger, ghc *github.Client, cfg *Config) (GithubReleaseContext, error) {
	rc := GithubReleaseContext{}

	releaseMeta, _, err := ghc.Repositories.GetReleaseByTag(ctx, cfg.GithubRepositoryOwner, cfg.githubRepository(), cfg.GithubRefName)
	if err != nil {
		err = fmt.Errorf("error fetching release metadata from github: %w", err)
		return GithubReleaseContext{}, err
	}

	binaryArtifacts := make([]*github.ReleaseAsset, 0)

	for _, asset := range releaseMeta.Assets {
		if asset.Name == nil || asset.URL == nil || asset.ID == nil {
			log.Debug().
				Interface("id", asset.ID).
				Interface("name", asset.Name).
				Interface("url", asset.URL).
				Msg("Skipping asset as at least one of name, url, and id are empty")
			continue
		}
		log := log.With().Str("asset-name", *asset.Name).Logger()
		if strings.HasSuffix(*asset.Name, shasumSuffix) {
			log.Info().Msg("Found shasum file")
			if sumFile, err := parseShasumFile(ctx, log, ghc, cfg, asset); err != nil {
				return GithubReleaseContext{}, err
			} else {
				rc.Shasum = sumFile
			}
		} else if strings.HasSuffix(*asset.Name, shasumSigSuffix) {
			log.Info().Msg("Found shasum sig file")
			ga.ShasumSig = asset
		} else if strings.HasPrefix(*asset.Name, sourceCodeArtifactName) {
			// skip these
			continue
		} else if strings.HasSuffix(*asset.Name, zipSuffix) {
			log.Info().Msg("Found binary asset")
			binaryArtifacts = append(binaryArtifacts, asset)
		}
	}
}

func parseReleaseAssets(ctx context.Context, log zerolog.Logger, rel *github.RepositoryRelease) (GoreleaserAssets, error) {

	if len(binaryArtifacts) == 0 {
		return ga, errors.New("zero binary artifacts found in release")
	}

	parsedSums, err := fetchAndParseShasumFile(ctx, ga.Shasum)

	log.Info().Msgf("Found %d binary artifacts", len(binaryArtifacts))

	for _, ba := range binaryArtifacts {

	}
}
