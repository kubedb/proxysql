package framework

import (
	"fmt"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func (f *Framework) CleanWorkloadLeftOvers(labelselector map[string]string) {
	// delete statefulset
	if err := f.kubeClient.AppsV1().StatefulSets(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Statefulset. Error: %v", err)
	}

	// delete pvc
	if err := f.kubeClient.CoreV1().PersistentVolumeClaims(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of PVC. Error: %v", err)
	}

	// delete secret
	if err := f.kubeClient.CoreV1().Secrets(f.namespace).DeleteCollection(deleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Secret. Error: %v", err)
	}
}
