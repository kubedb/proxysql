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

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	meta_util "kmodules.xyz/client-go/meta"
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
			TerminationPolicy: api.TerminationPolicyWipeOut,
		},
	}
}

func (f *Framework) CreatePerconaXtraDB(obj *api.PerconaXtraDB) error {
	_, err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) GetPerconaXtraDB(meta metav1.ObjectMeta) (*api.PerconaXtraDB, error) {
	return f.dbClient.KubedbV1alpha2().PerconaXtraDBs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchPerconaXtraDB(meta metav1.ObjectMeta, transform func(*api.PerconaXtraDB) *api.PerconaXtraDB) (*api.PerconaXtraDB, error) {
	px, err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	px, _, err = util.PatchPerconaXtraDB(context.TODO(), f.dbClient.KubedbV1alpha2(), px, transform, metav1.PatchOptions{})
	return px, err
}

func (f *Framework) DeletePerconaXtraDB(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha2().PerconaXtraDBs(meta.Namespace).Delete(context.TODO(), meta.Name, metav1.DeleteOptions{})
}

func (f *Framework) EventuallyPerconaXtraDBRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			px, err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return px.Status.Phase == api.DatabasePhaseReady
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) CleanPerconaXtraDB() {
	pxList, err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(f.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, px := range pxList.Items {
		if _, _, err := util.PatchPerconaXtraDB(context.TODO(), f.dbClient.KubedbV1alpha2(), &px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}, metav1.PatchOptions{}); err != nil {
			fmt.Printf("error Patching PerconaXtraDB. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of PerconaXtraDB. Error: %v", err)
	}
}

func (f *Framework) EventuallyPerconaXtraDB(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.dbClient.KubedbV1alpha2().PerconaXtraDBs(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}
