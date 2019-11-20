package e2e_test

import (
	"fmt"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/proxysql/test/e2e/framework"
	"kubedb.dev/proxysql/test/e2e/matcher"

	"github.com/appscode/go/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MySQL Group Replication Tests", func() {
	var (
		err          error
		f            *framework.Invocation
		mysql        *api.MySQL
		garbageMySQL *api.MySQLList
		dbName       string
		dbNameKubedb string

		proxysqlFlag bool
		proxysql     *api.ProxySQL
	)

	var createAndWaitForRunningMySQL = func() {
		By("Create MySQL: " + mysql.Name)
		err = f.CreateMySQL(mysql)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running mysql")
		f.EventuallyMySQLRunning(mysql.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(mysql.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(mysql.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for database to be ready")
		f.EventuallyDatabaseReady(mysql.ObjectMeta, proxysqlFlag, dbName, 0).Should(BeTrue())
	}

	var deleteMySQLResource = func() {
		if mysql == nil {
			log.Infoln("Skipping cleanup. Reason: mysql is nil")
			return
		}

		By("Check if mysql " + mysql.Name + " exists.")
		my, err := f.GetMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// MySQL was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete mysql")
		err = f.DeleteMySQL(mysql.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				log.Infoln("Skipping rest of the cleanup. Reason: MySQL does not exist.")
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if my.Spec.TerminationPolicy == api.TerminationPolicyPause {
			By("Wait for mysql to be paused")
			f.EventuallyDormantDatabaseStatus(mysql.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut mysql")
			_, err := f.PatchDormantDatabase(mysql.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(mysql.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for mysql resources to be wipedOut")
		f.EventuallyWipedOutMySQL(mysql.ObjectMeta).Should(Succeed())
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
		deleteMySQLResource()
		deleteProxySQLResource()
	}

	var deleteLeftOverStuffs = func() {
		// old MySQL are in garbageMySQL list. delete their resources.
		for _, my := range garbageMySQL.Items {
			*mysql = my
			deleteTestResource()
		}

		By("Delete left over workloads if exists any")
		f.CleanWorkloadLeftOvers(map[string]string{
			api.LabelDatabaseKind: api.ResourceKindMySQL,
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
		f.EventuallyCountRow(meta, proxysqlFlag, dbNameKubedb, podIndex).Should(Equal(expectedRowCnt))
	}

	var insertRows = func(meta metav1.ObjectMeta, podIndex, rowCntToInsert int, expected bool) {
		By(fmt.Sprintf("Insert row on member '%s-%d' should be %v", meta.Name, podIndex, expected))
		if expected {
			f.EventuallyInsertRow(meta, proxysqlFlag, dbNameKubedb, podIndex, rowCntToInsert).Should(BeTrue())
		} else {
			f.EventuallyInsertRow(meta, proxysqlFlag, dbNameKubedb, podIndex, rowCntToInsert).Should(BeFalse())
		}
	}

	var create_Database_N_Table = func(meta metav1.ObjectMeta, podIndex int) {
		By("Create Database")
		f.EventuallyCreateDatabase(meta, proxysqlFlag, dbName, podIndex).Should(BeTrue())

		By("Create Table")
		f.EventuallyCreateTable(meta, proxysqlFlag, dbNameKubedb, podIndex).Should(BeTrue())
	}

	var writeToPrimary = func(meta metav1.ObjectMeta, podIndex int) {
		By(fmt.Sprintf("Write on '%s-%d'", meta.Name, podIndex))
		insertRows(meta, podIndex, 1, true)
	}

	var readFromEachMember = func(meta metav1.ObjectMeta, clusterSize, rowCnt int) {
		for j := 0; j < clusterSize; j += 1 {
			countRows(meta, j, rowCnt)
		}
	}

	var writeTo_Primary_N_ReadFrom_EachMember = func(meta metav1.ObjectMeta, primaryPodIndex, clusterSize int) {
		writeToPrimary(meta, primaryPodIndex)
		readFromEachMember(meta, clusterSize, 1)
	}

	var replicationCheck = func(meta metav1.ObjectMeta, primaryPodIndex, clusterSize int) {
		By("Checking replication")
		create_Database_N_Table(meta, primaryPodIndex)
		writeTo_Primary_N_ReadFrom_EachMember(meta, primaryPodIndex, clusterSize)
	}

	var CheckProxySQLVersionForXtraDBCluster = func() {
		if framework.ProxySQLCatalogName != "2.0.4" {
			Skip("For XtraDB Cluster, currently supported ProxySQL version is '2.0.4'")
		}
	}

	BeforeEach(func() {
		f = root.Invoke()
		mysql = f.MySQLGroup()
		garbageMySQL = new(api.MySQLList)
		dbName = "mysql"
		dbNameKubedb = "kubedb"
		proxysqlFlag = false
	})

	Context("Proxysql", func() {
		BeforeEach(func() {
			if !framework.ProxySQLTest {
				Skip("For ProxySQL test, the value of '--proxysql' flag must be 'true' while running e2e-tests command")
			}

			CheckProxySQLVersionForXtraDBCluster()

			createAndWaitForRunningPerconaXtraDB()
			storeWsClusterStats()

			psql = f.ProxySQL(px.Name)
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
				f.EventuallyCheckCluster(px.ObjectMeta, proxysql, dbName, i, wsClusterStats).
					Should(Equal(true))
			}
			proxysql = true
			for i := 0; i < int(*psql.Spec.Replicas); i++ {
				By(fmt.Sprintf("Checking the cluster stats from Proxysql Pod '%s-%d'", psql.Name, i))
				f.EventuallyCheckCluster(psql.ObjectMeta, proxysql, dbName, i, wsClusterStats).
					Should(Equal(true))
			}
			replicationCheck(psql.ObjectMeta, int(*psql.Spec.Replicas))
			proxysql = false
			readFromEachPrimary(px.ObjectMeta, api.PerconaXtraDBDefaultClusterSize, int(*psql.Spec.Replicas))
		})
	})
})
