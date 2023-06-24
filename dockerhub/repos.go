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
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
	Filter   string   `json:"filter"`
}

type FilterStrateg string

const (
	FilterStrategyNone     FilterStrateg = "none"
	FilterStrategyPrefix   FilterStrateg = "prefix"
	FilterStrategySuffix   FilterStrateg = "suffix"
	FilterStrategyContains FilterStrateg = "contains"
	FilterStrategyEquals   FilterStrateg = "equals"
	FilterStrategyAll      FilterStrateg = "all"
)

func filter(filter string) (FilterStrateg, error) {
	if filter == "" {
		return FilterStrategyAll, nil
	}
	if strings.Contains(filter, "*") {
		if filter[0] == '*' && filter[len(filter)-1] == '*' {
			return FilterStrategyContains, nil
		}
		if filter[0] == '*' {
			if strings.LastIndex(filter, "*") == 0 {
				return FilterStrategySuffix, nil
			}
			return FilterStrategyNone, fmt.Errorf("your filter must has one char '*' , example *ccc ")
		}
		if filter[len(filter)-1] == '*' {
			if strings.LastIndex(filter, "*") == len(filter)-1 {
				return FilterStrategyPrefix, nil
			}
			return FilterStrategyNone, fmt.Errorf("your filter must has one char '*' , example ccc* ")
		}
		return FilterStrategyNone, fmt.Errorf("not spport char '*' in filter middle")
	} else {
		return FilterStrategyEquals, nil
	}
}

var specialRepos = []string{"kubernetes", "kubernetes-crio", "kubernetes-docker"}

func (r *RepoInfo) GetVersions() []string {
	type TagList struct {
		Next    string `json:"next"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}
	fetchURL := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/labring/%s/tags?page_size=100", r.Name)
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
				if stringInSlice(r.Name, specialRepos) {
					lower := strings.HasPrefix(tag.Name, "v1.19")
					power := strings.HasPrefix(tag.Name, "v1.2")
					if !lower && !power {
						continue
					}
				}
				s, _ := filter(r.Filter)
				switch s {
				case FilterStrategyAll:
					tagSet = tagSet.Insert(tag.Name)
				case FilterStrategyPrefix:
					if strings.HasPrefix(tag.Name, strings.TrimRight(r.Filter, "*")) {
						tagSet = tagSet.Insert(tag.Name)
					}
				case FilterStrategySuffix:
					if strings.HasSuffix(tag.Name, strings.TrimLeft(r.Filter, "*")) {
						tagSet = tagSet.Insert(tag.Name)
					}
				case FilterStrategyContains:
					if strings.Contains(tag.Name, strings.Trim(r.Filter, "*")) {
						tagSet = tagSet.Insert(tag.Name)
					}
				case FilterStrategyEquals:
					if tag.Name == r.Filter {
						tagSet = tagSet.Insert(tag.Name)
					}
				case FilterStrategyNone:
					logger.Error("filter error: %s", err.Error())
				}

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

	fetchURL := "https://hub.docker.com/v2/repositories/labring?page_size=10"

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
				} else if strings.HasPrefix(repo.Name, "sealos") {
					//TODO will add labring/sealos-patch
					if strings.HasPrefix(repo.Name, "sealos-cloud") || repo.Name == "sealos" {
						versions[repo.Name] = []RepoInfo{
							{Name: repo.Name, Filter: "v*"},
						}
					}
					//logger.Warn("sealos container image repo is deprecated, please use sealos cloud repo")
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
