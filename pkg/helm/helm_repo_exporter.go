// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"k8s.io/klog/v2"
)

const (
	helmExporterAPIRoot = "https://artifacthub.io/api/v1/helm-exporter"
)

var helmRepocacheFile = ""

// ArtifactHubCachedPackageClient provides the various pieces to interact with the ArtifactHubCached API.
type ArtifactHelmCachedPackageClient struct {
	APIRoot   string
	URL       *url.URL
	Client    *http.Client
	UserAgent string
}

func init() {
	if os.Getenv("ARTIFACT_HELM_REPO_HUB_CACHE_FILE") != "" {
		cacheFile = os.Getenv("ARTIFACT_HELM_REPO_HUB_CACHE_FILE")
	}
}

type ArtifactHelmRepoCachedPackagesList []ArtifactHelmPackage

// ArtifactHubRepository is a sub-struct of the Package struct, and represents the repository containing the package.
type ArtifactHelmRepository struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ArtifactHelmPackage struct {
	Name       string
	Version    string
	Repository ArtifactHubRepository
}

// NewArtifactHubCachedPackageClient returns a new client for the unauthenticated paths of the ArtifactHubCached API.
func NewHelmRepoHubCachedPackageClient(version string) (*ArtifactHelmCachedPackageClient, error) {
	apiRoot := helmExporterAPIRoot
	u, err := url.ParseRequestURI(apiRoot)
	if err != nil {
		return nil, err
	}
	client := new(http.Client)
	return &ArtifactHelmCachedPackageClient{
		APIRoot:   apiRoot,
		URL:       u,
		Client:    client,
		UserAgent: fmt.Sprintf("Fairwinds-Nova/%s ", version),
	}, nil
}

// List returns all packages from ArtifactHub
func (ac *ArtifactHelmCachedPackageClient) ListRepo() ([]ArtifactHelmPackage, error) {
	list := ArtifactHelmRepoCachedPackagesList{}
	if cacheFile == "" {
		resp, err := ac.get()
		if err != nil {
			return nil, err
		}
		err = json.NewDecoder(resp.Body).Decode(&list)
		if err != nil {
			return nil, err
		}
	} else {
		cache, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(cache, &list)
		if err != nil {
			return nil, err
		}
	}
	packages := make([]ArtifactHelmPackage, len(list))
	for idx, cachedPackage := range list {
		packages[idx] = ArtifactHelmPackage{
			Name:    cachedPackage.Name,
			Version: cachedPackage.Version,
			Repository: ArtifactHubRepository{
				Name: cachedPackage.Repository.Name,
				URL:  cachedPackage.Repository.URL,
			},
		}
	}

	return packages, nil
}

// get is the basic getter for the artifacthub cached package client
func (ac *ArtifactHelmCachedPackageClient) get() (*http.Response, error) {
	requestURL := *ac.URL
	urlString := requestURL.String()
	r, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	r.URL.RawQuery = q.Encode()
	r.Header.Add("accept", "application/json")
	resp, err := ac.Client.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		klog.V(3).Infof("failed to GET %s with status code: %v", urlString, resp.StatusCode)
		err = fmt.Errorf("error code: %d", resp.StatusCode)
	}
	return resp, err
}
