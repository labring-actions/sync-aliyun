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

var KubeVersions = []string{"18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"}

var KubeKinds = []string{"kubernetes", "kubernetes-crio", "kubernetes-docker"}

var K3sKinds = []string{"k3s", "k3s-crio", "k3s-docker"}

const groupSize = 5

var bigSync = []string{"kubegems", "kubesphere", "deepflow", "rancher"}

const defaultRepo = "labring"
