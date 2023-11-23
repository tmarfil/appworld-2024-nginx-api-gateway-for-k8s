// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/nginxinc/kubernetes-ingress/pkg/apis/configuration/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePolicies implements PolicyInterface
type FakePolicies struct {
	Fake *FakeK8sV1alpha1
	ns   string
}

var policiesResource = v1alpha1.SchemeGroupVersion.WithResource("policies")

var policiesKind = v1alpha1.SchemeGroupVersion.WithKind("Policy")

// Get takes name of the policy, and returns the corresponding policy object, and an error if there is any.
func (c *FakePolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Policy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(policiesResource, c.ns, name), &v1alpha1.Policy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Policy), err
}

// List takes label and field selectors, and returns the list of Policies that match those selectors.
func (c *FakePolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(policiesResource, policiesKind, c.ns, opts), &v1alpha1.PolicyList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PolicyList{ListMeta: obj.(*v1alpha1.PolicyList).ListMeta}
	for _, item := range obj.(*v1alpha1.PolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested policies.
func (c *FakePolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(policiesResource, c.ns, opts))

}

// Create takes the representation of a policy and creates it.  Returns the server's representation of the policy, and an error, if there is any.
func (c *FakePolicies) Create(ctx context.Context, policy *v1alpha1.Policy, opts v1.CreateOptions) (result *v1alpha1.Policy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(policiesResource, c.ns, policy), &v1alpha1.Policy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Policy), err
}

// Update takes the representation of a policy and updates it. Returns the server's representation of the policy, and an error, if there is any.
func (c *FakePolicies) Update(ctx context.Context, policy *v1alpha1.Policy, opts v1.UpdateOptions) (result *v1alpha1.Policy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(policiesResource, c.ns, policy), &v1alpha1.Policy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Policy), err
}

// Delete takes name of the policy and deletes it. Returns an error if one occurs.
func (c *FakePolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(policiesResource, c.ns, name, opts), &v1alpha1.Policy{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(policiesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PolicyList{})
	return err
}

// Patch applies the patch and returns the patched policy.
func (c *FakePolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Policy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(policiesResource, c.ns, name, pt, data, subresources...), &v1alpha1.Policy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Policy), err
}
