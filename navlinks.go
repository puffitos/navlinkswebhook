package main

import (
	"context"
	"time"

	v1 "github.com/rancher/rancher/pkg/apis/ui.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	rest "k8s.io/client-go/rest"
)

// NavLinksGetter has a method to return a NavLinkInterface.
// A group's client should implement this interface.
type NavLinksGetter interface {
	NavLinks() NavLinkInterface
}

// NavLinkInterface has methods to work with NavLink resources.
type NavLinkInterface interface {
	Create(ctx context.Context, navlink *v1.NavLink, opts metav1.CreateOptions) (*v1.NavLink, error)
	Update(ctx context.Context, navlink *v1.NavLink, opts metav1.UpdateOptions) (*v1.NavLink, error)
	UpdateStatus(ctx context.Context, navlink *v1.NavLink, opts metav1.UpdateOptions) (*v1.NavLink, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.NavLink, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.NavLinkList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NavLink, err error)
	NavlinkExpansion
}

// navlinks implements NavLinkInterface
type navlinks struct {
	client rest.Interface
}

// newNavLinks returns a NavLinks
func newNavLinks(c *UiV1Client) *navlinks {
	return &navlinks{
		client: c.RESTClient(),
	}
}

// Get takes name of the navlink, and returns the corresponding navlink object, and an error if there is any.
func (c *navlinks) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.NavLink, err error) {
	result = &v1.NavLink{}
	err = c.client.Get().
		Resource("navlinks").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NavLinks that match those selectors.
func (c *navlinks) List(ctx context.Context, opts metav1.ListOptions) (result *v1.NavLinkList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.NavLinkList{}
	err = c.client.Get().
		Resource("navlinks").
		VersionedParams(&opts, ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested navlinks.
func (c *navlinks) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("navlinks").
		VersionedParams(&opts, ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a navlink and creates it.  Returns the server's representation of the navlink, and an error, if there is any.
func (c *navlinks) Create(ctx context.Context, navlink *v1.NavLink, opts metav1.CreateOptions) (result *v1.NavLink, err error) {
	result = &v1.NavLink{}
	err = c.client.Post().
		Resource("navlinks").
		VersionedParams(&opts, ParameterCodec).
		Body(navlink).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a navlink and updates it. Returns the server's representation of the navlink, and an error, if there is any.
func (c *navlinks) Update(ctx context.Context, navlink *v1.NavLink, opts metav1.UpdateOptions) (result *v1.NavLink, err error) {
	result = &v1.NavLink{}
	err = c.client.Put().
		Resource("navlinks").
		Name(navlink.Name).
		VersionedParams(&opts, ParameterCodec).
		Body(navlink).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *navlinks) UpdateStatus(ctx context.Context, navlink *v1.NavLink, opts metav1.UpdateOptions) (result *v1.NavLink, err error) {
	result = &v1.NavLink{}
	err = c.client.Put().
		Resource("navlinks").
		Name(navlink.Name).
		SubResource("status").
		VersionedParams(&opts, ParameterCodec).
		Body(navlink).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the navlink and deletes it. Returns an error if one occurs.
func (c *navlinks) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("navlinks").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *navlinks) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("navlinks").
		VersionedParams(&listOpts, ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched navlink.
func (c *navlinks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NavLink, err error) {
	result = &v1.NavLink{}
	err = c.client.Patch(pt).
		Resource("navlinks").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
