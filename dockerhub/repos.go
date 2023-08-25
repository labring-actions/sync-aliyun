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
	"strings"
)

type registryName string

const defaultRegistryName registryName = "docker.io"

type registrySyncConfig struct {
	Images           map[string][]string `yaml:"images,omitempty"`              // Images map images name to slices with the images' tags
	ImagesByTagRegex map[string]string   `yaml:"images-by-tag-regex,omitempty"` // Images map images name to regular expression with the images' tags
	TLSVerify        bool                `yaml:"tls-verify"`                    // TLS verification mode (enabled by default)
}

type SkopeoList map[registryName]registrySyncConfig

var specialRepos = []string{"kubernetes", "kubernetes-crio", "kubernetes-docker"}

const defaultRepo = "labring"

type Repo struct {
	Name string `json:"name"`
}

type Repositories struct {
	Results []Repo `json:"results"`
	Next    string `json:"next"`
}

func (r *Repo) getName() string {
	return fmt.Sprintf("%s/%s", defaultRepo, r.Name)
}

func fetchDockerHubAllRepo() (map[string]SkopeoList, error) {

	fetchURL := "https://hub.docker.com/v2/repositories/labring?page_size=10"

	versions := make(map[string]SkopeoList)
	defaultRepos := make([]string, 0)
	if err := Retry(func() error {
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
			for _, repo := range repositories.Results {
				if stringInSlice(repo.Name, specialRepos) {
					versions[repo.Name] = SkopeoList{
						defaultRegistryName: {
							Images:           nil,
							ImagesByTagRegex: map[string]string{repo.getName(): "^v(1\\.2[0-9]\\.[1-9]?[0-9]?)(\\.)?$"},
							TLSVerify:        false,
						},
					}
				} else if strings.HasPrefix(repo.Name, "sealos") {
					versions[repo.Name] = SkopeoList{
						defaultRegistryName: {
							Images:           map[string][]string{repo.getName(): {"latest"}},
							ImagesByTagRegex: map[string]string{repo.getName(): "^v.*"},
							TLSVerify:        false,
						},
					}
				} else if strings.HasPrefix(repo.Name, "laf") {
					versions[repo.Name] = SkopeoList{
						defaultRegistryName: {
							Images:           map[string][]string{repo.getName(): {"latest"}},
							ImagesByTagRegex: map[string]string{repo.getName(): "^v.*"},
							TLSVerify:        false,
						},
					}
				} else {
					defaultRepos = append(defaultRepos, repo.getName())
				}
			}
			fetchURL = repositories.Next
		}
		return nil
	}); err != nil {
		logger.Error("get dockerhub repo error: %s", err.Error())
		return nil, err
	}
	count := 20
	defaultImages := make([]map[string][]string, count)
	for i, repo := range defaultRepos {
		index := i % count
		if defaultImages[index] == nil {
			defaultImages[index] = make(map[string][]string)
		}
		defaultImages[index][repo] = []string{}
	}
	for i, images := range defaultImages {
		versions[fmt.Sprintf("image-%d", i)] = SkopeoList{
			defaultRegistryName: {
				Images:    images,
				TLSVerify: false,
			},
		}
	}
	return versions, nil
}
