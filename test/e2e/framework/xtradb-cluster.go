package framework

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) EventuallyCheckCluster(
	pxMeta metav1.ObjectMeta, proxysql bool,
	dbName string, podIndex int,
	clusterStats map[string]string) GomegaAsyncAssertion {

	return Eventually(
		func() bool {
			tunnel, en, err := f.GetEngine(pxMeta, proxysql, dbName, podIndex)
			if err != nil {
				return false
			}
			defer tunnel.Close()
			defer en.Close()

			r := make([]map[string]string, 0)
			r, err = en.QueryString("show status like \"wsrep%%\"")
			if err != nil {
				return false
			}

			ch := true
			for _, m := range r {
				if m["Variable_name"] == "wsrep_local_state" {
					ch = ch && m["Value"] == clusterStats["wsrep_local_state"]
				}
				if m["Variable_name"] == "wsrep_local_state_comment" {
					ch = ch && m["Value"] == clusterStats["wsrep_local_state_comment"]
				}
				if m["Variable_name"] == "wsrep_incoming_addresses" {
					addrsExpected := strings.Split(clusterStats["wsrep_incoming_addresses"], ",")
					addrsGot := strings.Split(m["Value"], ",")
					var flag bool
					for _, addrG := range addrsGot {
						flag = false
						for _, addrE := range addrsExpected {
							if addrG == addrE {
								flag = true
							}
						}
						ch = ch && flag
					}
				}
				if m["Variable_name"] == "wsrep_evs_state" {
					ch = ch && m["Value"] == clusterStats["wsrep_evs_state"]
				}
				if m["Variable_name"] == "wsrep_cluster_size" {
					ch = ch && m["Value"] == clusterStats["wsrep_cluster_size"]
				}
				if m["Variable_name"] == "wsrep_cluster_status" {
					ch = ch && m["Value"] == clusterStats["wsrep_cluster_status"]
				}
				if m["Variable_name"] == "wsrep_connected" {
					ch = ch && m["Value"] == clusterStats["wsrep_connected"]
				}
				if m["Variable_name"] == "wsrep_ready" {
					ch = ch && m["Value"] == clusterStats["wsrep_ready"]
				}
			}

			return ch
		},
		time.Minute*10,
		time.Second*20,
	)
}

func (f *Framework) RemoverPrimary(meta metav1.ObjectMeta, primaryPodIndex int) error {
	if _, err := f.kubeClient.CoreV1().Pods(meta.Namespace).Get(
		fmt.Sprintf("%s-%d", meta.Name, primaryPodIndex),
		metav1.GetOptions{},
	); err != nil {
		return err
	}

	return f.kubeClient.CoreV1().Pods(meta.Namespace).Delete(
		fmt.Sprintf("%s-%d", meta.Name, primaryPodIndex),
		&metav1.DeleteOptions{},
	)
}
