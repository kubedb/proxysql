package controller

import (
	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
	meta_util "kmodules.xyz/client-go/meta"
	"kubedb.dev/apimachinery/apis"
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
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(proxysql.ObjectMeta)
		return nil
	}

	//// Delete Matching DormantDatabase if exists any
	//if err := c.deleteMatchingDormantDatabase(proxysql); err != nil {
	//	return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, proxysql.Namespace, proxysql.Name, err)
	//}

	if proxysql.Status.Phase == "" {
		proxysqlUpd, err := util.UpdateProxySQLStatus(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
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

	per, err := util.UpdateProxySQLStatus(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQLStatus) *api.ProxySQLStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(proxysql.Generation, meta_util.GenerationHash(proxysql))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	proxysql.Status = per.Status

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

	//_, err = c.ensureAppBinding(proxysql)
	//if err != nil {
	//	log.Errorln(err)
	//	return err
	//}

	return nil
}

func (c *Controller) terminate(proxysql *api.ProxySQL) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, proxysql)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if proxysql.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(proxysql, ref); err != nil {
			return err
		}

		//if _, err := c.createDormantDatabase(proxysql); err != nil {
		//	if kerr.IsAlreadyExists(err) {
		//		// if already exists, check if it is database of another Kind and return error in that case.
		//		// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
		//		// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
		//		// So reuse that DormantDB!
		//		ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(proxysql.Namespace).Get(proxysql.Name, metav1.GetOptions{})
		//		if err != nil {
		//			return err
		//		}
		//		if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindProxySQL {
		//			return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, proxysql.Name, val)
		//		}
		//	} else {
		//		return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, proxysql.Namespace, proxysql.Name, err)
		//	}
		//}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(proxysql, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(proxysql.ObjectMeta)

	if proxysql.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(proxysql); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(proxysql *api.ProxySQL, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(proxysql.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if proxysql.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		//if err := dynamic_util.EnsureOwnerReferenceForSelector(
		//	c.DynamicClient,
		//	api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		//	proxysql.Namespace,
		//	selector,
		//	ref); err != nil {
		//	return err
		//}
		if err := c.wipeOutDatabase(proxysql.ObjectMeta, proxysql.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		//if err := dynamic_util.RemoveOwnerReferenceForSelector(
		//	c.DynamicClient,
		//	api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		//	proxysql.Namespace,
		//	selector,
		//	ref); err != nil {
		//	return err
		//}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			proxysql.Namespace,
			proxysql.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		proxysql.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(proxysql *api.ProxySQL, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(proxysql.OffshootSelectors())

	//if err := dynamic_util.RemoveOwnerReferenceForSelector(
	//	c.DynamicClient,
	//	api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
	//	proxysql.Namespace,
	//	labelSelector,
	//	ref); err != nil {
	//	return err
	//}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		proxysql.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		proxysql.Namespace,
		proxysql.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}
