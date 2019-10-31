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
package e2e_test

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"
	"kubedb.dev/proxysql/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	kext_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/clientcmd"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	scs "stash.appscode.dev/stash/client/clientset/versioned"
)

var (
	storageClass   = "standard"
	kubeconfigPath = func() string {
		kubecfg := os.Getenv("KUBECONFIG")
		if kubecfg != "" {
			return kubecfg
		}
		return filepath.Join(homedir.HomeDir(), ".kube", "config")
	}()
	kubeContext = ""
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientSetScheme.Scheme))

	flag.StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&kubeContext, "kube-context", "", "Name of kube context")
	flag.StringVar(&storageClass, "storageclass", storageClass, "Kubernetes StorageClass name")
	flag.StringVar(&framework.ProxySQLCatalogName, "proxysql-catalog", framework.ProxySQLCatalogName, "ProxySQL version")
	flag.StringVar(&framework.MySQLCatalogName, "mysql-catalog", framework.MySQLCatalogName, "MySQL version")
	flag.StringVar(&framework.PerconaXtraDBCatalogName, "percona-xtradb-catalog", framework.PerconaXtraDBCatalogName, "MySQL version")
	flag.StringVar(&framework.DockerRegistry, "docker-registry", framework.DockerRegistry, "User provided docker repository")
	flag.BoolVar(&framework.SelfHostedOperator, "selfhosted-operator", framework.SelfHostedOperator, "Enable this for provided controller")
	flag.BoolVar(&framework.MySQLTest, "mysql", framework.MySQLTest, "Enable ProxySQL test for MySQL")
	flag.BoolVar(&framework.PerconaXtraDBTest, "percona-xtradb", framework.PerconaXtraDBTest, "Enable ProxySQL test for PerconaXtraDB")
}

const (
	TIMEOUT = 20 * time.Minute
)

var (
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	defer logs.FlushLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)

	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	By("Using kubeconfig from " + kubeconfigPath)
	config, err := clientcmd.BuildConfigFromContext(kubeconfigPath, kubeContext)
	Expect(err).NotTo(HaveOccurred())
	// raise throttling time. ref: https://github.com/appscode/voyager/issues/640
	config.Burst = 100
	config.QPS = 100

	// Clients
	kubeClient := kubernetes.NewForConfigOrDie(config)
	apiExtKubeClient := kext_cs.NewForConfigOrDie(config)
	dbClient := cs.NewForConfigOrDie(config)
	kaClient := ka.NewForConfigOrDie(config)
	appCatalogClient, err := appcat_cs.NewForConfig(config)
	stashClient := scs.NewForConfigOrDie(config)

	if err != nil {
		log.Fatalln(err)
	}

	// Framework
	root = framework.New(config, kubeClient, apiExtKubeClient, dbClient, kaClient, appCatalogClient, stashClient, storageClass)

	// Create namespace
	By("Using namespace " + root.Namespace())
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())

	if !framework.SelfHostedOperator {
		stopCh := genericapiserver.SetupSignalHandler()
		go root.RunOperatorAndServer(config, kubeconfigPath, stopCh)
	}

	root.EventuallyCRD().Should(Succeed())
	// TODO: this needs to be fixed, but for now it has to be commented. Because the apiservices can not be ready when proxysql operator runs along with mysql operator. It says,
	//  message: failing or missing response from https://<operator_service_ip>:443/apis/mutators.kubedb.com/v1alpha1: bad status from https://10.106.49.243:443/apis/mutators.kubedb.com/v1alpha1: 404,
	//  reason: FailedDiscoveryCheck
	//root.EventuallyAPIServiceReady().Should(Succeed())
})

var _ = AfterSuite(func() {

	By("Cleanup Left Overs")
	if !framework.SelfHostedOperator {
		By("Delete Admission Controller Configs")
		root.CleanAdmissionConfigs()
	}
	By("Delete left over MySQL objects")
	root.CleanMySQL()
	By("Delete left over ProxySQL objects")
	root.CleanProxySQL()
	By("Delete left over Dormant Database objects")
	root.CleanDormantDatabase()
	By("Delete Namespace")
	err := root.DeleteNamespace()
	Expect(err).NotTo(HaveOccurred())
})
