package e2e_test

import (
	"fmt"
	"os"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_util "kmodules.xyz/client-go/meta"
	store "kmodules.xyz/objectstore-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/percona-xtradb/test/e2e/framework"
	"kubedb.dev/percona-xtradb/test/e2e/matcher"
)

const (
	S3_BUCKET_NAME       = "S3_BUCKET_NAME"
	GCS_BUCKET_NAME      = "GCS_BUCKET_NAME"
	AZURE_CONTAINER_NAME = "AZURE_CONTAINER_NAME"
	SWIFT_CONTAINER_NAME = "SWIFT_CONTAINER_NAME"
	MYSQL_DATABASE       = "MYSQL_DATABASE"
	MYSQL_ROOT_PASSWORD  = "MYSQL_ROOT_PASSWORD"
)

var _ = Describe("PerconaXtraDB", func() {
	var (
		err                  error
		f                    *framework.Invocation
		perconaxtradb        *api.PerconaXtraDB
		garbagePerconaXtraDB *api.PerconaXtraDBList
		snapshot             *api.Snapshot
		secret               *core.Secret
		skipMessage          string
		skipDataChecking     bool
		dbName               string
	)

	BeforeEach(func() {
		f = root.Invoke()
		perconaxtradb = f.PerconaXtraDB()
		garbagePerconaXtraDB = new(api.PerconaXtraDBList)
		snapshot = f.Snapshot()
		skipMessage = ""
		skipDataChecking = true
		dbName = "mysql"
	})

	var createAndWaitForRunning = func() {
		By("Create PerconaXtraDB: " + perconaxtradb.Name)
		err = f.CreatePerconaXtraDB(perconaxtradb)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running PerconaXtraDB")
		f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(perconaxtradb.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(perconaxtradb.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())
	}

	var testGeneralBehaviour = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}
		// Create PerconaXtraDB
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())

		By("Inserting Rows")
		f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

		By("Delete PerconaXtraDB")
		err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for PerconaXtraDB to be paused")
		f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

		// Create PerconaXtraDB object again to resume it
		By("Create PerconaXtraDB: " + perconaxtradb.Name)
		err = f.CreatePerconaXtraDB(perconaxtradb)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for DormantDatabase to be deleted")
		f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

		By("Wait for Running PerconaXtraDB")
		f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

	}

	var shouldTakeSnapshot = func() {
		// Create and wait for running PerconaXtraDB
		createAndWaitForRunning()

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeed snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var shouldInsertDataAndTakeSnapshot = func() {
		// Create and wait for running PerconaXtraDB
		createAndWaitForRunning()

		By("Creating Table")
		f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())

		By("Inserting Row")
		f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())

		By("Checking Row Count of Table")
		f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

		By("Create Secret")
		err := f.CreateSecret(secret)
		Expect(err).NotTo(HaveOccurred())

		By("Create Snapshot")
		err = f.CreateSnapshot(snapshot)
		Expect(err).NotTo(HaveOccurred())

		By("Check for Succeed snapshot")
		f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

		if !skipDataChecking {
			By("Check for snapshot data")
			f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
		}
	}

	var deleteTestResource = func() {
		if perconaxtradb == nil {
			log.Infoln("Skipping cleanup. Reason: PerconaXtraDB object is nil")
			return
		}

		By("Check if PerconaXtraDB " + perconaxtradb.Name + " exists.")
		pc, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// PerconaXtraDB was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete PerconaXtraDB")
		err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: PerconaXtraDB does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if pc.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for PerconaXtraDB to be paused")
			f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut PerconaXtraDB")
			_, err := f.PatchDormantDatabase(perconaxtradb.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(perconaxtradb.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for PerconaXtraDB resources to be wipedOut")
		f.EventuallyWipedOut(perconaxtradb.ObjectMeta).Should(Succeed())
	}

	var deleteSnapshot = func() {

		By("Deleting Snapshot: " + snapshot.Name)
		err = f.DeleteSnapshot(snapshot.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		if !skipDataChecking {
			// do not try to check snapshot data if secret does not exist
			_, err = f.GetSecret(secret.ObjectMeta)
			if err != nil && kerr.IsNotFound(err) {
				log.Infof("Skipping checking snapshot data. Reason: secret %s not found", secret.Name)
				return
			}
			Expect(err).NotTo(HaveOccurred())

			By("Checking Snapshot's data wiped out from backend")
			f.EventuallySnapshotDataFound(snapshot).Should(BeFalse())
		}
	}

	AfterEach(func() {
		// delete resources for current PerconaXtraDB
		deleteTestResource()

		// old PerconaXtraDB are in garbagePerconaXtraDB list. delete their resources.
		for _, pc := range garbagePerconaXtraDB.Items {
			*perconaxtradb = pc
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers()
	})

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", testGeneralBehaviour)
			})
		})

		Context("Snapshot", func() {

			BeforeEach(func() {
				skipDataChecking = false
				snapshot.Spec.DatabaseName = perconaxtradb.Name
			})

			AfterEach(func() {
				// delete snapshot and check for data wipeOut
				deleteSnapshot()

				By("Deleting secret: " + secret.Name)
				err := f.DeleteSecret(secret.ObjectMeta)
				if err != nil && !kerr.IsNotFound(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			})

			Context("In Local", func() {

				BeforeEach(func() {
					skipDataChecking = true
					secret = f.SecretForLocalBackend()
					snapshot.Spec.StorageSecretName = secret.Name
				})

				Context("With EmptyDir as Snapshot's backend", func() {
					BeforeEach(func() {
						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								EmptyDir: &core.EmptyDirVolumeSource{},
							},
						}
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("With PVC as Snapshot's backend", func() {
					var snapPVC *core.PersistentVolumeClaim

					BeforeEach(func() {
						snapPVC = f.GetPersistentVolumeClaim()
						err := f.CreatePersistentVolumeClaim(snapPVC)
						Expect(err).NotTo(HaveOccurred())

						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
									ClaimName: snapPVC.Name,
								},
							},
						}
					})

					AfterEach(func() {
						err := f.DeletePersistentVolumeClaim(snapPVC.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should delete Snapshot successfully", func() {
						shouldTakeSnapshot()

						By("Deleting Snapshot")
						err := f.DeleteSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for Snapshot to be deleted")
						f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())
					})
				})
			})

			Context("In S3", func() {
				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &store.S3Spec{
						Bucket: os.Getenv(S3_BUCKET_NAME),
					}
				})

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)

				Context("faulty snapshot", func() {
					BeforeEach(func() {
						skipDataChecking = true
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.S3 = &store.S3Spec{
							Bucket: "nonexisting",
						}
					})

					It("snapshot should fail", func() {
						// Create and wait for running db
						createAndWaitForRunning()

						By("Create Secret")
						err := f.CreateSecret(secret)
						Expect(err).NotTo(HaveOccurred())

						By("Create Snapshot")
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for failed snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseFailed))
					})
				})

				Context("Delete One Snapshot keeping others", func() {
					BeforeEach(func() {
						perconaxtradb.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create PerconaXtraDB and take Snapshot
						shouldTakeSnapshot()

						oldSnapshot := snapshot.DeepCopy()

						// New snapshot that has old snapshot's name in prefix
						snapshot.Name += "-2"

						By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataChecking {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						// delete old snapshot
						By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
						err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for old Snapshot to be deleted")
						f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
						if !skipDataChecking {
							By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
							f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
						}

						// check remaining snapshot
						By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataChecking {
							By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}
					})
				})
			})

			Context("In GCS", func() {
				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
				})

				Context("Without Init", func() {
					It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
				})

				Context("With Init", func() {
					BeforeEach(func() {
						perconaxtradb.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("should take Snapshot successfully", shouldTakeSnapshot)
				})

				Context("Delete One Snapshot keeping others", func() {
					BeforeEach(func() {
						perconaxtradb.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create PerconaXtraDB and take Snapshot
						shouldTakeSnapshot()

						oldSnapshot := snapshot.DeepCopy()

						// New snapshot that has old snapshot's name in prefix
						snapshot.Name += "-2"

						By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataChecking {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						// delete old snapshot
						By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
						err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for old Snapshot to be deleted")
						f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
						if !skipDataChecking {
							By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
							f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
						}

						// check remaining snapshot
						By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataChecking {
							By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}
					})
				})

			})

			Context("In Azure", func() {
				BeforeEach(func() {
					secret = f.SecretForAzureBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Azure = &store.AzureSpec{
						Container: os.Getenv(AZURE_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)

				Context("Delete One Snapshot keeping others", func() {
					BeforeEach(func() {
						perconaxtradb.Spec.Init = &api.InitSpec{
							ScriptSource: &api.ScriptSourceSpec{
								VolumeSource: core.VolumeSource{
									GitRepo: &core.GitRepoVolumeSource{
										Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
										Directory:  ".",
									},
								},
							},
						}
					})

					It("Delete One Snapshot keeping others", func() {
						// Create PerconaXtraDB and take Snapshot
						shouldTakeSnapshot()

						oldSnapshot := snapshot.DeepCopy()

						// New snapshot that has old snapshot's name in prefix
						snapshot.Name += "-2"

						By(fmt.Sprintf("Create Snapshot %v", snapshot.Name))
						err = f.CreateSnapshot(snapshot)
						Expect(err).NotTo(HaveOccurred())

						By("Check for Succeeded snapshot")
						f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

						if !skipDataChecking {
							By("Check for snapshot data")
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}

						// delete old snapshot
						By(fmt.Sprintf("Delete old Snapshot %v", oldSnapshot.Name))
						err = f.DeleteSnapshot(oldSnapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Waiting for old Snapshot to be deleted")
						f.EventuallySnapshot(oldSnapshot.ObjectMeta).Should(BeFalse())
						if !skipDataChecking {
							By(fmt.Sprintf("Check data for old snapshot %v", oldSnapshot.Name))
							f.EventuallySnapshotDataFound(oldSnapshot).Should(BeFalse())
						}

						// check remaining snapshot
						By(fmt.Sprintf("Checking another Snapshot %v still exists", snapshot.Name))
						_, err = f.GetSnapshot(snapshot.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						if !skipDataChecking {
							By(fmt.Sprintf("Check data for remaining snapshot %v", snapshot.Name))
							f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
						}
					})
				})
			})

			Context("In Swift", func() {
				BeforeEach(func() {
					secret = f.SecretForSwiftBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.Swift = &store.SwiftSpec{
						Container: os.Getenv(SWIFT_CONTAINER_NAME),
					}
				})

				It("should take Snapshot successfully", shouldInsertDataAndTakeSnapshot)
			})

			Context("Snapshot PodVolume Template - In S3", func() {

				BeforeEach(func() {
					secret = f.SecretForS3Backend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.S3 = &store.S3Spec{
						Bucket: os.Getenv(S3_BUCKET_NAME),
					}
				})

				var shouldHandleJobVolumeSuccessfully = func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Get PerconaXtraDB")
					es, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					perconaxtradb.Spec = es.Spec

					By("Create Secret")
					err = f.CreateSecret(secret)
					Expect(err).NotTo(HaveOccurred())

					// determine pvcSpec and storageType for job
					// start
					pvcSpec := snapshot.Spec.PodVolumeClaimSpec
					if pvcSpec == nil {
						pvcSpec = perconaxtradb.Spec.Storage
					}
					st := snapshot.Spec.StorageType
					if st == nil {
						st = &perconaxtradb.Spec.StorageType
					}
					Expect(st).NotTo(BeNil())
					// end

					By("Create Snapshot")
					err = f.CreateSnapshot(snapshot)
					if *st == api.StorageTypeDurable && pvcSpec == nil {
						By("Create Snapshot should have failed")
						Expect(err).Should(HaveOccurred())
						return
					} else {
						Expect(err).NotTo(HaveOccurred())
					}

					By("Get Snapshot")
					snap, err := f.GetSnapshot(snapshot.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					snapshot.Spec = snap.Spec

					if *st == api.StorageTypeEphemeral {
						storageSize := "0"
						if pvcSpec != nil {
							if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
								storageSize = sz.String()
							}
						}
						By(fmt.Sprintf("Check for Job Empty volume size: %v", storageSize))
						f.EventuallyJobVolumeEmptyDirSize(snapshot.ObjectMeta).Should(Equal(storageSize))
					} else if *st == api.StorageTypeDurable {
						sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]
						Expect(found).NotTo(BeFalse())

						By("Check for Job PVC Volume size from snapshot")
						f.EventuallyJobPVCSize(snapshot.ObjectMeta).Should(Equal(sz.String()))
					}

					By("Check for succeeded snapshot")
					f.EventuallySnapshotPhase(snapshot.ObjectMeta).Should(Equal(api.SnapshotPhaseSucceeded))

					if !skipDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}
				}

				// db StorageType Scenarios
				// ==============> Start
				var dbStorageTypeScenarios = func() {
					Context("DBStorageType - Durable", func() {
						BeforeEach(func() {
							perconaxtradb.Spec.StorageType = api.StorageTypeDurable
							perconaxtradb.Spec.Storage = &core.PersistentVolumeClaimSpec{
								Resources: core.ResourceRequirements{
									Requests: core.ResourceList{
										core.ResourceStorage: resource.MustParse(framework.DBPvcStorageSize),
									},
								},
								StorageClassName: types.StringP(root.StorageClass),
							}

						})

						It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
					})

					Context("DBStorageType - Ephemeral", func() {
						BeforeEach(func() {
							perconaxtradb.Spec.StorageType = api.StorageTypeEphemeral
							perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
						})

						Context("DBPvcSpec is nil", func() {
							BeforeEach(func() {
								perconaxtradb.Spec.Storage = nil
							})

							It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
						})

						Context("DBPvcSpec is given [not nil]", func() {
							BeforeEach(func() {
								perconaxtradb.Spec.Storage = &core.PersistentVolumeClaimSpec{
									Resources: core.ResourceRequirements{
										Requests: core.ResourceList{
											core.ResourceStorage: resource.MustParse(framework.DBPvcStorageSize),
										},
									},
									StorageClassName: types.StringP(root.StorageClass),
								}
							})

							It("should Handle Job Volume Successfully", shouldHandleJobVolumeSuccessfully)
						})
					})
				}
				// End <==============

				// Snapshot PVC Scenarios
				// ==============> Start
				var snapshotPvcScenarios = func() {
					Context("Snapshot PVC is given [not nil]", func() {
						BeforeEach(func() {
							snapshot.Spec.PodVolumeClaimSpec = &core.PersistentVolumeClaimSpec{
								Resources: core.ResourceRequirements{
									Requests: core.ResourceList{
										core.ResourceStorage: resource.MustParse(framework.JobPvcStorageSize),
									},
								},
								StorageClassName: types.StringP(root.StorageClass),
							}
						})

						dbStorageTypeScenarios()
					})

					Context("Snapshot PVC is nil", func() {
						BeforeEach(func() {
							snapshot.Spec.PodVolumeClaimSpec = nil
						})

						dbStorageTypeScenarios()
					})
				}
				// End <==============

				Context("Snapshot StorageType is nil", func() {
					BeforeEach(func() {
						snapshot.Spec.StorageType = nil
					})

					snapshotPvcScenarios()
				})

				Context("Snapshot StorageType is Ephemeral", func() {
					BeforeEach(func() {
						ephemeral := api.StorageTypeEphemeral
						snapshot.Spec.StorageType = &ephemeral
					})

					snapshotPvcScenarios()
				})

				Context("Snapshot StorageType is Durable", func() {
					BeforeEach(func() {
						durable := api.StorageTypeDurable
						snapshot.Spec.StorageType = &durable
					})

					snapshotPvcScenarios()
				})
			})
		})

		Context("Initialize", func() {

			Context("With Script", func() {
				BeforeEach(func() {
					perconaxtradb.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should run successfully", func() {
					// Create PerconaXtraDB
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
				})
			})

			Context("With Snapshot", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				var shouldInitializeFromSnapshot = func() {
					// Create PerconaXtraDB and take Snapshot
					shouldInsertDataAndTakeSnapshot()

					oldPerconaXtraDB, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					garbagePerconaXtraDB.Items = append(garbagePerconaXtraDB.Items, *oldPerconaXtraDB)

					By("Create PerconaXtraDB from snapshot")
					perconaxtradb = f.PerconaXtraDB()
					perconaxtradb.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating init Snapshot Mysql without secret name" + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					perconaxtradb.Spec.DatabaseSecret = oldPerconaXtraDB.Spec.DatabaseSecret

					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
				}

				Context("From Local backend", func() {
					var snapPVC *core.PersistentVolumeClaim

					BeforeEach(func() {

						skipDataChecking = true
						snapPVC = f.GetPersistentVolumeClaim()
						err := f.CreatePersistentVolumeClaim(snapPVC)
						Expect(err).NotTo(HaveOccurred())

						secret = f.SecretForLocalBackend()
						snapshot.Spec.DatabaseName = perconaxtradb.Name
						snapshot.Spec.StorageSecretName = secret.Name

						snapshot.Spec.Local = &store.LocalSpec{
							MountPath: "/repo",
							VolumeSource: core.VolumeSource{
								PersistentVolumeClaim: &core.PersistentVolumeClaimVolumeSource{
									ClaimName: snapPVC.Name,
								},
							},
						}
					})

					AfterEach(func() {
						err := f.DeletePersistentVolumeClaim(snapPVC.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
					})

					It("should initialize successfully", shouldInitializeFromSnapshot)
				})

				Context("From GCS backend", func() {

					BeforeEach(func() {

						skipDataChecking = false
						secret = f.SecretForGCSBackend()
						snapshot.Spec.StorageSecretName = secret.Name
						snapshot.Spec.DatabaseName = perconaxtradb.Name

						snapshot.Spec.GCS = &store.GCSSpec{
							Bucket: os.Getenv(GCS_BUCKET_NAME),
						}
					})

					It("should initialize successfully", shouldInitializeFromSnapshot)
				})
			})
		})

		Context("Resume", func() {

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for PerconaXtraDB to be paused")
					f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

					// Create PerconaXtraDB object again to resume it
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					// Delete without caring if DB is resumed
					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for PerconaXtraDB to be deleted")
					f.EventuallyPerconaXtraDB(perconaxtradb.ObjectMeta).Should(BeFalse())

					// Create PerconaXtraDB object again to resume it
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Wait for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
				})
			})

			Context("Without Init", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Creating Table")
					f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())

					By("Inserting Row")
					f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for PerconaXtraDB to be paused")
					f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

					// Create PerconaXtraDB object again to resume it
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Wait for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
				})
			})

			Context("with init Script", func() {
				BeforeEach(func() {
					perconaxtradb.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for PerconaXtraDB to be paused")
					f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

					// Create PerconaXtraDB object again to resume it
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Wait for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					perconaxtradb, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(perconaxtradb.Spec.Init).NotTo(BeNil())

					By("Checking PerconaXtraDB crd does not have kubedb.com/initialized annotation")
					_, err = meta_util.GetString(perconaxtradb.Annotations, api.AnnotationInitialized)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("With Snapshot Init", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				BeforeEach(func() {
					skipDataChecking = false
					secret = f.SecretForGCSBackend()
					snapshot.Spec.StorageSecretName = secret.Name
					snapshot.Spec.GCS = &store.GCSSpec{
						Bucket: os.Getenv(GCS_BUCKET_NAME),
					}
					snapshot.Spec.DatabaseName = perconaxtradb.Name
				})

				It("should resume successfully", func() {
					// Create PerconaXtraDB and take Snapshot
					shouldInsertDataAndTakeSnapshot()

					oldPerconaXtraDB, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					garbagePerconaXtraDB.Items = append(garbagePerconaXtraDB.Items, *oldPerconaXtraDB)

					By("Create PerconaXtraDB from snapshot")
					perconaxtradb = f.PerconaXtraDB()
					perconaxtradb.Spec.Init = &api.InitSpec{
						SnapshotSource: &api.SnapshotSourceSpec{
							Namespace: snapshot.Namespace,
							Name:      snapshot.Name,
						},
					}

					By("Creating PerconaXtraDB without secret name to init from Snapshot: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).Should(HaveOccurred())

					// for snapshot init, user have to use older secret,
					// because the username & password  will be replaced to
					perconaxtradb.Spec.DatabaseSecret = oldPerconaXtraDB.Spec.DatabaseSecret

					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for PerconaXtraDB to be paused")
					f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

					// Create PerconaXtraDB object again to resume it
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Wait for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					perconaxtradb, err = f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
					Expect(perconaxtradb.Spec.Init).ShouldNot(BeNil())

					By("Checking PerconaXtraDB has kubedb.com/initialized annotation")
					_, err = meta_util.GetString(perconaxtradb.Annotations, api.AnnotationInitialized)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Multiple times with init", func() {

				BeforeEach(func() {
					perconaxtradb.Spec.Init = &api.InitSpec{
						ScriptSource: &api.ScriptSourceSpec{
							VolumeSource: core.VolumeSource{
								GitRepo: &core.GitRepoVolumeSource{
									Repository: "https://kubedb.dev/percona-xtradb-init-scripts.git",
									Directory:  ".",
								},
							},
						},
					}
				})

				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Checking Row Count of Table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")

						By("Delete PerconaXtraDB")
						err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for PerconaXtraDB to be paused")
						f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

						// Create PerconaXtraDB object again to resume it
						By("Create PerconaXtraDB: " + perconaxtradb.Name)
						err = f.CreatePerconaXtraDB(perconaxtradb)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

						By("Wait for Running PerconaXtraDB")
						f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

						By("Checking Row Count of Table")
						f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))

						perconaxtradb, err := f.GetPerconaXtraDB(perconaxtradb.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
						Expect(perconaxtradb.Spec.Init).ShouldNot(BeNil())

						By("Checking PerconaXtraDB crd does not have kubedb.com/initialized annotation")
						_, err = meta_util.GetString(perconaxtradb.Annotations, api.AnnotationInitialized)
						Expect(err).To(HaveOccurred())
					}
				})
			})
		})

		//Context("SnapshotScheduler", func() {
		//
		//	BeforeEach(func() {
		//		skipDataChecking = false
		//	})
		//
		//	AfterEach(func() {
		//		snapshotList, err := f.GetSnapshotList(perconaxtradb.ObjectMeta)
		//		Expect(err).NotTo(HaveOccurred())
		//
		//		for _, snap := range snapshotList.Items {
		//			snapshot = &snap
		//
		//			// delete snapshot and check for data wipeOut
		//			deleteSnapshot()
		//		}
		//
		//		By("Deleting secret: " + secret.Name)
		//		err = f.DeleteSecret(secret.ObjectMeta)
		//		if err != nil && !kerr.IsNotFound(err) {
		//			Expect(err).NotTo(HaveOccurred())
		//		}
		//	})
		//
		//	Context("With Startup", func() {
		//
		//		var shouldStartupSchedular = func() {
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			// Create and wait for running PerconaXtraDB
		//			createAndWaitForRunning()
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(perconaxtradb.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Remove Backup Scheduler from PerconaXtraDB")
		//			_, err = f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB{
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(perconaxtradb.ObjectMeta).Should(Succeed())
		//		}
		//
		//		Context("with local", func() {
		//			BeforeEach(func() {
		//				skipDataChecking = true
		//				secret = f.SecretForLocalBackend()
		//				perconaxtradb.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//			})
		//
		//			It("should run scheduler successfully", shouldStartupSchedular)
		//		})
		//
		//		Context("with GCS", func() {
		//			BeforeEach(func() {
		//				secret = f.SecretForGCSBackend()
		//				perconaxtradb.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 1m",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						GCS: &store.GCSSpec{
		//							Bucket: os.Getenv(GCS_BUCKET_NAME),
		//						},
		//					},
		//				}
		//			})
		//
		//			It("should run scheduler successfully", shouldStartupSchedular)
		//		})
		//	})
		//
		//	Context("With Update - with Local", func() {
		//
		//		BeforeEach(func() {
		//			skipDataChecking = true
		//			secret = f.SecretForLocalBackend()
		//		})
		//
		//		It("should run scheduler successfully", func() {
		//			// Create and wait for running PerconaXtraDB
		//			createAndWaitForRunning()
		//
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Update PerconaXtraDB")
		//			_, err = f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB{
		//				in.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(perconaxtradb.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Remove Backup Scheduler from PerconaXtraDB")
		//			_, err = f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB{
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(perconaxtradb.ObjectMeta).Should(Succeed())
		//		})
		//	})
		//
		//	Context("Re-Use DormantDatabase's scheduler", func() {
		//
		//		BeforeEach(func() {
		//			skipDataChecking = true
		//			secret = f.SecretForLocalBackend()
		//		})
		//
		//		It("should re-use scheduler successfully", func() {
		//			// Create and wait for running PerconaXtraDB
		//			createAndWaitForRunning()
		//
		//			By("Create Secret")
		//			err := f.CreateSecret(secret)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Update PerconaXtraDB")
		//			_, err = f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB{
		//				in.Spec.BackupSchedule = &api.BackupScheduleSpec{
		//					CronExpression: "@every 20s",
		//					Backend: store.Backend{
		//						StorageSecretName: secret.Name,
		//						Local: &store.LocalSpec{
		//							MountPath: "/repo",
		//							VolumeSource: core.VolumeSource{
		//								EmptyDir: &core.EmptyDirVolumeSource{},
		//							},
		//						},
		//					},
		//				}
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Creating Table")
		//			f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())
		//
		//			By("Inserting Row")
		//			f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())
		//
		//			By("Checking Row Count of Table")
		//			f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(perconaxtradb.ObjectMeta).Should(matcher.MoreThan(3))
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(perconaxtradb.ObjectMeta).Should(Succeed())
		//
		//			By("Delete PerconaXtraDB")
		//			err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Wait for PerconaXtraDB to be paused")
		//			f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())
		//
		//			// Create PerconaXtraDB object again to resume it
		//			By("Create PerconaXtraDB: " + perconaxtradb.Name)
		//			err = f.CreatePerconaXtraDB(perconaxtradb)
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Wait for DormantDatabase to be deleted")
		//			f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())
		//
		//			By("Wait for Running PerconaXtraDB")
		//			f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())
		//
		//			By("Checking Row Count of Table")
		//			f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
		//
		//			By("Count multiple Snapshot Object")
		//			f.EventuallySnapshotCount(perconaxtradb.ObjectMeta).Should(matcher.MoreThan(5))
		//
		//			By("Remove Backup Scheduler from PerconaXtraDB")
		//			_, err = f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB{
		//				in.Spec.BackupSchedule = nil
		//				return in
		//			})
		//			Expect(err).NotTo(HaveOccurred())
		//
		//			By("Verify multiple Succeeded Snapshot")
		//			f.EventuallyMultipleSnapshotFinishedProcessing(perconaxtradb.ObjectMeta).Should(Succeed())
		//		})
		//	})
		//})

		Context("Termination Policy", func() {

			BeforeEach(func() {
				skipDataChecking = false
				secret = f.SecretForGCSBackend()
				snapshot.Spec.StorageSecretName = secret.Name
				snapshot.Spec.GCS = &store.GCSSpec{
					Bucket: os.Getenv(GCS_BUCKET_NAME),
				}
				snapshot.Spec.DatabaseName = perconaxtradb.Name
			})

			Context("with TerminationDoNotTerminate", func() {
				BeforeEach(func() {
					skipDataChecking = true
					perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running PerconaXtraDB
					createAndWaitForRunning()

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("PerconaXtraDB is not paused. Check for PerconaXtraDB")
					f.EventuallyPerconaXtraDB(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Check for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Update PerconaXtraDB to set spec.terminationPolicy = Pause")
					_, err := f.PatchPerconaXtraDB(perconaxtradb.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("with TerminationPolicyPause (default)", func() {

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should create DormantDatabase and resume from it", func() {
					// Run PerconaXtraDB and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Deleting PerconaXtraDB crd")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means PerconaXtraDB object is deleted
					By("Waiting for PerconaXtraDB to be paused")
					f.EventuallyDormantDatabaseStatus(perconaxtradb.ObjectMeta).Should(matcher.HavePaused())

					By("Checking PVC hasn't been deleted")
					f.EventuallyPVCCount(perconaxtradb.ObjectMeta).Should(Equal(1))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(perconaxtradb.ObjectMeta).Should(Equal(1))

					By("Checking snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipDataChecking {
						By("Check for snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}

					// Create PerconaXtraDB object again to resume it
					By("Create (resume) PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Wait for Running PerconaXtraDB")
					f.EventuallyPerconaXtraDBRunning(perconaxtradb.ObjectMeta).Should(BeTrue())

					By("Checking row count of table")
					f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
				})
			})

			Context("with TerminationPolicyDelete", func() {

				BeforeEach(func() {
					perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				AfterEach(func() {
					// delete snapshot and check for data wipeOut
					deleteSnapshot()

					By("Deleting secret: " + secret.Name)
					err := f.DeleteSecret(secret.ObjectMeta)
					if err != nil && !kerr.IsNotFound(err) {
						Expect(err).NotTo(HaveOccurred())
					}
				})

				It("should not create DormantDatabase and should not delete secret and snapshot", func() {
					// Run PerconaXtraDB and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until PerconaXtraDB is deleted")
					f.EventuallyPerconaXtraDB(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Checking PVC has been deleted")
					f.EventuallyPVCCount(perconaxtradb.ObjectMeta).Should(Equal(0))

					By("Checking Secret hasn't been deleted")
					f.EventuallyDBSecretCount(perconaxtradb.ObjectMeta).Should(Equal(1))

					By("Checking Snapshot hasn't been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeTrue())

					if !skipDataChecking {
						By("Check for intact snapshot data")
						f.EventuallySnapshotDataFound(snapshot).Should(BeTrue())
					}
				})
			})

			Context("with TerminationPolicyWipeOut", func() {

				BeforeEach(func() {
					perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				It("should not create DormantDatabase and should wipeOut all", func() {
					// Run PerconaXtraDB and take snapshot
					shouldInsertDataAndTakeSnapshot()

					By("Delete PerconaXtraDB")
					err = f.DeletePerconaXtraDB(perconaxtradb.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until PerconaXtraDB is deleted")
					f.EventuallyPerconaXtraDB(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(perconaxtradb.ObjectMeta).Should(BeFalse())

					By("Checking PVCs has been deleted")
					f.EventuallyPVCCount(perconaxtradb.ObjectMeta).Should(Equal(0))

					By("Checking Snapshots has been deleted")
					f.EventuallySnapshot(snapshot.ObjectMeta).Should(BeFalse())

					By("Checking Secrets has been deleted")
					f.EventuallyDBSecretCount(perconaxtradb.ObjectMeta).Should(Equal(0))
				})
			})
		})

		Context("EnvVars", func() {

			Context("Database Name as EnvVar", func() {

				It("should create DB with name provided in EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					perconaxtradb.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()
				})
			})

			Context("Root Password as EnvVar", func() {

				It("should reject to create PerconaXtraDB CRD", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					perconaxtradb.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_ROOT_PASSWORD,
							Value: "not@secret",
						},
					}
					By("Create PerconaXtraDB: " + perconaxtradb.Name)
					err = f.CreatePerconaXtraDB(perconaxtradb)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("Update EnvVar", func() {

				It("should not reject to update EvnVar", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					dbName = f.App()
					perconaxtradb.Spec.PodTemplate.Spec.Env = []core.EnvVar{
						{
							Name:  MYSQL_DATABASE,
							Value: dbName,
						},
					}
					//test general behaviour
					testGeneralBehaviour()

					By("Patching EnvVar")
					_, _, err = util.PatchPerconaXtraDB(f.ExtClient().KubedbV1alpha1(), perconaxtradb, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  MYSQL_DATABASE,
								Value: "patched-db",
							},
						}
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("Custom config", func() {

			customConfigs := []string{
				"max_connections=200",
				"read_buffer_size=1048576", // 1MB
			}

			Context("from configMap", func() {
				var userConfig *core.ConfigMap

				BeforeEach(func() {
					userConfig = f.GetCustomConfig(customConfigs)
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					err := f.DeleteConfigMap(userConfig.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

				})

				It("should set configuration provided in configMap", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					By("Creating configMap: " + userConfig.Name)
					err := f.CreateConfigMap(userConfig)
					Expect(err).NotTo(HaveOccurred())

					perconaxtradb.Spec.ConfigSource = &core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: userConfig.Name,
							},
						},
					}

					// Create PerconaXtraDB
					createAndWaitForRunning()

					By("Checking PerconaXtraDB configured from provided custom configuration")
					for _, cfg := range customConfigs {
						f.EventuallyPerconaXtraDBVariable(perconaxtradb.ObjectMeta, false, dbName, 0, cfg).Should(matcher.UseCustomConfig(cfg))
					}
				})
			})
		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create PerconaXtraDB
				createAndWaitForRunning()

				By("Creating Table")
				f.EventuallyCreateTable(perconaxtradb.ObjectMeta, false, dbName, 0).Should(BeTrue())

				By("Inserting Rows")
				f.EventuallyInsertRow(perconaxtradb.ObjectMeta, false, dbName, 0, 3).Should(BeTrue())

				By("Checking Row Count of Table")
				f.EventuallyCountRow(perconaxtradb.ObjectMeta, false, dbName, 0).Should(Equal(3))
			}

			Context("Ephemeral", func() {

				Context("General Behaviour", func() {

					BeforeEach(func() {
						perconaxtradb.Spec.StorageType = api.StorageTypeEphemeral
						perconaxtradb.Spec.Storage = nil
						perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						perconaxtradb.Spec.StorageType = api.StorageTypeEphemeral
						perconaxtradb.Spec.Storage = nil
						perconaxtradb.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create PerconaXtraDB object", func() {

						By("Creating PerconaXtraDB: " + perconaxtradb.Name)
						err := f.CreatePerconaXtraDB(perconaxtradb)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
