/*
Copyright 2023 cuisongliu@qq.com.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dockerhub

import (
	"encoding/json"
	"fmt"
	"github.com/cuisongliu/logger"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

type RepoInfo struct {
	Name         string   `json:"name"`
	Versions     []string `json:"versions"`
	FixedVersion bool     `json:"fixed_version"`
}

func (r *RepoInfo) GetVersions() []string {
	type TagList struct {
		Next    string `json:"next"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	fetchURL := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/labring/%s/tags", r.Name)
	tagSet := sets.Set[string]{}
	if err := Retry(func() error {
		for fetchURL != "" {
			logger.Debug("fetch dockerhub url: %s", fetchURL)
			data, err := Request(fetchURL, "GET", []byte(""), 0)
			if err != nil {
				return err
			}
			var tags TagList
			if err = json.Unmarshal(data, &tags); err != nil {
				return err
			}
			for _, tag := range tags.Results {
				if strings.HasSuffix(tag.Name, "-amd64") {
					continue
				}
				if strings.HasSuffix(tag.Name, "-arm64") {
					continue
				}
				tagSet = tagSet.Insert(tag.Name)
			}
			fetchURL = tags.Next
		}
		r.Versions = sets.List(tagSet)
		return nil
	}); err != nil {
		logger.Error("fetch dockerhub url: %s error: %s", fetchURL, err.Error())
		r.Versions = []string{}
		return nil
	}
	return r.Versions
}

func fetchDockerHubAllRepo() (map[string][]RepoInfo, error) {
	type Repo struct {
		Name string `json:"name"`
	}

	type Repositories struct {
		Results []Repo `json:"results"`
		Next    string `json:"next"`
	}

	fetchURL := "https://hub.docker.com/v2/repositories/labring/"
	specialRepos := []string{"kubernetes", "kubernetes-crio", "kubernetes-docker"}

	versions := make(map[string][]RepoInfo)
	if err := Retry(func() error {
		index := 0
		for fetchURL != "" {
			logger.Debug("fetch dockerhub url: %s", fetchURL)
			data, err := Request(fetchURL, "GET", []byte(""), 0)
			if err != nil {
				return err
			}
			var repositories Repositories
			if err = json.Unmarshal(data, &repositories); err != nil {
				return err
			}
			newRepos := make([]RepoInfo, 0)
			for _, repo := range repositories.Results {
				if stringInSlice(repo.Name, specialRepos) {
					versions[repo.Name] = []RepoInfo{
						{Name: repo.Name},
					}
				} else if strings.HasPrefix(repo.Name, "sealos-cloud") {
					if repo.Name == "sealos-patch" || strings.HasPrefix(repo.Name, "sealos-cloud") || repo.Name == "sealos" {
						versions[repo.Name] = []RepoInfo{
							{Name: repo.Name},
						}
					}
					logger.Warn("sealos container image repo is deprecated, please use sealos cloud repo")
				} else {
					newRepos = append(newRepos, RepoInfo{Name: repo.Name})
				}
			}
			versions[fmt.Sprintf("image-%d", index)] = newRepos
			index++
			fetchURL = repositories.Next
		}
		return nil
	}); err != nil {
		logger.Error("get dockerhub repo error: %s", err.Error())
		return nil, err
	}
	return versions, nil
}
