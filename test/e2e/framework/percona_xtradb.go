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

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Invocation) PerconaXtraDB() *api.PerconaXtraDB {
	return &api.PerconaXtraDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("percona-xtradb"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.PerconaXtraDBSpec{
			Replicas:    types.Int32P(api.PerconaXtraDBDefaultClusterSize),
			Version:     PerconaXtraDBCatalogName,
			StorageType: api.StorageTypeDurable,
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse(DBPvcStorageSize),
					},
				},
				StorageClassName: types.StringP(f.StorageClass),
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			TerminationPolicy: api.TerminationPolicyWipeOut,
		},
	}
}

func (f *Framework) CreatePerconaXtraDB(obj *api.PerconaXtraDB) error {
	_, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetPerconaXtraDB(meta metav1.ObjectMeta) (*api.PerconaXtraDB, error) {
	return f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) DeletePerconaXtraDB(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyPerconaXtraDBRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			px, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return px.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) CleanPerconaXtraDB() {
	pxList, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, px := range pxList.Items {
		if _, _, err := util.PatchPerconaXtraDB(f.dbClient.KubedbV1alpha1(), &px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			fmt.Printf("error Patching PerconaXtraDB. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of PerconaXtraDB. Error: %v", err)
	}
}
