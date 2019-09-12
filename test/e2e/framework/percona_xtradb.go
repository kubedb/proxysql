package framework

import (
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/wait"
	kutil "kmodules.xyz/client-go"

	//"github.com/appscode/go/crypto/rand"
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

var (
	JobPvcStorageSize = "50Mi"
	DBPvcStorageSize  = "50Mi"
)

func (f *Invocation) PerconaXtraDB() *api.PerconaXtraDB {
	return &api.PerconaXtraDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("proxysql"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.PerconaXtraDBSpec{
			Version: jsonTypes.StrYo(DBCatalogName),
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse(DBPvcStorageSize),
					},
				},
				StorageClassName: types.StringP(f.StorageClass),
			},
		},
	}
}

func (f *Invocation) PerconaXtraDBCluster() *api.PerconaXtraDB {
	perconaxtradb := f.PerconaXtraDB()
	perconaxtradb.Spec.Replicas = types.Int32P(api.PerconaXtraDBDefaultClusterSize)
	//perconaxtradb.Spec.PXC = &api.PXCSpec{
	//	ClusterName: perconaxtradb.Name,
	//	Proxysql: api.ProxysqlSpec{
	//		Replicas: types.Int32P(1),
	//	},
	//}

	return perconaxtradb
}

func (f *Framework) CreatePerconaXtraDB(obj *api.PerconaXtraDB) error {
	_, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetPerconaXtraDB(meta metav1.ObjectMeta) (*api.PerconaXtraDB, error) {
	return f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchPerconaXtraDB(meta metav1.ObjectMeta, transform func(*api.PerconaXtraDB) *api.PerconaXtraDB) (*api.PerconaXtraDB, error) {
	px, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	px, _, err = util.PatchPerconaXtraDB(f.dbClient.KubedbV1alpha1(), px, transform)
	return px, err
}

func (f *Framework) DeletePerconaXtraDB(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyPerconaXtraDB(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
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

func (f *Framework) EventuallyPerconaXtraDBPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() api.DatabasePhase {
			db, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return db.Status.Phase
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyPerconaXtraDBRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			px, err := f.dbClient.KubedbV1alpha1().PerconaXtraDBs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return px.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*15,
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

func (f *Framework) WaitUntilPerconaXtraDBReplicasBePatched(meta metav1.ObjectMeta, count int32) error {
	return wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		px, err := f.GetPerconaXtraDB(meta)
		if err != nil {
			return false, nil
		}

		if *px.Spec.Replicas != count {
			return false, nil
		}

		return true, nil
	})
}
