/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"
	"fmt"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	. "github.com/onsi/gomega"
	"gomodules.xyz/pointer"
	"gomodules.xyz/x/crypto/rand"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

func (f *Invocation) ProxySQL(backendResourceKind, backendObjName string) *api.ProxySQL {
	var mode api.LoadBalanceMode

	if backendResourceKind == api.ResourceKindMySQL {
		mode = api.LoadBalanceModeGroupReplication
	} else {
		mode = api.LoadBalanceModeGalera
	}

	return &api.ProxySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("proxysql"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.ProxySQLSpec{
			Version:  ProxySQLCatalogName,
			Replicas: pointer.Int32P(1),
			Mode:     &mode,
			Backend: &api.ProxySQLBackendSpec{
				Ref: &corev1.TypedLocalObjectReference{
					APIGroup: pointer.StringP(kubedb.GroupName),
					Kind:     backendResourceKind,
					Name:     backendObjName,
				},
				Replicas: pointer.Int32P(api.PerconaXtraDBDefaultClusterSize),
			},
		},
	}
}

func (f *Framework) CreateProxySQL(obj *api.ProxySQL) error {
	_, err := f.dbClient.KubedbV1alpha2().ProxySQLs(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) GetProxySQL(meta metav1.ObjectMeta) (*api.ProxySQL, error) {
	return f.dbClient.KubedbV1alpha2().ProxySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
}

func (f *Framework) DeleteProxySQL(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha2().ProxySQLs(meta.Namespace).Delete(context.TODO(), meta.Name, metav1.DeleteOptions{})
}

func (f *Framework) EventuallyProxySQLPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() api.DatabasePhase {
			db, err := f.dbClient.KubedbV1alpha2().ProxySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return db.Status.Phase
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) CleanProxySQL() {
	mysqlList, err := f.dbClient.KubedbV1alpha2().ProxySQLs(f.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range mysqlList.Items {
		if _, _, err := util.PatchProxySQL(context.TODO(), f.dbClient.KubedbV1alpha2(), &e, func(in *api.ProxySQL) *api.ProxySQL {
			in.ObjectMeta.Finalizers = nil
			return in
		}, metav1.PatchOptions{}); err != nil {
			fmt.Printf("error Patching MySQL. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha2().MySQLs(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of MySQL. Error: %v", err)
	}
}
