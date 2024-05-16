/*
Copyright The KubeStellar Authors.

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
	"context"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"

	controlv1alpha1 "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	versioned "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
)

// BindingInformer provides access to a shared informer and lister for
// Bindings.
type BindingInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.BindingLister
}

type bindingInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewBindingInformer constructs a new informer for Binding type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBindingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBindingInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredBindingInformer constructs a new informer for Binding type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBindingInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ControlV1alpha1().Bindings().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ControlV1alpha1().Bindings().Watch(context.TODO(), options)
			},
		},
		&controlv1alpha1.Binding{},
		resyncPeriod,
		indexers,
	)
}

func (f *bindingInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBindingInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *bindingInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&controlv1alpha1.Binding{}, f.defaultInformer)
}

func (f *bindingInformer) Lister() v1alpha1.BindingLister {
	return v1alpha1.NewBindingLister(f.Informer().GetIndexer())
}
