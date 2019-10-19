package framework

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetSecretCred(secretMeta metav1.ObjectMeta, key string) (string, error) {
	secret, err := f.kubeClient.CoreV1().Secrets(secretMeta.Namespace).Get(secretMeta.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	data := string(secret.Data[key])
	return data, nil
}
