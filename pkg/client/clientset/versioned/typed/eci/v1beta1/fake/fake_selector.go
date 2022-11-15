/*
Copyright The Kubernetes Authors.

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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "eci.io/eci-profile/pkg/apis/eci/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSelectors implements SelectorInterface
type FakeSelectors struct {
	Fake *FakeEciV1beta1
}

var selectorsResource = schema.GroupVersionResource{Group: "eci.aliyun.com", Version: "v1beta1", Resource: "selectors"}

var selectorsKind = schema.GroupVersionKind{Group: "eci.aliyun.com", Version: "v1beta1", Kind: "Selector"}

// Get takes name of the selector, and returns the corresponding selector object, and an error if there is any.
func (c *FakeSelectors) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.Selector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(selectorsResource, name), &v1beta1.Selector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Selector), err
}

// List takes label and field selectors, and returns the list of Selectors that match those selectors.
func (c *FakeSelectors) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.SelectorList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(selectorsResource, selectorsKind, opts), &v1beta1.SelectorList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.SelectorList{ListMeta: obj.(*v1beta1.SelectorList).ListMeta}
	for _, item := range obj.(*v1beta1.SelectorList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested selectors.
func (c *FakeSelectors) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(selectorsResource, opts))
}

// Create takes the representation of a selector and creates it.  Returns the server's representation of the selector, and an error, if there is any.
func (c *FakeSelectors) Create(ctx context.Context, selector *v1beta1.Selector, opts v1.CreateOptions) (result *v1beta1.Selector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(selectorsResource, selector), &v1beta1.Selector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Selector), err
}

// Update takes the representation of a selector and updates it. Returns the server's representation of the selector, and an error, if there is any.
func (c *FakeSelectors) Update(ctx context.Context, selector *v1beta1.Selector, opts v1.UpdateOptions) (result *v1beta1.Selector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(selectorsResource, selector), &v1beta1.Selector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Selector), err
}

// Delete takes name of the selector and deletes it. Returns an error if one occurs.
func (c *FakeSelectors) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(selectorsResource, name, opts), &v1beta1.Selector{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSelectors) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(selectorsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.SelectorList{})
	return err
}

// Patch applies the patch and returns the patched selector.
func (c *FakeSelectors) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.Selector, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(selectorsResource, name, pt, data, subresources...), &v1beta1.Selector{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.Selector), err
}
