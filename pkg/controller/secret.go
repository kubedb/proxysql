package controller

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

const (
	proxysqlUser = "proxysql"
)

func (c *Controller) ensureProxySQLSecret(proxysql *api.ProxySQL) error {
	if proxysql.Spec.ProxySQLSecret == nil {
		secretVolumeSource, err := c.createProxySQLSecret(proxysql)
		if err != nil {
			return err
		}

		proxysqlPathced, _, err := util.PatchProxySQL(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQL) *api.ProxySQL {
			in.Spec.ProxySQLSecret = secretVolumeSource
			return in
		})
		if err != nil {
			return err
		}
		proxysql.Spec.ProxySQLSecret = proxysqlPathced.Spec.ProxySQLSecret
	}

	return nil
}

func (c *Controller) createProxySQLSecret(proxysql *api.ProxySQL) (*core.SecretVolumeSource, error) {
	authSecretName := proxysql.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, proxysql)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randProxysqlPassword := ""

		// if the password starts with "-", it will cause error in bash scripts (in proxysql-tools)
		for randProxysqlPassword = rand.GeneratePassword(); randProxysqlPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: proxysql.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				api.ProxySQLUserKey:     proxysqlUser,
				api.ProxySQLPasswordKey: randProxysqlPassword,
			},
		}

		if _, err := c.Client.CoreV1().Secrets(proxysql.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func (c *Controller) checkSecret(secretName string, proxysql *api.ProxySQL) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(proxysql.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		secret.Labels[api.LabelProxySQLName] != proxysql.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, proxysql.Namespace, secretName)
	}
	return secret, nil
}
