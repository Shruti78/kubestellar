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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	v1alpha1 "github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// BindingLister helps list Bindings.
// All objects returned here must be treated as read-only.
type BindingLister interface {
	// List lists all Bindings in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Binding, err error)
	// Get retrieves the Binding from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Binding, error)
	BindingListerExpansion
}

// bindingLister implements the BindingLister interface.
type bindingLister struct {
	indexer cache.Indexer
}

// NewBindingLister returns a new BindingLister.
func NewBindingLister(indexer cache.Indexer) BindingLister {
	return &bindingLister{indexer: indexer}
}

// List lists all Bindings in the indexer.
func (s *bindingLister) List(selector labels.Selector) (ret []*v1alpha1.Binding, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Binding))
	})
	return ret, err
}

// Get retrieves the Binding from the index for a given name.
func (s *bindingLister) Get(name string) (*v1alpha1.Binding, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("binding"), name)
	}
	return obj.(*v1alpha1.Binding), nil
}
