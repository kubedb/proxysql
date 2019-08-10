package controller

import (
	"fmt"

	batch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core_util "kmodules.xyz/client-go/core/v1"
	"kubedb.dev/apimachinery/apis"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	amv "kubedb.dev/apimachinery/pkg/validator"
)

func (c *Controller) GetDatabase(meta metav1.ObjectMeta) (runtime.Object, error) {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return nil, err
	}

	return px, nil
}

func (c *Controller) SetDatabaseStatus(meta metav1.ObjectMeta, phase api.DatabasePhase, reason string) error {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}
	_, err = util.UpdatePerconaXtraDBStatus(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDBStatus) *api.PerconaXtraDBStatus {
		in.Phase = phase
		in.Reason = reason
		return in
	}, apis.EnableStatusSubresource)
	return err
}

func (c *Controller) UpsertDatabaseAnnotation(meta metav1.ObjectMeta, annotation map[string]string) error {
	px, err := c.pxLister.PerconaXtraDBs(meta.Namespace).Get(meta.Name)
	if err != nil {
		return err
	}

	_, _, err = util.PatchPerconaXtraDB(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
		in.Annotations = core_util.UpsertMap(in.Annotations, annotation)
		return in
	})
	return err
}

func (c *Controller) ValidateSnapshot(snapshot *api.Snapshot) error {
	// Database name can't empty
	databaseName := snapshot.Spec.DatabaseName
	if databaseName == "" {
		return fmt.Errorf(`object 'DatabaseName' is missing in '%v'`, snapshot.Spec)
	}

	if _, err := c.pxLister.PerconaXtraDBs(snapshot.Namespace).Get(databaseName); err != nil {
		return err
	}

	return amv.ValidateSnapshotSpec(snapshot.Spec.Backend)
}

func (c *Controller) GetSnapshotter(snapshot *api.Snapshot) (*batch.Job, error) {
	return nil, nil
}

func (c *Controller) WipeOutSnapshot(snapshot *api.Snapshot) error {
	// wipeOut not possible for local backend.
	// Ref: https://github.com/kubedb/project/issues/261
	if snapshot.Spec.Local != nil {
		return nil
	}
	return c.DeleteSnapshotData(snapshot)
}
