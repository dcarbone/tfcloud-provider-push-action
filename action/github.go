package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
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
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	)

	ghc := github.NewClient(oauth2.NewClient(ctx, ts))

	return ghc, nil
}

type ShasumFileEntry struct {
	Shasum   string
	Filename string
	Version  string
	OS       string
	Arch     string
}

func (fe ShasumFileEntry) MarshalZerologObject(ev *zerolog.Event) {
	ev.Str("shasum", fe.Shasum)
	ev.Str("filename", fe.Filename)
	ev.Str("version", fe.Version)
	ev.Str("os", fe.OS)
	ev.Str("arch", fe.Arch)
}

type ShasumFile struct {
	Filename string
	Bytes    []byte
	Entries  []ShasumFileEntry
}

func (sf ShasumFile) entryByFilename(fname string) (ShasumFileEntry, bool) {
	for _, fe := range sf.Entries {
		if fe.Filename == fname {
			return fe, true
		}
	}
	return ShasumFileEntry{}, false
}

type ShasumSigFile struct {
	Filename string
	Bytes    []byte
}

type ProviderArtifact struct {
	ShasumFileEntry ShasumFileEntry
	Asset           *github.ReleaseAsset
}

type GithubReleaseContext struct {
	Shasum            ShasumFile
	ShasumSig         ShasumSigFile
	ProviderArtifacts []ProviderArtifact
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

func parseShasumFile(ctx context.Context, _ zerolog.Logger, ghc *github.Client, cfg *Config, asset *github.ReleaseAsset) (ShasumFile, error) {
	ctx, cancel := cfg.ghRequestContext(ctx)
	defer cancel()
	rdr, _, err := ghc.Repositories.DownloadReleaseAsset(ctx, cfg.GithubRepositoryOwner, cfg.githubRepository(), *asset.ID, cleanhttp.DefaultClient())
	if rdr != nil {
		defer drainReader(rdr)
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
		line := scanner.Bytes()

		if bytes.HasSuffix(line, []byte(".zip")) {
			entry, err := shasumFileEntryFromLine(line)
			if err != nil {
				return ShasumFile{}, err
			}
			sumFile.Entries = append(sumFile.Entries, entry)
		}

		sumFile.Bytes = append(sumFile.Bytes, append(line, '\n')...)
	}

	return sumFile, nil
}

func fetchShasumSigFile(ctx context.Context, _ zerolog.Logger, ghc *github.Client, cfg *Config, asset *github.ReleaseAsset) (ShasumSigFile, error) {
	ctx, cancel := cfg.ghRequestContext(ctx)
	defer cancel()
	rdr, _, err := ghc.Repositories.DownloadReleaseAsset(ctx, cfg.GithubRepositoryOwner, cfg.githubRepository(), *asset.ID, cleanhttp.DefaultClient())
	if rdr != nil {
		defer drainReader(rdr)
	}
	if err != nil {
		return ShasumSigFile{}, err
	}

	sigFile := ShasumSigFile{
		Filename: *asset.Name,
	}

	if sigFile.Bytes, err = ioutil.ReadAll(rdr); err != nil {
		return ShasumSigFile{}, fmt.Errorf("error reading body bytes: %w", err)
	}

	return sigFile, nil
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
			if sigFile, err := fetchShasumSigFile(ctx, log, ghc, cfg, asset); err != nil {
				return GithubReleaseContext{}, err
			} else {
				rc.ShasumSig = sigFile
			}
		} else if strings.HasPrefix(*asset.Name, sourceCodeArtifactName) {
			// skip these
			continue
		} else if strings.HasSuffix(*asset.Name, zipSuffix) {
			log.Info().Msg("Found binary asset")
			binaryArtifacts = append(binaryArtifacts, asset)
		}
	}

	if l := len(binaryArtifacts); l == 0 {
		return GithubReleaseContext{}, errors.New("zero binary artifacts found in release")
	} else {
		log.Info().Msgf("Found %d binary artifacts", l)
	}

	rc.ProviderArtifacts = make([]ProviderArtifact, 0)

	for _, ba := range binaryArtifacts {
		log := log.With().Str("provider-artifact", *ba.Name).Logger()
		if fe, ok := rc.Shasum.entryByFilename(*ba.Name); ok {
			log.Debug().Object("entry", fe).Msg("Found shasum entry")
			rc.ProviderArtifacts = append(rc.ProviderArtifacts, ProviderArtifact{
				ShasumFileEntry: fe,
				Asset:           ba,
			})
		}
	}

	if len(rc.ProviderArtifacts) != len(binaryArtifacts) {
		return GithubReleaseContext{},
			fmt.Errorf("count mismatch: binaryArtifacts=%d; ProviderArtifacts=%d", len(rc.ProviderArtifacts), len(binaryArtifacts))
	}

	return rc, nil
}
