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

import "github.com/cuisongliu/logger"

func Do(syncDir string) {
	workflowDir := ".github/workflows"
	logger.Cfg(true, false)
	err := autoRemoveGenerator(syncDir)
	if err != nil {
		logger.Error("autoRemoveGenerator sync config error %s", err.Error())
		return
	}
	err = autoRemoveGenerator(workflowDir)
	if err != nil {
		logger.Error("autoRemoveGenerator workflow config error %s", err.Error())
		return
	}
	got, err := fetchDockerHubAllRepo()
	if err != nil {
		logger.Error("fetchDockerHubAllRepo error %s", err.Error())
		return
	}
	data, err := getCIRun(".cirun.yml")
	if err != nil {
		logger.Error("getCIRun error %s", err.Error())
		return
	}
	logger.Info("get docker hub all repo success")
	for k, v := range got {
		err = generatorSyncFile(syncDir, k, v)
		if err != nil {
			logger.Error("generatorSyncFile %s error %s", k, err.Error())
			continue
		}
		err = generatorWorkflowFile(workflowDir, syncDir, k, data)
		if err != nil {
			logger.Error("generatorWorkflowFile %s error %s", k, err.Error())
			continue
		}
	}
}
