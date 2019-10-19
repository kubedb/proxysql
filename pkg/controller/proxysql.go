package controller

import (
	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/proxysql/pkg/admission"
)

func (c *Controller) create(proxysql *api.ProxySQL) error {
	if err := validator.ValidateProxySQL(c.Client, c.ExtClient, proxysql, true); err != nil {
		c.recorder.Event(
			proxysql,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil
	}

	if proxysql.Status.Phase == "" {
		proxysqlUpd, err := util.UpdateProxySQLStatus(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		})
		if err != nil {
			return err
		}
		proxysql.Status = proxysqlUpd.Status
	}

	// create Governing Service
	if err := c.CreateGoverningService(c.GoverningService, proxysql.Namespace); err != nil {
		return err
	}

	if c.EnableRBAC {
		// Ensure ClusterRoles for statefulsets
		if err := c.ensureRBACStuff(proxysql); err != nil {
			return err
		}
	}

	// ensure database Service
	vt1, err := c.ensureService(proxysql)
	if err != nil {
		return err
	}

	if err := c.ensureProxySQLSecret(proxysql); err != nil {
		return err
	}

	// ensure proxysql StatefulSet
	vt2, err := c.ensureProxySQLNode(proxysql)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created ProxySQL",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched ProxySQL",
		)
	}

	proxysqlUpd, err := util.UpdateProxySQLStatus(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(proxysql.Generation, meta_util.GenerationHash(proxysql))
		return in
	})
	if err != nil {
		return err
	}
	proxysql.Status = proxysqlUpd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(proxysql); err != nil {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(proxysql); err != nil {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	return nil
}

func (c *Controller) terminate(proxysql *api.ProxySQL) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, proxysql)
	if rerr != nil {
		return rerr
	}

	// delete PVC
	selector := labels.SelectorFromSet(proxysql.OffshootSelectors())
	if err := dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		proxysql.Namespace,
		selector,
		ref); err != nil {
		return err
	}

	if proxysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(proxysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}

	return nil
}
