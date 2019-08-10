package controller

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/appscode/go/types"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
	config_api "kubedb.dev/apimachinery/apis/config/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
)

func (c *Controller) ensureAppBinding(db *api.PerconaXtraDB) (kutil.VerbType, error) {
	appmeta := db.AppBindingMeta()

	meta := metav1.ObjectMeta{
		Name:      appmeta.Name(),
		Namespace: db.Namespace,
	}

	ref, err := reference.GetReference(clientsetscheme.Scheme, db)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	var peers []string
	for i := 0; i < int(*db.Spec.Replicas); i += 1 {
		peers = append(peers, db.PeerName(i))
	}

	garbdCnfJson, err := json.Marshal(config_api.GaleraArbitratorConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: config_api.SchemeGroupVersion.String(),
			Kind:       config_api.ResourceKindGaleraArbitratorConfiguration,
		},
		Address:   fmt.Sprintf("gcomm://%s", strings.Join(peers, ",")),
		Group:     db.Name,
		SSTMethod: config_api.GarbdXtrabackupSSTMethod,
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	_, vt, err := appcat_util.CreateOrPatchAppBinding(c.AppCatalogClient, meta, func(in *appcat.AppBinding) *appcat.AppBinding {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = db.OffshootLabels()

		in.Spec.Type = appmeta.Type()
		in.Spec.ClientConfig.URL = types.StringP(fmt.Sprintf("tcp(%s:%d)/", db.ServiceName(), defaultDBPort.Port))
		in.Spec.ClientConfig.Service = &appcat.ServiceReference{
			Scheme: "mysql",
			Name:   db.ServiceName(),
			Port:   defaultDBPort.Port,
			Path:   "/",
		}
		in.Spec.ClientConfig.InsecureSkipTLSVerify = false

		in.Spec.Secret = &core.LocalObjectReference{
			Name: db.Spec.DatabaseSecret.SecretName,
		}

		in.Spec.Parameters = &runtime.RawExtension{
			Raw: garbdCnfJson,
		}

		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			db,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s appbinding",
			vt,
		)
	}
	return vt, nil
}
