package controller

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

const (
	mysqlUser = "root"

	KeyProxySQLUser     = "username"
	KeyProxySQLPassword = "password"
)

func (c *Controller) ensureDatabaseSecret(proxysql *api.ProxySQL) error {
	if proxysql.Spec.ProxySQLSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(proxysql)
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
		return nil
	}
	return c.upgradeDatabaseSecret(proxysql)
}

func (c *Controller) createDatabaseSecret(proxysql *api.ProxySQL) (*core.SecretVolumeSource, error) {
	authSecretName := proxysql.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, proxysql)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-", it will cause error in bash scripts (in proxysql-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: proxysql.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyProxySQLUser:     mysqlUser,
				KeyProxySQLPassword: randPassword,
			},
		}

		if proxysql.Spec.PXC != nil {
			randProxysqlPassword := ""

			// if the password starts with "-", it will cause error in bash scripts (in proxysql-tools)
			for randProxysqlPassword = rand.GeneratePassword(); randProxysqlPassword[0] == '-'; {
			}

			secret.StringData[api.ProxysqlUser] = "proxysql"
			secret.StringData[api.ProxysqlPassword] = randProxysqlPassword
		}

		if _, err := c.Client.CoreV1().Secrets(proxysql.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/proxysql/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Controller) upgradeDatabaseSecret(proxysql *api.ProxySQL) error {
	meta := metav1.ObjectMeta{
		Name:      proxysql.Spec.ProxySQLSecret.SecretName,
		Namespace: proxysql.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(c.Client, meta, func(in *core.Secret) *core.Secret {
		if _, ok := in.Data[KeyProxySQLUser]; !ok {
			if val, ok2 := in.Data["user"]; ok2 {
				in.StringData = map[string]string{KeyProxySQLUser: string(val)}
			}
		}
		return in
	})
	return err
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
		secret.Labels[api.LabelDatabaseName] != proxysql.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, proxysql.Namespace, secretName)
	}
	return secret, nil
}
