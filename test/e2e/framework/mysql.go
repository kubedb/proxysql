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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha2/util"

	. "github.com/onsi/gomega"
	"gomodules.xyz/pointer"
	"gomodules.xyz/x/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
)

var (
	DBPvcStorageSize = "1Gi"
)

func (f *Invocation) MySQL() *api.MySQL {
	return &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("mysql"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.MySQLSpec{
			Version:     MySQLCatalogName,
			StorageType: api.StorageTypeDurable,
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse(DBPvcStorageSize),
					},
				},
				StorageClassName: pointer.StringP(f.StorageClass),
			},
			TerminationPolicy: api.TerminationPolicyWipeOut,
		},
	}
}

func (f *Invocation) MySQLGroup() *api.MySQL {
	mysql := f.MySQL()
	mysql.Spec.Replicas = pointer.Int32P(api.MySQLDefaultGroupSize)
	clusterMode := api.MySQLClusterModeGroup
	mysql.Spec.Topology = &api.MySQLClusterTopology{
		Mode: &clusterMode,
		Group: &api.MySQLGroupSpec{
			Name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b",
		},
	}

	return mysql
}

func (f *Framework) CreateMySQL(obj *api.MySQL) error {
	_, err := f.dbClient.KubedbV1alpha2().MySQLs(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) GetMySQL(meta metav1.ObjectMeta) (*api.MySQL, error) {
	return f.dbClient.KubedbV1alpha2().MySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchMySQL(meta metav1.ObjectMeta, transform func(*api.MySQL) *api.MySQL) (*api.MySQL, error) {
	mysql, err := f.dbClient.KubedbV1alpha2().MySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	mysql, _, err = util.PatchMySQL(context.TODO(), f.dbClient.KubedbV1alpha2(), mysql, transform, metav1.PatchOptions{})
	return mysql, err
}

func (f *Framework) DeleteMySQL(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha2().MySQLs(meta.Namespace).Delete(context.TODO(), meta.Name, metav1.DeleteOptions{})
}

func (f *Framework) EventuallyMySQLRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			mysql, err := f.dbClient.KubedbV1alpha2().MySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return mysql.Status.Phase == api.DatabasePhaseReady
		},
		time.Minute*15,
		time.Second*5,
	)
}

func (f *Framework) CleanMySQL() {
	mysqlList, err := f.dbClient.KubedbV1alpha2().MySQLs(f.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range mysqlList.Items {
		if _, _, err := util.PatchMySQL(context.TODO(), f.dbClient.KubedbV1alpha2(), &e, func(in *api.MySQL) *api.MySQL {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}, metav1.PatchOptions{}); err != nil {
			fmt.Printf("error Patching MySQL. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha2().MySQLs(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of MySQL. Error: %v", err)
	}
}

func (f *Framework) EventuallyMySQL(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.dbClient.KubedbV1alpha2().MySQLs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		},
		time.Minute*12,
		time.Second*5,
	)
}
