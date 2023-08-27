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
	"fmt"
	"github.com/cuisongliu/logger"
	"testing"
)

func TestFetchDockerHubAllVersion(t *testing.T) {
	logger.Cfg(true, false)
	err := autoRemoveGenerator("../skopeo")
	if err != nil {
		t.Error(err)
		return
	}
	err = autoRemoveGenerator("../.github/workflows")
	if err != nil {
		t.Error(err)
		return
	}
	got, err := fetchDockerHubAllRepo()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("get docker hub all repo success")
	for k, v := range got {
		err = generatorSyncFile("../skopeo", k, v)
		if err != nil {
			t.Errorf("generatorSyncFile %s error %s", k, err.Error())
			continue
		}
		err = generatorWorkflowFile("../.github/workflows", "skopeo", k, nil)
		if err != nil {
			t.Errorf("generatorWorkflowFile %s error %s", k, err.Error())
			continue
		}
	}
}

func TestGroups(t *testing.T) {
	defaultRepos := []string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
		"repo5",
		"repo6",
		"repo7",
		"repo8",
		"repo9",
		"repo10",
	}
	groupSize := 5
	groups := make(map[int][]string)
	for i, repo := range defaultRepos {
		groupIndex := i / groupSize
		groups[groupIndex] = append(groups[groupIndex], repo)
	}
	for groupIndex, images := range groups {
		fmt.Printf("Group %d:\n", groupIndex+1)
		fmt.Printf("Group data %+v:\n", images)
	}

	data := fmt.Sprintf("^v(1\\.%s\\.[1-9]?[0-9]?)(\\.)?$", "20")
	fmt.Println(data)
}
