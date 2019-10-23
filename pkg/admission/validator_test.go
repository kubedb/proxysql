package admission

import (
	"net/http"
	"testing"

	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	"kubedb.dev/apimachinery/apis/kubedb"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	extFake "kubedb.dev/apimachinery/client/clientset/versioned/fake"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"

	"github.com/appscode/go/types"
	admission "k8s.io/api/admission/v1beta1"
	apps "k8s.io/api/apps/v1"
	authenticationV1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	storageV1beta1 "k8s.io/api/storage/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	"kmodules.xyz/client-go/meta"
)

func init() {
	scheme.AddToScheme(clientSetScheme.Scheme)
}

var requestKind = metav1.GroupVersionKind{
	Group:   api.SchemeGroupVersion.Group,
	Version: api.SchemeGroupVersion.Version,
	Kind:    api.ResourceKindProxySQL,
}

func TestPerconaXtraDBValidator_Admit(t *testing.T) {
	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			validator := ProxySQLValidator{}

			validator.initialized = true
			validator.extClient = extFake.NewSimpleClientset(
				&catalog.ProxySQLVersion{
					ObjectMeta: metav1.ObjectMeta{
						Name: "2.0.4",
					},
					Spec: catalog.ProxySQLVersionSpec{
						Version: "2.0.4",
					},
				},
				&api.PerconaXtraDB{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar",
						Namespace: "default",
					},
				},
				&api.MySQL{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar",
						Namespace: "default",
					},
				},
			)
			validator.client = fake.NewSimpleClientset(
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-auth",
						Namespace: "default",
					},
				},
				&storageV1beta1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "standard",
					},
				},
			)

			objJS, err := meta.MarshalToJson(&c.object, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}
			oldObjJS, err := meta.MarshalToJson(&c.oldObject, api.SchemeGroupVersion)
			if err != nil {
				panic(err)
			}

			req := new(admission.AdmissionRequest)

			req.Kind = c.kind
			req.Name = c.objectName
			req.Namespace = c.namespace
			req.Operation = c.operation
			req.UserInfo = authenticationV1.UserInfo{}
			req.Object.Raw = objJS
			req.OldObject.Raw = oldObjJS

			if c.heatUp {
				if _, err := validator.extClient.KubedbV1alpha1().ProxySQLs(c.namespace).Create(&c.object); err != nil && !kerr.IsAlreadyExists(err) {
					t.Errorf(err.Error())
				}
			}
			if c.operation == admission.Delete {
				req.Object = runtime.RawExtension{}
			}
			if c.operation != admission.Update {
				req.OldObject = runtime.RawExtension{}
			}

			response := validator.Admit(req)
			if c.result == true {
				if response.Allowed != true {
					t.Errorf("expected: 'Allowed=true'. but got response: %v", response)
				}
			} else if c.result == false {
				if response.Allowed == true || response.Result.Code == http.StatusInternalServerError {
					t.Errorf("expected: 'Allowed=false', but got response: %v", response)
				}
			}
		})
	}
}

