package framework

import (
	"fmt"

	"github.com/appscode/go/types"
	"github.com/appscode/go/wait"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	appcat_api "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/controller"
	"stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashV1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashv1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
	"stash.appscode.dev/stash/pkg/util"
)

var (
	StashPerconaXtraDBBackupTask  = "proxysql-backup-5.7"
	StashPerconaXtraDBRestoreTask = "proxysql-restore-5.7"
)

func (f *Framework) FoundStashCRDs() bool {
	return controller.FoundStashCRDs(f.apiExtKubeClient)
}

func (f *Invocation) BackupConfiguration(meta metav1.ObjectMeta) *stashv1beta1.BackupConfiguration {
	return &stashv1beta1.BackupConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
		},
		Spec: stashv1beta1.BackupConfigurationSpec{
			Task: stashv1beta1.TaskRef{
				Name: StashPerconaXtraDBBackupTask,
			},
			Repository: corev1.LocalObjectReference{
				Name: meta.Name,
			},
			Schedule: "*/3 * * * *",
			Target: &stashv1beta1.BackupTarget{
				Ref: stashv1beta1.TargetRef{
					APIVersion: appcat_api.SchemeGroupVersion.String(),
					Kind:       appcat_api.ResourceKindApp,
					Name:       meta.Name,
				},
			},
			RetentionPolicy: v1alpha1.RetentionPolicy{
				KeepLast: 5,
				Prune:    true,
			},
		},
	}
}

func (f *Framework) CreateBackupConfiguration(backupCfg *stashv1beta1.BackupConfiguration) error {
	_, err := f.stashClient.StashV1beta1().BackupConfigurations(backupCfg.Namespace).Create(backupCfg)
	return err
}

func (f *Framework) DeleteBackupConfiguration(meta metav1.ObjectMeta) error {
	return f.stashClient.StashV1beta1().BackupConfigurations(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) WaitUntilBackkupSessionBeCreated(bcMeta metav1.ObjectMeta) (bs *stashv1beta1.BackupSession, err error) {
	err = wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		bsList, err := f.stashClient.StashV1beta1().BackupSessions(bcMeta.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set{
				util.LabelBackupConfiguration: bcMeta.Name,
			}.String(),
		})
		if err != nil {
			return false, err
		}
		if len(bsList.Items) == 0 {
			return false, nil
		}

		bs = &bsList.Items[0]

		return true, nil
	})

	return
}

func (f *Framework) EventuallyBackupSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() (phase stashv1beta1.BackupSessionPhase) {
			bs, err := f.stashClient.StashV1beta1().BackupSessions(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return bs.Status.Phase
		},
	)
}

func (f *Invocation) Repository(meta metav1.ObjectMeta, secretName string) *stashV1alpha1.Repository {
	return &stashV1alpha1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
		},
	}
}

func (f *Framework) CreateRepository(repo *stashV1alpha1.Repository) error {
	_, err := f.stashClient.StashV1alpha1().Repositories(repo.Namespace).Create(repo)
	return err
}

func (f *Framework) DeleteRepository(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1alpha1().Repositories(meta.Namespace).Delete(meta.Name, deleteInBackground())
	return err
}

func (f *Invocation) RestoreSession(meta, oldMeta metav1.ObjectMeta, replicas *int32) *stashv1beta1.RestoreSession {
	return &stashv1beta1.RestoreSession{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: f.namespace,
			Labels: map[string]string{
				"app":                 f.app,
				api.LabelDatabaseKind: api.ResourceKindPerconaXtraDB,
			},
		},
		Spec: stashv1beta1.RestoreSessionSpec{
			Task: stashv1beta1.TaskRef{
				Name: StashPerconaXtraDBRestoreTask,
			},
			Repository: corev1.LocalObjectReference{
				Name: oldMeta.Name,
			},
			Rules: []stashv1beta1.Rule{
				{
					Snapshots:   []string{"latest"},
					TargetHosts: []string{},
					SourceHost:  "host-0",
				},
			},
			Target: &stashv1beta1.RestoreTarget{
				Replicas: replicas,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      fmt.Sprintf("data-%s", meta.Name),
						MountPath: "/var/lib/mysql",
					},
				},
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("data-%s-${POD_ORDINAL}", meta.Name),
							Annotations: map[string]string{
								"volume.beta.kubernetes.io/storage-class": "standard",
							},
							CreationTimestamp: metav1.Now(),
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							StorageClassName: types.StringP("standard"),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse(DBPvcStorageSize),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (f *Framework) CreateRestoreSession(restoreSession *stashv1beta1.RestoreSession) error {
	_, err := f.stashClient.StashV1beta1().RestoreSessions(restoreSession.Namespace).Create(restoreSession)
	return err
}

func (f *Framework) DeleteRestoreSession(meta metav1.ObjectMeta) error {
	err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
	return err
}

func (f *Framework) EventuallyRestoreSessionPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(func() stashv1beta1.RestoreSessionPhase {
		restoreSession, err := f.stashClient.StashV1beta1().RestoreSessions(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		return restoreSession.Status.Phase
	})
}
