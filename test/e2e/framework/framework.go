package framework

import (
	"github.com/appscode/go/crypto/rand"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	scs "stash.appscode.dev/stash/client/clientset/versioned"
)

var (
	DockerRegistry     = "kubedbci"
	SelfHostedOperator = false
	DBCatalogName      = "5.7"
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
	name             string
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

func (fi *Invocation) App() string {
	return fi.app
}

func (fi *Invocation) ExtClient() cs.Interface {
	return fi.dbClient
}

func (fi *Invocation) KubeClient() kubernetes.Interface {
	return fi.kubeClient
}

type Invocation struct {
	*Framework
	app string
}
