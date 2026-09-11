package pkgmgr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Registry is the small network boundary used by the installer. Its client is
// injectable so installs can be tested against a deterministic local registry.
type Registry struct {
	BaseURL string
	Client  *http.Client
}

func (r Registry) endpoint(name string) (string, error) {
	base := strings.TrimRight(r.BaseURL, "/")
	if base == "" {
		base = "https://registry.npmjs.org"
	}
	u, err := url.Parse(base + "/" + url.PathEscape(name))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid registry URL %q", r.BaseURL)
	}
	return u.String(), nil
}

// FetchPackage returns all published versions for one package name.
func (r Registry) FetchPackage(name string) (map[string]PackageManifest, error) {
	endpoint, err := r.endpoint(name)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Name     string                     `json:"name"`
		Versions map[string]PackageManifest `json:"versions"`
	}
	if err := r.getJSON(endpoint, &payload); err != nil {
		return nil, fmt.Errorf("fetch package %q metadata: %w", name, err)
	}
	if len(payload.Versions) == 0 {
		return nil, fmt.Errorf("registry returned no versions for %q", name)
	}
	return payload.Versions, nil
}

func (r Registry) getJSON(endpoint string, target any) error {
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.npm.install-v1+json, application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("registry returned HTTP %s", response.Status)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<20)).Decode(target); err != nil {
		return err
	}
	return nil
}

func (r Registry) download(endpoint string) ([]byte, error) {
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned HTTP %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 512<<20))
	if err != nil {
		return nil, err
	}
	return data, nil
}
