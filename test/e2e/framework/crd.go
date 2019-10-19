package framework

import (
	"errors"
	"time"

	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) EventuallyCRD() GomegaAsyncAssertion {
	return Eventually(
		func() error {
			// Check ProxySQL
			if _, err := f.dbClient.KubedbV1alpha1().ProxySQLs(core.NamespaceAll).List(metav1.ListOptions{}); err != nil {
				return errors.New("CRD ProxySQL is not ready")
			}

			// Check MySQL TPR
			if MySQLTest {
				if _, err := f.dbClient.KubedbV1alpha1().MySQLs(core.NamespaceAll).List(metav1.ListOptions{}); err != nil {
					return errors.New("CRD MySQL is not ready")
				}
			}

			// Check DormantDatabases TPR
			if _, err := f.dbClient.KubedbV1alpha1().DormantDatabases(core.NamespaceAll).List(metav1.ListOptions{}); err != nil {
				return errors.New("CRD DormantDatabase is not ready")
			}

			return nil
		},
		time.Minute*2,
		time.Second*10,
	)
}
