/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"context"
	"fmt"
	"strconv"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/proxysql/test/e2e/framework"

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
			log.Infoln("Skipping cleanup. Reason: PerconaXtraDB object is nil")
			return
		}

		By("Check if perconaxtradb " + px.Name + " exists.")
		px, err := f.GetPerconaXtraDB(px.ObjectMeta)
		if err != nil && kerr.IsNotFound(err) {
			// PerconaXtraDB was not created. Hence, rest of cleanup is not necessary.
			return
		}
		Expect(err).NotTo(HaveOccurred())

		By("Update perconaxtradb to set spec.terminationPolicy = WipeOut")
		_, err = f.PatchPerconaXtraDB(px.ObjectMeta, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		})
		Expect(err).NotTo(HaveOccurred())

		By("Delete perconaxtradb")
		err = f.DeletePerconaXtraDB(px.ObjectMeta)
		if err != nil && kerr.IsNotFound(err) {
			// PerconaXtraDB was not created. Hence, rest of cleanup is not necessary.
			return
		}
		Expect(err).NotTo(HaveOccurred())

		By("Wait for perconaxtradb to be deleted")
		f.EventuallyPerconaXtraDB(px.ObjectMeta).Should(BeFalse())

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
		pods, err := f.KubeClient().CoreV1().Pods(px.Namespace).List(context.TODO(), metav1.ListOptions{
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

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			f.PrintDebugHelpers("", px.Name, 0, int(*px.Spec.Replicas))
		}
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
