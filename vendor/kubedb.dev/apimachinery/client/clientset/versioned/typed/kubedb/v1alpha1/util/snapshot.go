/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/golang/glog"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	kutil "kmodules.xyz/client-go"
)

func CreateOrPatchSnapshot(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.Snapshot) *api.Snapshot) (*api.Snapshot, kutil.VerbType, error) {
	cur, err := c.Snapshots(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	if kerr.IsNotFound(err) {
		glog.V(3).Infof("Creating Snapshot %s/%s.", meta.Namespace, meta.Name)
		out, err := c.Snapshots(meta.Namespace).Create(transform(&api.Snapshot{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Snapshot",
				APIVersion: api.SchemeGroupVersion.String(),
			},
			ObjectMeta: meta,
		}))
		return out, kutil.VerbCreated, err
	} else if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	return PatchSnapshot(c, cur, transform)
}

func PatchSnapshot(c cs.KubedbV1alpha1Interface, cur *api.Snapshot, transform func(*api.Snapshot) *api.Snapshot) (*api.Snapshot, kutil.VerbType, error) {
	return PatchSnapshotObject(c, cur, transform(cur.DeepCopy()))
}

func PatchSnapshotObject(c cs.KubedbV1alpha1Interface, cur, mod *api.Snapshot) (*api.Snapshot, kutil.VerbType, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	modJson, err := json.Marshal(mod)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}

	patch, err := jsonpatch.CreateMergePatch(curJson, modJson)
	if err != nil {
		return nil, kutil.VerbUnchanged, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, kutil.VerbUnchanged, nil
	}
	glog.V(3).Infof("Patching Snapshot %s/%s with %s.", cur.Namespace, cur.Name, string(patch))
	out, err := c.Snapshots(cur.Namespace).Patch(cur.Name, types.MergePatchType, patch)
	return out, kutil.VerbPatched, err
}

func TryUpdateSnapshot(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta, transform func(*api.Snapshot) *api.Snapshot) (result *api.Snapshot, err error) {
	attempt := 0
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		cur, e2 := c.Snapshots(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if kerr.IsNotFound(e2) {
			return false, e2
		} else if e2 == nil {
			result, e2 = c.Snapshots(cur.Namespace).Update(transform(cur.DeepCopy()))
			return e2 == nil, nil
		}
		glog.Errorf("Attempt %d failed to update Snapshot %s/%s due to %v.", attempt, cur.Namespace, cur.Name, e2)
		return false, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update Snapshot %s/%s after %d attempts due to %v", meta.Namespace, meta.Name, attempt, err)
	}
	return
}

func WaitUntilSnapshotCompletion(c cs.KubedbV1alpha1Interface, meta metav1.ObjectMeta) (result *api.Snapshot, err error) {
	err = wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		result, err = c.Snapshots(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if result.Status.CompletionTime != nil {
			return true, nil
		}
		return false, nil
	})
	return
}

func UpdateSnapshotStatus(
	c cs.KubedbV1alpha1Interface,
	in *api.Snapshot,
	transform func(*api.SnapshotStatus) *api.SnapshotStatus,
) (result *api.Snapshot, err error) {
	apply := func(x *api.Snapshot) *api.Snapshot {
		return &api.Snapshot{
			TypeMeta:   x.TypeMeta,
			ObjectMeta: x.ObjectMeta,
			Spec:       x.Spec,
			Status:     *transform(in.Status.DeepCopy()),
		}
	}

	attempt := 0
	cur := in.DeepCopy()
	err = wait.PollImmediate(kutil.RetryInterval, kutil.RetryTimeout, func() (bool, error) {
		attempt++
		var e2 error
		result, e2 = c.Snapshots(in.Namespace).UpdateStatus(apply(cur))
		if kerr.IsConflict(e2) {
			latest, e3 := c.Snapshots(in.Namespace).Get(in.Name, metav1.GetOptions{})
			switch {
			case e3 == nil:
				cur = latest
				return false, nil
			case kutil.IsRequestRetryable(e3):
				return false, nil
			default:
				return false, e3
			}
		} else if err != nil && !kutil.IsRequestRetryable(e2) {
			return false, e2
		}
		return e2 == nil, nil
	})

	if err != nil {
		err = fmt.Errorf("failed to update status of Snapshot %s/%s after %d attempts due to %v", in.Namespace, in.Name, attempt, err)
	}
	return
}

func MarkAsFailedSnapshot(
	c cs.KubedbV1alpha1Interface,
	cur *api.Snapshot,
	reason string,
) (*api.Snapshot, error) {
	return UpdateSnapshotStatus(c, cur, func(in *api.SnapshotStatus) *api.SnapshotStatus {
		t := metav1.Now()
		in.CompletionTime = &t
		in.Phase = api.SnapshotPhaseFailed
		in.Reason = reason
		return in
	})
}
