package e2e_test

import (
	"fmt"
	"strconv"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/proxysql/test/e2e/framework"
	"kubedb.dev/proxysql/test/e2e/matcher"

	"github.com/appscode/go/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var _ = Describe("PerconaXtraDB Cluster Tests", func() {
	var (
		err          error
		f            *framework.Invocation
		px           *api.PerconaXtraDB
		garbagePX    *api.PerconaXtraDBList
		dbName       string
		dbNameKubedb string

		wsClusterStats map[string]string
		proxysqlFlag   bool
		proxysql       *api.ProxySQL
	)

	var createAndWaitForRunningPerconaXtraDB = func() {
		By("Create PerconaXtraDB: " + px.Name)
		err = f.CreatePerconaXtraDB(px)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running PerconaXtraDB")
		f.EventuallyPerconaXtraDBRunning(px.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(px.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(px.ObjectMeta, api.ResourceKindPerconaXtraDB)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(px.ObjectMeta, proxysqlFlag, api.ResourceKindPerconaXtraDB, dbName, 0).Should(BeTrue())
	}

	var deletePerconaXtraDBResource = func() {
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

		By("Delete perconaxtradb")
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

			By("WipeOut perconaxtradb")
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
		f.EventuallyWipedOut(px.ObjectMeta, api.ResourceKindPerconaXtraDB).Should(Succeed())
	}

	var deleteProxySQLResource = func() {
		if proxysql == nil {
			log.Infoln("Skipping cleanup. Reason: ProxySQL object is nil")
			return
		}
		By("Check if ProxySQL " + proxysql.Name + " exists.")
		_, err = f.GetProxySQL(proxysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// ProxySQL was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}
		By("Delete ProxySQL")
		err = f.DeleteProxySQL(proxysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: ProxySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}
	}

	var deleteTestResource = func() {
		deletePerconaXtraDBResource()
		deleteProxySQLResource()
	}

	var deleteLeftOverStuffs = func() {
		// old PerconaXtraDBs are in garbagePX list. delete their resources.
		for _, p := range garbagePX.Items {
			*px = p
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers(map[string]string{
			api.LabelDatabaseKind: api.ResourceKindPerconaXtraDB,
		})
	}

	var createAndWaitForRunningProxySQL = func() {
		By("Create ProxySQL: " + proxysql.Name)
		err = f.CreateProxySQL(proxysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running ProxySQL")
		f.EventuallyProxySQLPhase(proxysql.ObjectMeta).Should(Equal(api.DatabasePhaseRunning))
	}

	var countRows = func(meta metav1.ObjectMeta, podIndex, expectedRowCnt int) {
		By(fmt.Sprintf("Read row from member '%s-%d'", meta.Name, podIndex))
		f.EventuallyCountRow(meta, proxysqlFlag, api.ResourceKindPerconaXtraDB, dbNameKubedb, podIndex).Should(Equal(expectedRowCnt))
	}

	var insertRows = func(meta metav1.ObjectMeta, podIndex, rowCntToInsert int) {
		By(fmt.Sprintf("Insert row on member '%s-%d'", meta.Name, podIndex))
		f.EventuallyInsertRow(meta, proxysqlFlag, dbNameKubedb, podIndex, rowCntToInsert).Should(BeTrue())
	}

	var create_Database_N_Table = func(meta metav1.ObjectMeta, podIndex int) {
		By("Create Database")
		f.EventuallyCreateDatabase(meta, proxysqlFlag, api.ResourceKindPerconaXtraDB, dbName, podIndex).Should(BeTrue())

		By("Create Table")
		f.EventuallyCreateTable(meta, proxysqlFlag, api.ResourceKindPerconaXtraDB, dbNameKubedb, podIndex).Should(BeTrue())
	}

	var readFromEachMember = func(meta metav1.ObjectMeta, clusterSize, rowCnt int) {
		for j := 0; j < clusterSize; j += 1 {
			countRows(meta, j, rowCnt)
		}
	}

	var writeTo_N_ReadFrom_EachMember = func(meta metav1.ObjectMeta, clusterSize, existingRowCnt int) {
		for i := 0; i < clusterSize; i += 1 {
			totalRowCnt := existingRowCnt + i + 1
			insertRows(meta, i, 1)
			readFromEachMember(meta, clusterSize, totalRowCnt)
		}
	}

	var replicationCheck = func(meta metav1.ObjectMeta, clusterSize int) {
		By("Checking replication")
		create_Database_N_Table(meta, 0)
		writeTo_N_ReadFrom_EachMember(meta, clusterSize, 0)
	}

	var storeWsClusterStats = func() {
		pods, err := f.KubeClient().CoreV1().Pods(px.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.Set(px.OffshootSelectors()).String(),
		})
		Expect(err).NotTo(HaveOccurred())

		wsClusterStats = map[string]string{
			"wsrep_local_state":         strconv.Itoa(4),
			"wsrep_local_state_comment": "Synced",
			"wsrep_evs_state":           "OPERATIONAL",
			"wsrep_cluster_size":        strconv.Itoa(len(pods.Items)),
			"wsrep_cluster_status":      "Primary",
			"wsrep_connected":           "ON",
			"wsrep_ready":               "ON",
		}
	}

	var CheckProxySQLVersion = func() {
		if framework.ProxySQLCatalogName != "2.0.4" {
			Skip("ProxySQL version must be '2.0.4'")
		}
	}

	var CheckPerconaXtraDBVersion = func() {
		if framework.PerconaXtraDBCatalogName != "5.7-cluster" {
			Skip("PerconaXtraDB version must be '5.7-cluster'")
		}
	}

	BeforeEach(func() {
		f = root.Invoke()
		px = f.PerconaXtraDB()
		garbagePX = new(api.PerconaXtraDBList)
		dbName = "mysql"
		dbNameKubedb = "kubedb"
		proxysqlFlag = false
	})

	Context("Proxysql", func() {
		BeforeEach(func() {
			if !framework.PerconaXtraDBTest {
				Skip("Value of '--percona-xtradb' flag must be 'true' to test ProxySQL for PerconaXtraDB")
			}

			CheckProxySQLVersion()
			CheckPerconaXtraDBVersion()

			createAndWaitForRunningPerconaXtraDB()
			storeWsClusterStats()

			proxysql = f.ProxySQL(api.ResourceKindPerconaXtraDB, px.Name)
			createAndWaitForRunningProxySQL()
		})

		AfterEach(func() {
			// delete resources for current PerconaXtraDB
			deleteTestResource()
			deleteLeftOverStuffs()
		})

		It("should configure poxysql for backend servers", func() {
			for i := 0; i < api.PerconaXtraDBDefaultClusterSize; i++ {
				By(fmt.Sprintf("Checking the cluster stats from Pod '%s-%d'", px.Name, i))
				f.EventuallyCheckCluster(px.ObjectMeta, proxysqlFlag, dbName, i, wsClusterStats).
					Should(Equal(true))
			}
			proxysqlFlag = true
			for i := 0; i < int(*proxysql.Spec.Replicas); i++ {
				By(fmt.Sprintf("Checking the cluster stats from Proxysql Pod '%s-%d'", proxysql.Name, i))
				f.EventuallyCheckCluster(proxysql.ObjectMeta, proxysqlFlag, dbName, i, wsClusterStats).
					Should(Equal(true))
			}
			replicationCheck(proxysql.ObjectMeta, int(*proxysql.Spec.Replicas))
		})
	})
})
