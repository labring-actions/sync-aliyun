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
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const tmpl = `docker.io:
  images:
    {{- range . }}
    labring/{{ . }}: [ ]
    {{- end }}
  tls-verify: false
`

const workflowTmpl = `name: skopeo-sync
on:
  push:
    branches: [ main ]
    paths:
      - "skopeo/{{ .PREFIX }}*"
      - ".github/workflows/{{ .PREFIX }}*"
  schedule:
    - cron: '0 16 * * *'
  workflow_dispatch:

env:
  USERNAME: {{ .USER_KEY }}
  PASSWORD: {{ .PASSWORD_KEY }}

jobs:
  image-sync:
    runs-on: ubuntu-22.04

    steps:
      - name: Checkout
        uses: actions/checkout@v2

      - name: check podman
        run: |
          sudo podman version

      - name: sync images
        run: |
          sudo podman run -it --rm -v ${PWD}:/workspace -w /workspace quay.io/skopeo/stable:latest \
          sync --src yaml --dest docker {{ .SYNC_FILE }} {{ .REGISTRY_KEY }}/{{ .REPOSITORY_KEY }} \
          --dest-username $USERNAME --dest-password "$PASSWORD" \
          --keep-going --retry-times 2 --all
`

const prefix = "auto-sync"

func autoRemoveGenerator(dir string) error {
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果不是目录，并且文件名以指定的前缀开始，就删除它
		if !info.IsDir() && strings.HasPrefix(info.Name(), prefix) {
			os.Remove(path)
		}

		return nil
	}); err != nil {
		return err
	}
	logger.Info("auto remove %s files success", dir)
	return nil
}

func generatorSyncFile(dir, key string, repos []string) error {
	f, err := os.Create(path.Join(dir, fmt.Sprintf("%s-%s.yaml", prefix, key)))
	if err != nil {
		return err
	}
	defer f.Close()
	t := template.Must(template.New("repos").Parse(tmpl))

	err = t.Execute(f, repos)
	if err != nil {
		return err
	}
	logger.Info("generator sync config %s-%s.yaml success", prefix, key)
	return nil
}

func generatorWorkflowFile(dir, syncDir, key string) error {
	syncFile := path.Join(syncDir, fmt.Sprintf("%s-%s.yaml", prefix, key))
	f, err := os.Create(path.Join(dir, fmt.Sprintf("%s-%s.yaml", prefix, key)))
	if err != nil {
		return err
	}
	defer f.Close()
	t := template.Must(template.New("repos").Parse(workflowTmpl))

	err = t.Execute(f, map[string]string{
		"PREFIX":         prefix,
		"SYNC_FILE":      syncFile,
		"USER_KEY":       "${{ vars.A_REGISTRY_USERNAME }}",
		"PASSWORD_KEY":   "${{ secrets.A_REGISTRY_TOKEN }}",
		"REGISTRY_KEY":   "${{ vars.A_REGISTRY_NAME }}",
		"REPOSITORY_KEY": "${{ vars.A_REGISTRY_REPOSITORY }}",
	})
	if err != nil {
		return err
	}
	logger.Info("generator workflow config %s-%s.yaml success", prefix, key)
	return nil
}