var cases = []struct {
	testName   string
	kind       metav1.GroupVersionKind
	objectName string
	namespace  string
	operation  admission.Operation
	object     api.ProxySQL
	oldObject  api.ProxySQL
	heatUp     bool
	result     bool
}{
	{"Create Valid ProxySQL backed for PerconaXtraDB",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleProxySQLForPerconaXtraDB(),
		api.ProxySQL{},
		false,
		true,
	},
	{"Create Valid ProxySQL backed for MySQL",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleProxySQLForMySQL(),
		api.ProxySQL{},
		false,
		true,
	},
	{"Create Invalid proxysql",
		requestKind,
		"foo",
		"default",
		admission.Create,
		getAwkwardProxySQL(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL without single replicas",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithoutSingleReplica(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with empty mode",
		requestKind,
		"foo",
		"default",
		admission.Create,
		sampleProxySQL(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with invalid mode",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithInvalidMode(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with empty backend",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithEmptyBackend(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with empty backend replicas",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithEmptyBackendReplicas(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with empty backend replicas",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithEmptyBackendRef(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with empty backend replicas",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithEmptyBackendAPIGroup(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with invalid group kind",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithInvalidGroupKind(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with unmatched mode-backend-01",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithUnmatchedModeBackend1(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with unmatched mode-backend-02",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithUnmatchedModeBackend2(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with unavailable PerconaXtraDB object",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithUnavailablePerconaXtraDB(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Create ProxySQL with unavailable MySQL object",
		requestKind,
		"foo",
		"default",
		admission.Create,
		proxysqlWithUnavailableMySQL(),
		api.ProxySQL{},
		false,
		false,
	},
	{"Edit ProxySQL .spec.ProxySQLSecret with Existing Secret",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editExistingSecret(sampleProxySQLForPerconaXtraDB()),
		sampleProxySQLForPerconaXtraDB(),
		false,
		true,
	},
	{"Edit ProxySQL Spec.DatabaseSecret with non Existing Secret",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editNonExistingSecret(sampleProxySQLForPerconaXtraDB()),
		sampleProxySQLForPerconaXtraDB(),
		false,
		true,
	},
	{"Edit Status",
		requestKind,
		"foo",
		"default",
		admission.Update,
		editStatus(sampleProxySQLForPerconaXtraDB()),
		sampleProxySQLForPerconaXtraDB(),
		false,
		true,
	},
	{"Delete ProxySQL",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		sampleProxySQLForPerconaXtraDB(),
		api.ProxySQL{},
		true,
		true,
	},
	{"Delete Non Existing ProxySQL",
		requestKind,
		"foo",
		"default",
		admission.Delete,
		api.ProxySQL{},
		api.ProxySQL{},
		false,
		true,
	},
}

func sampleProxySQL() api.ProxySQL {
	return api.ProxySQL{
		TypeMeta: metav1.TypeMeta{
			Kind:       api.ResourceKindProxySQL,
			APIVersion: api.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindProxySQL,
			},
		},
		Spec: api.ProxySQLSpec{
			Version:  "2.0.4",
			Replicas: types.Int32P(1),
			Backend: &api.ProxySQLBackendSpec{
				Ref: &corev1.TypedLocalObjectReference{
					APIGroup: types.StringP(kubedb.GroupName),
					Name:     "bar",
				},
			},
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
		},
	}
}

func sampleProxySQLForPerconaXtraDB() api.ProxySQL {
	proxysql := sampleProxySQL()
	mode := api.LoadBalanceModeGalera
	proxysql.Spec.Mode = &mode
	proxysql.Spec.Backend.Ref.Kind = api.ResourceKindPerconaXtraDB
	proxysql.Spec.Backend.Replicas = types.Int32P(api.PerconaXtraDBDefaultClusterSize)

	return proxysql
}

func sampleProxySQLForMySQL() api.ProxySQL {
	proxysql := sampleProxySQL()
	mode := api.LoadBalanceModeGroupReplication
	proxysql.Spec.Mode = &mode
	proxysql.Spec.Backend.Ref.Kind = api.ResourceKindMySQL
	proxysql.Spec.Backend.Replicas = types.Int32P(api.MySQLDefaultGroupSize)

	return proxysql
}

func getAwkwardProxySQL() api.ProxySQL {
	proxysql := sampleProxySQL()
	proxysql.Spec.Version = "0.1"
	return proxysql
}

func proxysqlWithoutSingleReplica() api.ProxySQL {
	proxysql := sampleProxySQL()
	proxysql.Spec.Replicas = types.Int32P(3)
	return proxysql
}

func proxysqlWithInvalidMode() api.ProxySQL {
	proxysql := sampleProxySQL()
	mode := api.LoadBalanceMode("Sentinel")
	proxysql.Spec.Mode = &mode
	return proxysql
}

func proxysqlWithEmptyBackend() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend = nil
	return proxysql
}

func proxysqlWithEmptyBackendReplicas() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend.Replicas = nil
	return proxysql
}

func proxysqlWithEmptyBackendRef() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend.Ref = nil
	return proxysql
}

func proxysqlWithEmptyBackendAPIGroup() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend.Ref = nil
	return proxysql
}

func proxysqlWithInvalidGroupKind() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend.Ref.Kind = api.ResourceKindPostgres
	return proxysql
}

func proxysqlWithUnmatchedModeBackend1() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	mode := api.LoadBalanceModeGroupReplication
	proxysql.Spec.Mode = &mode
	return proxysql
}

func proxysqlWithUnmatchedModeBackend2() api.ProxySQL {
	proxysql := sampleProxySQLForMySQL()
	mode := api.LoadBalanceModeGalera
	proxysql.Spec.Mode = &mode
	return proxysql
}

func proxysqlWithUnavailablePerconaXtraDB() api.ProxySQL {
	proxysql := sampleProxySQLForPerconaXtraDB()
	proxysql.Spec.Backend.Ref.Name = "unavailable-bar"
	return proxysql
}

func proxysqlWithUnavailableMySQL() api.ProxySQL {
	proxysql := sampleProxySQLForMySQL()
	proxysql.Spec.Backend.Ref.Name = "unavailable-bar"
	return proxysql
}

func editExistingSecret(old api.ProxySQL) api.ProxySQL {
	old.Spec.ProxySQLSecret = &corev1.SecretVolumeSource{
		SecretName: "foo-auth",
	}
	return old
}

func editNonExistingSecret(old api.ProxySQL) api.ProxySQL {
	old.Spec.ProxySQLSecret = &corev1.SecretVolumeSource{
		SecretName: "foo-auth-fused",
	}
	return old
}

func editStatus(old api.ProxySQL) api.ProxySQL {
	old.Status = api.ProxySQLStatus{
		Phase: api.DatabasePhaseCreating,
	}
	return old
}
