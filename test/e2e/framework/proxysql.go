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
	"fmt"
	"time"

	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Replicas: types.Int32P(1),
			Mode:     &mode,
			Backend: &api.ProxySQLBackendSpec{
				Ref: &corev1.TypedLocalObjectReference{
					APIGroup: types.StringP(kubedb.GroupName),
					Kind:     backendResourceKind,
					Name:     backendObjName,
				},
				Replicas: types.Int32P(api.PerconaXtraDBDefaultClusterSize),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

func (f *Framework) CreateProxySQL(obj *api.ProxySQL) error {
	_, err := f.dbClient.KubedbV1alpha1().ProxySQLs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetProxySQL(meta metav1.ObjectMeta) (*api.ProxySQL, error) {
	return f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) DeleteProxySQL(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyProxySQLPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() api.DatabasePhase {
			db, err := f.dbClient.KubedbV1alpha1().ProxySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return db.Status.Phase
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) CleanProxySQL() {
	mysqlList, err := f.dbClient.KubedbV1alpha1().ProxySQLs(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range mysqlList.Items {
		if _, _, err := util.PatchProxySQL(f.dbClient.KubedbV1alpha1(), &e, func(in *api.ProxySQL) *api.ProxySQL {
			in.ObjectMeta.Finalizers = nil
			return in
		}); err != nil {
			fmt.Printf("error Patching MySQL. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha1().MySQLs(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of MySQL. Error: %v", err)
	}
}
