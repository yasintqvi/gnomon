package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// owner/repo are Gnomon's own canonical, public GitHub repository — the same one README.md
// already points Humans at for releases and issues. This is a fixed property of the project
// itself, not a per-build or per-installation setting, so it is a plain constant rather than
// something Options exposes for override.
const (
	owner = "yasintqvi"
	repo  = "gnomon"
)

// Asset is one downloadable file attached to a GitHub Release.
type Asset struct {
	Name string
	URL  string // the asset's browser_download_url
}

// Release is the subset of a GitHub Release this package needs. Draft and Prerelease are carried
// through rather than filtered out by Fetcher itself, so the "ignore prerelease/draft" rule (Run,
// needsRelease) is ordinary, directly testable logic in this package rather than something a
// Fetcher implementation is trusted to have already enforced.
type Release struct {
	TagName    string
	Draft      bool
	Prerelease bool
	Assets     []Asset
}

// Fetcher is the one seam between this package and the network — every live GitHub/HTTP call
// goes through it, so tests supply a fake instead of ever reaching a real endpoint. Kept to
// exactly the two operations Run actually needs.
type Fetcher interface {
	// LatestRelease returns the most recent release GitHub reports for owner/repo, by whatever
	// its own API considers "latest" — callers decide separately whether it is actually usable
	// (see Release.Draft/Prerelease).
	LatestRelease(owner, repo string) (Release, error)
	// Download fetches the raw bytes at url — a release asset's browser_download_url.
	Download(url string) ([]byte, error)
}

// githubFetcher is Fetcher's real implementation: the GitHub REST API's own "get the latest
// release" endpoint, which already excludes draft and prerelease releases by definition — Run's
// own Draft/Prerelease check (see needsRelease) is defense in depth on top of that, never the
// only thing enforcing it.
type githubFetcher struct {
	client *http.Client
}

// NewGitHubFetcher returns the real Fetcher used outside tests.
func NewGitHubFetcher() Fetcher {
	return &githubFetcher{client: &http.Client{Timeout: 30 * time.Second}}
}

type releaseResponse struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (f *githubFetcher) LatestRelease(owner, repo string) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	// The GitHub API rejects requests with no User-Agent header; it is not otherwise significant.
	req.Header.Set("User-Agent", "gnomon-update")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("checking for the latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Release{}, fmt.Errorf("checking for the latest release: GitHub returned %s: %s", resp.Status, bodyPreview(body))
	}

	var parsed releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return Release{}, fmt.Errorf("reading the latest release: %w", err)
	}

	rel := Release{TagName: parsed.TagName, Draft: parsed.Draft, Prerelease: parsed.Prerelease}
	for _, a := range parsed.Assets {
		rel.Assets = append(rel.Assets, Asset{Name: a.Name, URL: a.BrowserDownloadURL})
	}
	return rel, nil
}

func (f *githubFetcher) Download(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gnomon-update")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("downloading %s: GitHub returned %s: %s", url, resp.Status, bodyPreview(body))
	}
	return io.ReadAll(resp.Body)
}

func bodyPreview(body []byte) string {
	if len(body) == 0 {
		return "(empty response)"
	}
	return string(body)
}
