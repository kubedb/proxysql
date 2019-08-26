package admission

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	admission "k8s.io/api/admission/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	meta_util "kmodules.xyz/client-go/meta"
	hookapi "kmodules.xyz/webhook-runtime/admission/v1beta1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amv "kubedb.dev/apimachinery/pkg/validator"
)

// ProxySQLValidator implements the AdmissionHook interface to validate the ProxySQL resources
type ProxySQLValidator struct {
	client      kubernetes.Interface
	extClient   cs.Interface
	lock        sync.RWMutex
	initialized bool
}

var _ hookapi.AdmissionHook = &ProxySQLValidator{}

var forbiddenEnvVars = []string{
	"MYSQL_ROOT_PASSWORD",
	"MYSQL_ALLOW_EMPTY_PASSWORD",
	"MYSQL_RANDOM_ROOT_PASSWORD",
	"MYSQL_ONETIME_PASSWORD",
}

// Resource is the resource to use for hosting validating admission webhook.
func (a *ProxySQLValidator) Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validators.kubedb.com",
			Version:  "v1alpha1",
			Resource: "ProxySQLvalidators",
		},
		"ProxySQLvalidator"
}

// Initialize is called as a post-start hook
func (a *ProxySQLValidator) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.initialized = true

	var err error
	if a.client, err = kubernetes.NewForConfig(config); err != nil {
		return err
	}
	if a.extClient, err = cs.NewForConfig(config); err != nil {
		return err
	}
	return err
}

// Admit is called to decide whether to accept the admission request.
func (a *ProxySQLValidator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}

	if (req.Operation != admission.Create && req.Operation != admission.Update && req.Operation != admission.Delete) ||
		len(req.SubResource) != 0 ||
		req.Kind.Group != api.SchemeGroupVersion.Group ||
		req.Kind.Kind != api.ResourceKindProxySQL {
		status.Allowed = true
		return status
	}

	a.lock.RLock()
	defer a.lock.RUnlock()
	if !a.initialized {
		return hookapi.StatusUninitialized()
	}

	switch req.Operation {
	case admission.Delete:
		if req.Name != "" {
			// req.Object.Raw = nil, so read from kubernetes
			obj, err := a.extClient.KubedbV1alpha1().ProxySQLs(req.Namespace).Get(req.Name, metav1.GetOptions{})
			if err != nil && !kerr.IsNotFound(err) {
				return hookapi.StatusInternalServerError(err)
			} else if err == nil && obj.Spec.TerminationPolicy == api.TerminationPolicyDoNotTerminate {
				return hookapi.StatusBadRequest(fmt.Errorf(`proxysql "%v/%v" can't be paused. To delete, change spec.terminationPolicy`, req.Namespace, req.Name))
			}
		}
	default:
		obj, err := meta_util.UnmarshalFromJSON(req.Object.Raw, api.SchemeGroupVersion)
		if err != nil {
			return hookapi.StatusBadRequest(err)
		}
		if req.Operation == admission.Update {
			// validate changes made by user
			oldObject, err := meta_util.UnmarshalFromJSON(req.OldObject.Raw, api.SchemeGroupVersion)
			if err != nil {
				return hookapi.StatusBadRequest(err)
			}

			proxysql := obj.(*api.ProxySQL).DeepCopy()
			oldProxysql := oldObject.(*api.ProxySQL).DeepCopy()
			oldProxysql.SetDefaults()
			// Allow changing Database Secret only if there was no secret have set up yet.
			if oldProxysql.Spec.ProxySQLSecret == nil {
				oldProxysql.Spec.ProxySQLSecret = proxysql.Spec.ProxySQLSecret
			}

			if err := validateUpdate(proxysql, oldProxysql, req.Kind.Kind); err != nil {
				return hookapi.StatusBadRequest(fmt.Errorf("%v", err))
			}
		}
		// validate database specs
		if err = ValidateProxySQL(a.client, a.extClient, obj.(*api.ProxySQL), false); err != nil {
			return hookapi.StatusForbidden(err)
		}
	}
	status.Allowed = true
	return status
}

// validatePXC checks whether the configurations for ProxySQL Cluster are ok
func validatePXC(proxysql *api.ProxySQL) error {


	return nil
}

