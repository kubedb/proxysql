package controller

import (
	"fmt"

	"github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	validator "kubedb.dev/percona-xtradb/pkg/admission"
)

func (c *Controller) create(px *api.PerconaXtraDB) error {
	if err := validator.ValidatePerconaXtraDB(c.Client, c.ExtClient, px, true); err != nil {
		c.recorder.Event(
			px,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		// stop Scheduler in case there is any.
		c.cronController.StopBackupScheduling(px.ObjectMeta)
		return nil
	}

	// Delete Matching DormantDatabase if exists any
	if err := c.deleteMatchingDormantDatabase(px); err != nil {
		return fmt.Errorf(`failed to delete dormant Database : "%v/%v". Reason: %v`, px.Namespace, px.Name, err)
	}

	if px.Status.Phase == "" {
		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status
	}

	// Set status as "Initializing" until specified restoresession object be succeeded, if provided
	if _, err := meta_util.GetString(px.Annotations, api.AnnotationInitialized); err == kutil.ErrNotFound &&
		px.Spec.Init != nil && px.Spec.Init.StashRestoreSession != nil {

		if px.Status.Phase == api.DatabasePhaseInitializing {
			return nil
		}

		perconaxtradb, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
			in.Phase = api.DatabasePhaseInitializing
			return in
		}, apis.EnableStatusSubresource)
		if err != nil {
			return err
		}
		px.Status = perconaxtradb.Status

		log.Debugf("PerconaXtraDB %v/%v is waiting for restoreSession to be succeeded", px.Namespace, px.Name)
		return nil
	}

	// create Governing Service
	governingService, err := c.createPerconaXtraDBGoverningService(px)
	if err != nil {
		return fmt.Errorf(`failed to create Service: "%v/%v". Reason: %v`, px.Namespace, governingService, err)
	}
	c.GoverningService = governingService

	if c.EnableRBAC {
		// Ensure ClusterRoles for statefulsets
		if err := c.ensureRBACStuff(px); err != nil {
			return err
		}
	}

	// ensure database Service
	vt1, err := c.ensureService(px)
	if err != nil {
		return err
	}

	if err := c.ensureDatabaseSecret(px); err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensurePerconaXtraDBNode(px)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created PerconaXtraDB",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			px,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched PerconaXtraDB",
		)
	}

	per, err := util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = types.NewIntHash(px.Generation, meta_util.GenerationHash(px))
		return in
	}, apis.EnableStatusSubresource)
	if err != nil {
		return err
	}
	px.Status = per.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(px); err != nil {
		c.recorder.Eventf(
			px,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	if err := c.manageMonitor(px); err != nil {
		c.recorder.Eventf(
			px,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorln(err)
		return nil
	}

	_, err = c.ensureAppBinding(px)
	if err != nil {
		log.Errorln(err)
		return err
	}

	return nil
}

func (c *Controller) terminate(px *api.PerconaXtraDB) error {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, px)
	if rerr != nil {
		return rerr
	}

	// If TerminationPolicy is "pause", keep everything (ie, PVCs,Secrets,Snapshots) intact.
	// In operator, create dormantdatabase
	if px.Spec.TerminationPolicy == api.TerminationPolicyPause {
		if err := c.removeOwnerReferenceFromOffshoots(px, ref); err != nil {
			return err
		}

		if _, err := c.createDormantDatabase(px); err != nil {
			if kerr.IsAlreadyExists(err) {
				// if already exists, check if it is database of another Kind and return error in that case.
				// If the Kind is same, we can safely assume that the DormantDB was not deleted in before,
				// Probably because, User is more faster (create-delete-create-again-delete...) than operator!
				// So reuse that DormantDB!
				ddb, err := c.ExtClient.KubedbV1alpha1().DormantDatabases(px.Namespace).Get(px.Name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if val, _ := meta_util.GetStringValue(ddb.Labels, api.LabelDatabaseKind); val != api.ResourceKindPerconaXtraDB {
					return fmt.Errorf(`DormantDatabase "%v" of kind %v already exists`, px.Name, val)
				}
			} else {
				return fmt.Errorf(`failed to create DormantDatabase: "%v/%v". Reason: %v`, px.Namespace, px.Name, err)
			}
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(px, ref); err != nil {
			return err
		}
	}

	c.cronController.StopBackupScheduling(px.ObjectMeta)

	if px.Spec.Monitor != nil {
		if _, err := c.deleteMonitor(px); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(px *api.PerconaXtraDB, ref *core.ObjectReference) error {
	selector := labels.SelectorFromSet(px.OffshootSelectors())

	// If TerminationPolicy is "wipeOut", delete snapshots and secrets,
	// else, keep it intact.
	if px.Spec.TerminationPolicy == api.TerminationPolicyWipeOut {
		if err := dynamic_util.EnsureOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			px.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := c.wipeOutDatabase(px.ObjectMeta, px.Spec.GetSecrets(), ref); err != nil {
			return errors.Wrap(err, "error in wiping out database.")
		}
	} else {
		// Make sure snapshot and secret's ownerreference is removed.
		if err := dynamic_util.RemoveOwnerReferenceForSelector(
			c.DynamicClient,
			api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
			px.Namespace,
			selector,
			ref); err != nil {
			return err
		}
		if err := dynamic_util.RemoveOwnerReferenceForItems(
			c.DynamicClient,
			core.SchemeGroupVersion.WithResource("secrets"),
			px.Namespace,
			px.Spec.GetSecrets(),
			ref); err != nil {
			return err
		}
	}
	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		px.Namespace,
		selector,
		ref)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(px *api.PerconaXtraDB, ref *core.ObjectReference) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(px.OffshootSelectors())

	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		api.SchemeGroupVersion.WithResource(api.ResourcePluralSnapshot),
		px.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForSelector(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		px.Namespace,
		labelSelector,
		ref); err != nil {
		return err
	}
	if err := dynamic_util.RemoveOwnerReferenceForItems(
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("secrets"),
		px.Namespace,
		px.Spec.GetSecrets(),
		ref); err != nil {
		return err
	}
	return nil
}
