package e2e_test

import (
	"fmt"
	"os"
	"strconv"

	"github.com/appscode/go/log"
	"github.com/appscode/go/strings"
	"github.com/appscode/go/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	store "kmodules.xyz/objectstore-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/percona-xtradb/test/e2e/framework"
	"kubedb.dev/percona-xtradb/test/e2e/matcher"
	stashV1alpha1 "stash.appscode.dev/stash/apis/stash/v1alpha1"
	stashV1beta1 "stash.appscode.dev/stash/apis/stash/v1beta1"
)

var _ = Describe("PerconaXtraDB cluster Tests", func() {
	const (
		googleProjectIDKey          = "GOOGLE_PROJECT_ID"
		googleServiceAccountJsonKey = "GOOGLE_SERVICE_ACCOUNT_JSON_KEY"
		googleBucketNameKey         = "GCS_BUCKET_NAME"
	)

	var (
		err                  error
		f                    *framework.Invocation
		px                   *api.PerconaXtraDB
		garbagePerconaXtraDB *api.PerconaXtraDBList
		dbName               string
		dbNameKubedb         string
		wsClusterStats       map[string]string
		secret               *corev1.Secret
	)

	var isSetEnv = func(key string) bool {
		_, set := os.LookupEnv(key)

		return set
	}

	var createAndWaitForRunning = func() {
		By("Create PerconaXtraDB: " + px.Name)
		err = f.CreatePerconaXtraDB(px)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running PerconaXtraDB")
		f.EventuallyPerconaXtraDBRunning(px.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(px.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(px.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(px.ObjectMeta, false, dbName, 0).Should(BeTrue())
	}

	var deleteTestResource = func() {
		if px == nil {
			log.Infoln("Skipping cleanup. Reason: perconaxtradb is nil")
			return
		}

		By("Check if perconaxtradb " + px.Name + " exists.")
		my, err := f.GetPerconaXtraDB(px.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// PerconaXtraDB was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete px")
		err = f.DeletePerconaXtraDB(px.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: PerconaXtraDB does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if my.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for perconaxtradb to be paused")
			f.EventuallyDormantDatabaseStatus(px.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut PerconaXtraDB")
			_, err := f.PatchDormantDatabase(px.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(px.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for perconaxtradb resources to be wipedOut")
		f.EventuallyWipedOut(px.ObjectMeta).Should(Succeed())
	}

	var deleteLeftOverStuffs = func() {
		// old PerconaXtraDB are in garbagePerconaXtraDB list. delete their resources.
		for _, p := range garbagePerconaXtraDB.Items {
			*px = p
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers()
	}

	var baseName = func(proxysql bool) string {
		if proxysql {
			return px.ProxysqlName()
		}

		return px.Name
	}

	var countRows = func(proxysql bool, podIndex, expectedRowCnt int) {
		By(fmt.Sprintf("Read row from member '%s-%d'", baseName(proxysql), podIndex))
		f.EventuallyCountRow(px.ObjectMeta, proxysql, dbNameKubedb, podIndex).Should(Equal(expectedRowCnt))
	}

	var insertRows = func(proxysql bool, podIndex, rowCntToInsert int) {
		By(fmt.Sprintf("Insert row on member '%s-%d'", baseName(proxysql), podIndex))
		f.EventuallyInsertRow(px.ObjectMeta, proxysql, dbNameKubedb, podIndex, rowCntToInsert).Should(BeTrue())
	}

	var create_Database_N_Table = func(proxysql bool) {
		By("Create Database")
		f.EventuallyCreateDatabase(px.ObjectMeta, proxysql, dbName, 0).Should(BeTrue())

		By("Create Table")
		f.EventuallyCreateTable(px.ObjectMeta, proxysql, dbNameKubedb, 0).Should(BeTrue())
	}

	var readFromEachPrimary = func(clusterSize, rowCnt int, proxysql bool) {
		for j := 0; j < clusterSize; j += 1 {
			countRows(proxysql, j, rowCnt)
		}
	}

	var writeTo_N_ReadFrom_EachPrimary = func(clusterSize, existingRowCnt int, proxysql bool) {
		for i := 0; i < clusterSize; i += 1 {
			rowCnt := existingRowCnt + i + 1
			insertRows(proxysql, i, 1)
			readFromEachPrimary(clusterSize, rowCnt, proxysql)
		}
	}

	var replicationCheck = func(clusterSize int, proxysql bool) {
		By("Checking replication")
		create_Database_N_Table(proxysql)

		writeTo_N_ReadFrom_EachPrimary(clusterSize, 0, proxysql)
	}

	var storeWsClusterStats = func() {
		pods, err := f.KubeClient().CoreV1().Pods(px.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set(px.ClusterSelectors()).String(),
		})
		Expect(err).NotTo(HaveOccurred())
		clusterMembersAddr := make([]*string, 0)
		for _, pod := range pods.Items {
			addr := fmt.Sprintf("%s:%d", pod.Status.PodIP, api.MySQLNodePort)
			clusterMembersAddr = append(clusterMembersAddr, &addr)
		}

		wsClusterStats = map[string]string{
			"wsrep_local_state":         strconv.Itoa(4),
			"wsrep_local_state_comment": "Synced",
			"wsrep_incoming_addresses":  strings.Join(clusterMembersAddr, ","),
			"wsrep_evs_state":           "OPERATIONAL",
			"wsrep_cluster_size":        strconv.Itoa(len(pods.Items)),
			"wsrep_cluster_status":      "Primary",
			"wsrep_connected":           "ON",
			"wsrep_ready":               "ON",
		}
	}

	var CheckDBVersionForXtraDBCluster = func() {
		if framework.DBCatalogName != "5.7" {
			Skip("For XtraDB Cluster, currently supported DB version is '5.7'")
		}
	}

	BeforeEach(func() {
		f = root.Invoke()
		px = f.PerconaXtraDBCluster()
		garbagePerconaXtraDB = new(api.PerconaXtraDBList)
		dbName = "mysql"
		dbNameKubedb = "kubedb"

		CheckDBVersionForXtraDBCluster()
	})

	Context("Behaviour tests", func() {

		AfterEach(func() {
			// delete resources for current PerconaXtraDB
			deleteTestResource()

			deleteLeftOverStuffs()
		})

		Context("Basic Cluster with 3 member", func() {
			BeforeEach(func() {
				createAndWaitForRunning()
				storeWsClusterStats()
			})

			It("should be possible to create a basic 3 member cluster", func() {
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}

				replicationCheck(api.PerconaXtraDBDefaultClusterSize, false)
			})
		})

		Context("Failover", func() {
			BeforeEach(func() {
				createAndWaitForRunning()
				storeWsClusterStats()
			})

			It("should failover successfully", func() {
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}
				replicationCheck(api.PerconaXtraDBDefaultClusterSize, false)

				By(fmt.Sprintf("Taking down the primary '%s-%d'", px.Name, 0))
				err = f.RemoverPrimary(px.ObjectMeta, 0)
				Expect(err).NotTo(HaveOccurred())

				By(fmt.Sprintf("Checking status after failing primary '%s-%d'", px.Name, 0))
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats member count from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}

				By("Checking for data after failover")
				readFromEachPrimary(api.PerconaXtraDBDefaultClusterSize, 3, false)
			})
		})

		Context("Scale up", func() {
			BeforeEach(func() {
				createAndWaitForRunning()
				storeWsClusterStats()
			})

			It("should be possible to scale up", func() {
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}
				replicationCheck(api.PerconaXtraDBDefaultClusterSize, false)

				By("Scaling up")
				px, err = f.PatchPerconaXtraDB(px.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
					in.Spec.Replicas = types.Int32P(api.PerconaXtraDBDefaultClusterSize + 1)

					return in
				})
				Expect(err).NotTo(HaveOccurred())
				By("Wait for perconaxtradb be patched")
				Expect(f.WaitUntilPerconaXtraDBReplicasBePatched(px.ObjectMeta, api.PerconaXtraDBDefaultClusterSize+1)).
					NotTo(HaveOccurred())

				By("Wait for new member to be ready")
				Expect(f.WaitUntilPodRunningBySelector(px, false)).NotTo(HaveOccurred())
				By("Wait for proxysql to be ready")
				Expect(f.WaitUntilPodRunningBySelector(px, true)).NotTo(HaveOccurred())

				By("Checking status after scaling up")
				storeWsClusterStats()
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize+1; i++ {
					By(fmt.Sprintf("Checking the cluster stats member count from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}

				By("Checking for data after scaling up")
				readFromEachPrimary(api.PerconaXtraDBDefaultClusterSize+1, 3, false)
				writeTo_N_ReadFrom_EachPrimary(api.PerconaXtraDBDefaultClusterSize+1, 3, false)
			})
		})

		Context("Scale down", func() {
			BeforeEach(func() {
				px.Spec.Replicas = types.Int32P(4)

				createAndWaitForRunning()
				storeWsClusterStats()
			})

			It("Should be possible to scale down", func() {
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize+1; i++ {
					By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}
				replicationCheck(api.PerconaXtraDBDefaultClusterSize+1, false)

				By("Scaling down")
				px, err = f.PatchPerconaXtraDB(px.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
					in.Spec.Replicas = types.Int32P(api.PerconaXtraDBDefaultClusterSize)

					return in
				})
				Expect(err).NotTo(HaveOccurred())
				By("Wait for perconaxtradb be patched")
				Expect(f.WaitUntilPerconaXtraDBReplicasBePatched(px.ObjectMeta, api.PerconaXtraDBDefaultClusterSize)).
					NotTo(HaveOccurred())

				By("Wait for new member to be ready")
				Expect(f.WaitUntilPodRunningBySelector(px, false)).NotTo(HaveOccurred())
				By("Wait for proxysql to be ready")
				Expect(f.WaitUntilPodRunningBySelector(px, true)).NotTo(HaveOccurred())

				By("Checking status after scaling down")
				storeWsClusterStats()
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats member count from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}

				By("Checking for data after scaling down")
				readFromEachPrimary(api.PerconaXtraDBDefaultClusterSize, 4, false)
				writeTo_N_ReadFrom_EachPrimary(api.PerconaXtraDBDefaultClusterSize, 4, false)
			})
		})

		Context("Proxysql", func() {
			BeforeEach(func() {
				createAndWaitForRunning()
				storeWsClusterStats()

			})

			It("should configure poxysql for backend server", func() {
				for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
					By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
					f.EventuallyCheckCluster(px.ObjectMeta, false, dbName, i, wsClusterStats).
						Should(Equal(true))
				}
				for i := 0; i < int(*px.Spec.PXC.Proxysql.Replicas); i++ {
					By(fmt.Sprintf("Checking the cluster stats from Proxysql Pod '%s-%d'", px.ProxysqlName(), i))
					f.EventuallyCheckCluster(px.ObjectMeta, true, dbName, i, wsClusterStats).
						Should(Equal(true))
				}
				replicationCheck(int(*px.Spec.PXC.Proxysql.Replicas), true)
			})
		})
	})

	Context("Initialize", func() {
		// To run this test,
		// 1st: Deploy stash latest operator
		// 2nd: create mysql related tasks and functions either
		//	 or	from helm chart in `stash.appscode.dev/percona-xtradb/chart/stash-percona-xtradb`
		Context("With Stash/Restic", func() {
			var bc *stashV1beta1.BackupConfiguration
			var bs *stashV1beta1.BackupSession
			var rs *stashV1beta1.RestoreSession
			var repo *stashV1alpha1.Repository

			BeforeEach(func() {
				if !f.FoundStashCRDs() {
					Skip("Skipping tests for stash integration. reason: stash operator is not running.")
				}

				if !isSetEnv(googleProjectIDKey) ||
					!isSetEnv(googleServiceAccountJsonKey) ||
					!isSetEnv(googleBucketNameKey) {

					Skip("Skipping tests for stash integration. reason: " +
						fmt.Sprintf("env vars %q, %q and %q are required",
							googleProjectIDKey, googleServiceAccountJsonKey, googleBucketNameKey))
				}
			})

			AfterEach(func() {
				By("Deleting BackupConfiguration")
				err := f.DeleteBackupConfiguration(bc.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Deleting RestoreSession")
				err = f.DeleteRestoreSession(rs.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Deleting Repository")
				err = f.DeleteRepository(repo.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				deleteTestResource()
				deleteLeftOverStuffs()
			})

			var createAndWaitForInitializing = func() {
				By("Creating MySQL: " + px.Name)
				err = f.CreatePerconaXtraDB(px)
				Expect(err).NotTo(HaveOccurred())

				By("Wait for Initializing mysql")
				f.EventuallyPerconaXtraDBPhase(px.ObjectMeta).Should(Equal(api.DatabasePhaseInitializing))
			}

			var shouldInitializeFromStash = func() {
				// Create and wait for running MySQL
				createAndWaitForRunning()

				create_Database_N_Table(false)
				insertRows(false, 0, 3)
				countRows(false, 0, 3)

				By("Create Secret")
				err = f.CreateSecret(secret)
				Expect(err).NotTo(HaveOccurred())

				By("Create Repositories")
				err = f.CreateRepository(repo)
				Expect(err).NotTo(HaveOccurred())

				By("Create BackupConfiguration")
				err = f.CreateBackupConfiguration(bc)
				Expect(err).NotTo(HaveOccurred())

				By("Wait until BackupSession be created")
				bs, err = f.WaitUntilBackkupSessionBeCreated(bc.ObjectMeta)

				// eventually backupsession succeeded
				By("Check for Succeeded backupsession")
				f.EventuallyBackupSessionPhase(bs.ObjectMeta).Should(Equal(stashV1beta1.BackupSessionSucceeded))

				oldPerconaXtraDB, err := f.GetPerconaXtraDB(px.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				garbagePerconaXtraDB.Items = append(garbagePerconaXtraDB.Items, *oldPerconaXtraDB)

				By("Create perconaxtradb for initializing from stash")
				*px = *f.PerconaXtraDBCluster()
				rs = f.RestoreSession(px.ObjectMeta, oldPerconaXtraDB.ObjectMeta, oldPerconaXtraDB.Spec.Replicas)
				px.Spec.DatabaseSecret = oldPerconaXtraDB.Spec.DatabaseSecret
				px.Spec.Init = &api.InitSpec{
					StashRestoreSession: &corev1.LocalObjectReference{
						Name: rs.Name,
					},
				}

				// Create and wait for running MySQL
				createAndWaitForInitializing()

				By("Create RestoreSession")
				err = f.CreateRestoreSession(rs)
				Expect(err).NotTo(HaveOccurred())

				// eventually restoresession succeeded
				By("Check for Succeeded restoreSession")
				f.EventuallyRestoreSessionPhase(rs.ObjectMeta).Should(Equal(stashV1beta1.RestoreSessionSucceeded))

				By("Wait for Running mysql")
				f.EventuallyPerconaXtraDBRunning(px.ObjectMeta).Should(BeTrue())

				By("Wait for AppBinding to create")
				f.EventuallyAppBinding(px.ObjectMeta).Should(BeTrue())

				By("Check valid AppBinding Specs")
				err = f.CheckAppBindingSpec(px.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				By("Waiting for database to be ready")
				f.EventuallyDatabaseReady(px.ObjectMeta, false, dbName, 0).Should(BeTrue())

				countRows(false, 0, 3)
			}

			Context("From GCS backend", func() {

				BeforeEach(func() {
					secret = f.SecretForGCSBackend()
					secret = f.PatchSecretForRestic(secret)
					bc = f.BackupConfiguration(px.ObjectMeta)
					repo = f.Repository(px.ObjectMeta, secret.Name)

					repo.Spec.Backend = store.Backend{
						GCS: &store.GCSSpec{
							Bucket: os.Getenv(googleBucketNameKey),
							Prefix: fmt.Sprintf("stash/%v/%v", px.Namespace, px.Name),
						},
						StorageSecretName: secret.Name,
					}
				})

				It("should run successfully", shouldInitializeFromStash)
			})

		})

	})
})
