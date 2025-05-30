package deploy

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	llamav1alpha1 "github.com/llamastack/llama-stack-k8s-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KustomizeClient handles deployment of resources using kustomize manifests
type KustomizeClient struct {
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

// NewKustomizeClient creates a new KustomizeClient
func NewKustomizeClient(client client.Client, scheme *runtime.Scheme, log logr.Logger) *KustomizeClient {
	return &KustomizeClient{
		client: client,
		scheme: scheme,
		log:    log,
	}
}

// ApplyKustomize applies resources from a kustomize directory
func (k *KustomizeClient) ApplyKustomize(ctx context.Context, instance *llamav1alpha1.LlamaStackDistribution, kustomizeDir string) error {
	// Create kustomize options
	opts := krusty.MakeDefaultOptions()
	kustomizer := krusty.MakeKustomizer(opts)

	// Create filesystem
	fs := filesys.MakeFsOnDisk()

	// Build the kustomization
	resMap, err := kustomizer.Run(fs, kustomizeDir)
	if err != nil {
		return fmt.Errorf("failed to build kustomization: %w", err)
	}

	// Apply each resource
	return k.applyResources(ctx, instance, resMap)
}

// applyResources applies each resource from the kustomize build
func (k *KustomizeClient) applyResources(ctx context.Context, instance *llamav1alpha1.LlamaStackDistribution, resMap resmap.ResMap) error {
	for _, res := range resMap.Resources() {
		// Convert to unstructured
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   res.GetGvk().Group,
			Version: res.GetGvk().Version,
			Kind:    res.GetGvk().Kind,
		})

		// Get the YAML
		yamlBytes, err := res.AsYAML()
		if err != nil {
			return fmt.Errorf("failed to get YAML for resource: %w", err)
		}

		// Unmarshal into unstructured
		if err := yaml.Unmarshal(yamlBytes, obj); err != nil {
			return fmt.Errorf("failed to unmarshal resource: %w", err)
		}

		// Set namespace if not set
		if obj.GetNamespace() == "" {
			obj.SetNamespace(instance.Namespace)
		}

		// Set owner reference
		if err := k.setOwnerReference(instance, obj); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		// Apply the resource
		if err := k.applyResource(ctx, obj); err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName(), err)
		}
	}

	return nil
}

// applyResource applies a single resource
func (k *KustomizeClient) applyResource(ctx context.Context, obj client.Object) error {
	// Check if resource exists
	existing := obj.DeepCopyObject().(client.Object)
	err := k.client.Get(ctx, types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, existing)

	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to get resource: %w", err)
		}
		// Create if not found
		k.log.Info("Creating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
		return k.client.Create(ctx, obj)
	}

	// Update if found
	obj.SetResourceVersion(existing.GetResourceVersion())
	k.log.Info("Updating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
	return k.client.Update(ctx, obj)
}

// setOwnerReference sets the owner reference for a resource
func (k *KustomizeClient) setOwnerReference(instance *llamav1alpha1.LlamaStackDistribution, obj client.Object) error {
	return controllerutil.SetControllerReference(instance, obj, k.scheme)
}
