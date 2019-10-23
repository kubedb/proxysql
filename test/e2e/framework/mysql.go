package framework

import (
	"fmt"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	JobPvcStorageSize = "2Gi"
	DBPvcStorageSize  = "1Gi"
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
			Version:     jsonTypes.StrYo(MySQLCatalogName),
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

func (f *Invocation) MySQLGroup() *api.MySQL {
	mysql := f.MySQL()
	mysql.Spec.Replicas = types.Int32P(api.MySQLDefaultGroupSize)
	clusterMode := api.MySQLClusterModeGroup
	mysql.Spec.Topology = &api.MySQLClusterTopology{
		Mode: &clusterMode,
		Group: &api.MySQLGroupSpec{
			Name:         "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b",
			BaseServerID: types.UIntP(api.MySQLDefaultBaseServerID),
		},
	}

	return mysql
}

func (f *Framework) CreateMySQL(obj *api.MySQL) error {
	_, err := f.dbClient.KubedbV1alpha1().MySQLs(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetMySQL(meta metav1.ObjectMeta) (*api.MySQL, error) {
	return f.dbClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) DeleteMySQL(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().MySQLs(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyMySQLRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			mysql, err := f.dbClient.KubedbV1alpha1().MySQLs(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return mysql.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*15,
		time.Second*5,
	)
}

func (f *Framework) CleanMySQL() {
	mysqlList, err := f.dbClient.KubedbV1alpha1().MySQLs(f.namespace).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range mysqlList.Items {
		if _, _, err := util.PatchMySQL(f.dbClient.KubedbV1alpha1(), &e, func(in *api.MySQL) *api.MySQL {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			fmt.Printf("error Patching MySQL. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha1().MySQLs(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of MySQL. Error: %v", err)
	}
}