// ValidateProxySQL checks if the object satisfies all the requirements.
// It is not method of Interface, because it is referenced from controller package too.
func ValidateProxySQL(client kubernetes.Interface, extClient cs.Interface, proxysql *api.ProxySQL, strictValidation bool) error {
	if proxysql.Spec.Version == "" {
		return errors.New(`'spec.version' is missing`)
	}
	var proxysqlVersion api.ProxySQLVersion
	var err error
	if proxysqlVersion, err = extClient.CatalogV1alpha1().ProxySQLVersions().Get(string(proxysql.Spec.Version), metav1.GetOptions{}); err != nil {
		return err
	}

	if proxysql.Spec.Replicas == nil {
		return fmt.Errorf(`'spec.replicas' "%v" invalid. Value must be 1 for standalone proxysql server, but for proxysql cluster, value must be greater than 0`,
			proxysql.Spec.Replicas)
	}

	if *proxysql.Spec.Replicas != 1 {
		return errors.Errorf(`'.spec.replicas' "%v" is invalid. Currently, supported replicas for proxysql is 1`,
			proxysql.Spec.Replicas)
	}

	if err = amv.ValidateEnvVar(proxysql.Spec.PodTemplate.Spec.Env, forbiddenEnvVars, api.ResourceKindProxySQL); err != nil {
		return err
	}

	if proxysql.Spec.StorageType == "" {
		return fmt.Errorf(`'spec.storageType' is missing`)
	}
	if proxysql.Spec.StorageType != api.StorageTypeDurable && proxysql.Spec.StorageType != api.StorageTypeEphemeral {
		return fmt.Errorf(`'spec.storageType' %s is invalid`, proxysql.Spec.StorageType)
	}
	if err = amv.ValidateStorage(client, proxysql.Spec.StorageType, proxysql.Spec.Storage); err != nil {
		return err
	}

	proxysqlSecret := proxysql.Spec.ProxySQLSecret

	if strictValidation {
		if proxysqlSecret != nil {
			if _, err = client.CoreV1().Secrets(proxysql.Namespace).Get(proxysqlSecret.SecretName, metav1.GetOptions{}); err != nil {
				return err
			}
		}

		// Check if proxysql Version is deprecated.
		// If deprecated, return error
		if proxysqlVersion.Spec.Deprecated {
			return fmt.Errorf("proxysql %s/%s is using deprecated version %v. Skipped processing", proxysql.Namespace, proxysql.Name, proxysqlVersion.Name)
		}
	}

	if proxysql.Spec.UpdateStrategy.Type == "" {
		return fmt.Errorf(`'spec.updateStrategy.type' is missing`)
	}

	if proxysql.Spec.TerminationPolicy == "" {
		return fmt.Errorf(`'spec.terminationPolicy' is missing`)
	}

	if proxysql.Spec.StorageType == api.StorageTypeEphemeral && proxysql.Spec.TerminationPolicy == api.TerminationPolicyPause {
		return fmt.Errorf(`'spec.terminationPolicy: Pause' can not be used for 'Ephemeral' storage`)
	}

	monitorSpec := proxysql.Spec.Monitor
	if monitorSpec != nil {
		if err = amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}
	}

	return nil
}

func validateUpdate(obj, oldObj runtime.Object, kind string) error {
	preconditions := getPreconditionFunc()
	_, err := meta_util.CreateStrategicPatch(oldObj, obj, preconditions...)
	if err != nil {
		if mergepatch.IsPreconditionFailed(err) {
			return fmt.Errorf("%v.%v", err, preconditionFailedError(kind))
		}
		return err
	}
	return nil
}

func getPreconditionFunc() []mergepatch.PreconditionFunc {
	preconditions := []mergepatch.PreconditionFunc{
		mergepatch.RequireKeyUnchanged("apiVersion"),
		mergepatch.RequireKeyUnchanged("kind"),
		mergepatch.RequireMetadataKeyUnchanged("name"),
		mergepatch.RequireMetadataKeyUnchanged("namespace"),
	}

	for _, field := range preconditionSpecFields {
		preconditions = append(preconditions,
			meta_util.RequireChainKeyUnchanged(field),
		)
	}
	return preconditions
}

var preconditionSpecFields = []string{
	"spec.storageType",
	"spec.storage",
	"spec.databaseSecret",
	"spec.init",
	"spec.podTemplate.spec.nodeSelector",
}

func preconditionFailedError(kind string) error {
	str := preconditionSpecFields
	strList := strings.Join(str, "\n\t")
	return fmt.Errorf(strings.Join([]string{`At least one of the following was changed:
	apiVersion
	kind
	name
	namespace`, strList}, "\n\t"))
}
