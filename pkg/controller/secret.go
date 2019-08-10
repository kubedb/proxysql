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

	KeyPerconaXtraDBUser     = "username"
	KeyPerconaXtraDBPassword = "password"
)

func (c *Controller) ensureDatabaseSecret(px *api.PerconaXtraDB) error {
	if px.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(px)
		if err != nil {
			return err
		}

		per, _, err := util.PatchPerconaXtraDB(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			return err
		}
		px.Spec.DatabaseSecret = per.Spec.DatabaseSecret
		return nil
	}
	return c.upgradeDatabaseSecret(px)
}

func (c *Controller) createDatabaseSecret(px *api.PerconaXtraDB) (*core.SecretVolumeSource, error) {
	authSecretName := px.Name + "-auth"

	sc, err := c.checkSecret(authSecretName, px)
	if err != nil {
		return nil, err
	}
	if sc == nil {
		randPassword := ""

		// if the password starts with "-", it will cause error in bash scripts (in percona-xtradb-tools)
		for randPassword = rand.GeneratePassword(); randPassword[0] == '-'; {
		}

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:   authSecretName,
				Labels: px.OffshootSelectors(),
			},
			Type: core.SecretTypeOpaque,
			StringData: map[string]string{
				KeyPerconaXtraDBUser:     mysqlUser,
				KeyPerconaXtraDBPassword: randPassword,
			},
		}

		if px.Spec.PXC != nil {
			randProxysqlPassword := ""

			// if the password starts with "-", it will cause error in bash scripts (in percona-xtradb-tools)
			for randProxysqlPassword = rand.GeneratePassword(); randProxysqlPassword[0] == '-'; {
			}

			secret.StringData[api.ProxysqlUser] = "proxysql"
			secret.StringData[api.ProxysqlPassword] = randProxysqlPassword
		}

		if _, err := c.Client.CoreV1().Secrets(px.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}
	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

// This is done to fix 0.8.0 -> 0.9.0 upgrade due to
// https://github.com/kubedb/percona-xtradb/pull/115/files#diff-10ddaf307bbebafda149db10a28b9c24R17 commit
func (c *Controller) upgradeDatabaseSecret(px *api.PerconaXtraDB) error {
	meta := metav1.ObjectMeta{
		Name:      px.Spec.DatabaseSecret.SecretName,
		Namespace: px.Namespace,
	}

	_, _, err := core_util.CreateOrPatchSecret(c.Client, meta, func(in *core.Secret) *core.Secret {
		if _, ok := in.Data[KeyPerconaXtraDBUser]; !ok {
			if val, ok2 := in.Data["user"]; ok2 {
				in.StringData = map[string]string{KeyPerconaXtraDBUser: string(val)}
			}
		}
		return in
	})
	return err
}

func (c *Controller) checkSecret(secretName string, px *api.PerconaXtraDB) (*core.Secret, error) {
	secret, err := c.Client.CoreV1().Secrets(px.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if secret.Labels[api.LabelDatabaseKind] != api.ResourceKindPerconaXtraDB ||
		secret.Labels[api.LabelDatabaseName] != px.Name {
		return nil, fmt.Errorf(`intended secret "%v/%v" already exists`, px.Namespace, secretName)
	}
	return secret, nil
}
