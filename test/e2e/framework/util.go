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

package framework

import (
	"context"
	"fmt"

	shell "github.com/codeskyblue/go-sh"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	meta_util "kmodules.xyz/client-go/meta"
)

func (f *Framework) CleanWorkloadLeftOvers(labelselector map[string]string) {
	// delete statefulset
	if err := f.kubeClient.AppsV1().StatefulSets(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Statefulset. Error: %v", err)
	}

	// delete pvc
	if err := f.kubeClient.CoreV1().PersistentVolumeClaims(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of PVC. Error: %v", err)
	}

	// delete secret
	if err := f.kubeClient.CoreV1().Secrets(f.namespace).DeleteCollection(context.TODO(), meta_util.DeleteInForeground(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelselector).String(),
	}); err != nil && !kerr.IsNotFound(err) {
		fmt.Printf("error in deletion of Secret. Error: %v", err)
	}
}

func (f *Framework) PrintDebugHelpers(myName, pxName string, myReplicas, pxReplicas int) {
	sh := shell.NewSession()

	fmt.Println("\n======================================[ Apiservices ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "get", "apiservice", "v1alpha1.mutators.kubedb.com", "-o=jsonpath=\"{.status}\"").Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println()
	if err := sh.Command("/usr/bin/kubectl", "get", "apiservice", "v1alpha1.validators.kubedb.com", "-o=jsonpath=\"{.status}\"").Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ Describe Job ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "job", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ Describe Pod ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "po", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ Describe ProxySQL ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "proxysql", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ Describe MySQL ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "mysql", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ MySQL Server Log ]===================================================")
	for i := 0; i < myReplicas; i++ {
		if err := sh.Command("/usr/bin/kubectl", "logs", fmt.Sprintf("%s-%d", myName, i), "-n", f.Namespace()).Run(); err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("\n======================================[ Describe PerconaXtraDB ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "px", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n======================================[ Percona Server Log ]===================================================")
	for i := 0; i < pxReplicas; i++ {
		if err := sh.Command("/usr/bin/kubectl", "logs", fmt.Sprintf("%s-%d", pxName, i), "-n", f.Namespace()).Run(); err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("\n======================================[ Describe Nodes ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "nodes").Run(); err != nil {
		fmt.Println(err)
	}
}
