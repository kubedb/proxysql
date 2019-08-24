package framework

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kmodules.xyz/client-go/tools/portforward"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/proxysql/pkg/controller"
)

type KubedbTable struct {
	Id      int64
	PodName string
}

func proxysqlName(perconaName string) string {
	return perconaName + "-proxysql"
}

func (f *Framework) forwardPort(meta metav1.ObjectMeta, clientPodIndex, remotePort int) (*portforward.Tunnel, error) {
	clientPodName := fmt.Sprintf("%v-%d", meta.Name, clientPodIndex)
	tunnel := portforward.NewTunnel(
		f.kubeClient.CoreV1().RESTClient(),
		f.restConfig,
		meta.Namespace,
		clientPodName,
		remotePort,
	)

	if err := tunnel.ForwardPort(); err != nil {
		return nil, err
	}
	return tunnel, nil
}

func (f *Framework) getPerconaXtraDBClient(meta metav1.ObjectMeta, tunnel *portforward.Tunnel, dbName string, proxysql bool) (*xorm.Engine, error) {
	px, err := f.GetPerconaXtraDB(meta)
	if err != nil {
		return nil, err
	}

	var user, pass string
	var userErr, passErr error

	if !proxysql {
		user, userErr = f.GetMySQLCred(px, controller.KeyPerconaXtraDBUser)
		pass, passErr = f.GetMySQLCred(px, controller.KeyPerconaXtraDBPassword)
	} else {
		user, userErr = f.GetMySQLCred(px, api.ProxysqlUser)
		pass, passErr = f.GetMySQLCred(px, api.ProxysqlPassword)
	}
	if userErr != nil {
		return nil, userErr
	}
	if passErr != nil {
		return nil, passErr
	}

	cnnstr := fmt.Sprintf("%v:%v@tcp(127.0.0.1:%v)/%s", user, pass, tunnel.Local, dbName)
	return xorm.NewEngine("mysql", cnnstr)
}

func (f *Framework) EventuallyDatabaseReady(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return false
			}
			defer tunnel.Close()
			defer en.Close()

			return true
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) GetEngine(
	pxMeta metav1.ObjectMeta, proxysql bool,
	dbName string, forwardingPodIndex int) (*portforward.Tunnel, *xorm.Engine, error) {

	var (
		tunnel *portforward.Tunnel
		en     *xorm.Engine
		err    error
		sts    *appsv1.StatefulSet
		port   int
	)
	if proxysql {
		port = 6033
		sts, err = f.kubeClient.AppsV1().StatefulSets(pxMeta.Namespace).Get(proxysqlName(pxMeta.Name), metav1.GetOptions{})
	} else {
		port = 3306
		sts, err = f.kubeClient.AppsV1().StatefulSets(pxMeta.Namespace).Get(pxMeta.Name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, nil, err
	}

	tunnel, err = f.forwardPort(sts.ObjectMeta, forwardingPodIndex, port)
	if err != nil {
		return nil, nil, err
	}

	en, err = f.getPerconaXtraDBClient(pxMeta, tunnel, dbName, proxysql)
	if err != nil {
		return nil, nil, err
	}

	if err = en.Ping(); err != nil {
		return nil, nil, err
	}

	return tunnel, en, nil
}

func (f *Framework) EventuallyCreateDatabase(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return false
			}
			defer tunnel.Close()
			defer en.Close()

			_, err = en.Exec("CREATE DATABASE kubedb")
			if err != nil {
				return false
			}
			return true
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) EventuallyCreateTable(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return false
			}
			defer tunnel.Close()
			defer en.Close()

			err = en.Sync(new(KubedbTable))
			if err != nil {
				fmt.Println("creation error", err)
				return false
			}
			return true
		},
		time.Minute*10,
		time.Second*20,
	)
	return nil
}

func (f *Framework) EventuallyInsertRow(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex, rowsCntToInsert int) GomegaAsyncAssertion {
	count := 0
	return Eventually(
		func() bool {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return false
			}
			defer tunnel.Close()
			defer en.Close()

			for i := count; i < rowsCntToInsert; i++ {
				if _, err := en.Insert(&KubedbTable{
					//Id:      int64(i),
					PodName: fmt.Sprintf("%s-%v", pxMeta.Name, podIndex),
				}); err != nil {
					return false
				}
				count++
			}
			return true
		},
		time.Minute*10,
		time.Second*10,
	)
	return nil
}

func (f *Framework) EventuallyCountRow(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex int) GomegaAsyncAssertion {
	return Eventually(
		func() int {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return -1
			}
			defer tunnel.Close()
			defer en.Close()

			kubedb := new(KubedbTable)
			total, err := en.Count(kubedb)
			if err != nil {
				return -1
			}
			return int(total)
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) EventuallyPerconaXtraDBVariable(pxMeta metav1.ObjectMeta, proxysql bool, dbName string, podIndex int, config string) GomegaAsyncAssertion {
	configPair := strings.Split(config, "=")
	sql := fmt.Sprintf("SHOW VARIABLES LIKE '%s';", configPair[0])
	return Eventually(
		func() []map[string][]byte {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return nil
			}
			defer tunnel.Close()
			defer en.Close()

			results, err := en.Query(sql)
			if err != nil {
				return nil
			}
			return results
		},
		time.Minute*5,
		time.Second*5,
	)
}
