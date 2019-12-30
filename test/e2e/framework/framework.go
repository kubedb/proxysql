/*
Copyright The KubeDB Authors.

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
package framework

import (
	cs "kubedb.dev/apimachinery/client/clientset/versioned"

	"github.com/appscode/go/crypto/rand"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	scs "stash.appscode.dev/stash/client/clientset/versioned"
)

var (
	DockerRegistry           = "kubedbci"
	SelfHostedOperator       = true
	MySQLCatalogName         = "5.7-v2"
	PerconaXtraDBCatalogName = "5.7-cluster"
	ProxySQLCatalogName      = "2.0.4"

	MySQLTest         = true
	PerconaXtraDBTest = true
)

type Framework struct {
	restConfig       *rest.Config
	kubeClient       kubernetes.Interface
	apiExtKubeClient kext_cs.ApiextensionsV1beta1Interface
	dbClient         cs.Interface
	kaClient         ka.Interface
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface
	stashClient      scs.Interface
	namespace        string
	StorageClass     string
}

func New(
	restConfig *rest.Config,
	kubeClient kubernetes.Interface,
	apiExtKubeClient kext_cs.ApiextensionsV1beta1Interface,
	extClient cs.Interface,
	kaClient ka.Interface,
	appCatalogClient appcat_cs.AppcatalogV1alpha1Interface,
	stashClient scs.Interface,
	storageClass string,
) *Framework {
	return &Framework{
		restConfig:       restConfig,
		kubeClient:       kubeClient,
		apiExtKubeClient: apiExtKubeClient,
		dbClient:         extClient,
		kaClient:         kaClient,
		appCatalogClient: appCatalogClient,
		stashClient:      stashClient,
		namespace:        rand.WithUniqSuffix("proxysql"),
		StorageClass:     storageClass,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("proxysql-e2e"),
	}
}

type Invocation struct {
	*Framework
	app string
}

func (fi *Invocation) KubeClient() kubernetes.Interface {
	return fi.kubeClient
}
