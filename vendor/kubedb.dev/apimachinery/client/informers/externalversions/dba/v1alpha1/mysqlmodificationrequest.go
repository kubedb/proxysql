/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	dbav1alpha1 "kubedb.dev/apimachinery/apis/dba/v1alpha1"
	versioned "kubedb.dev/apimachinery/client/clientset/versioned"
	internalinterfaces "kubedb.dev/apimachinery/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kubedb.dev/apimachinery/client/listers/dba/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// MySQLModificationRequestInformer provides access to a shared informer and lister for
// MySQLModificationRequests.
type MySQLModificationRequestInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.MySQLModificationRequestLister
}

type mySQLModificationRequestInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewMySQLModificationRequestInformer constructs a new informer for MySQLModificationRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMySQLModificationRequestInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMySQLModificationRequestInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredMySQLModificationRequestInformer constructs a new informer for MySQLModificationRequest type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMySQLModificationRequestInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DbaV1alpha1().MySQLModificationRequests().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DbaV1alpha1().MySQLModificationRequests().Watch(options)
			},
		},
		&dbav1alpha1.MySQLModificationRequest{},
		resyncPeriod,
		indexers,
	)
}

func (f *mySQLModificationRequestInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMySQLModificationRequestInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *mySQLModificationRequestInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&dbav1alpha1.MySQLModificationRequest{}, f.defaultInformer)
}

func (f *mySQLModificationRequestInformer) Lister() v1alpha1.MySQLModificationRequestLister {
	return v1alpha1.NewMySQLModificationRequestLister(f.Informer().GetIndexer())
}
