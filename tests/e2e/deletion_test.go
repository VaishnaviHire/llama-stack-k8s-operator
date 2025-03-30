package e2e

import (
	"testing"

	"github.com/meta-llama/llama-stack-k8s-operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestDeletionSuite(t *testing.T) {
	if TestOpts.SkipDeletion {
		t.Skip("Skipping deletion test suite")
	}

	t.Run("should delete LlamaStackDistribution CR and cleanup resources", func(t *testing.T) {
		instance := &v1alpha1.LlamaStackDistribution{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "llamastackdistribution-sample",
				Namespace: "llama-stack-test",
			},
		}

		// Delete the instance
		err := TestEnv.Client.Delete(TestEnv.Ctx, instance)
		require.NoError(t, err)

		// Verify deployment is deleted
		deployment := &appsv1.Deployment{}
		err = TestEnv.Client.Get(TestEnv.Ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Name,
		}, deployment)
		require.Error(t, err, "Deployment should be deleted")

		// Verify service is deleted
		service := &corev1.Service{}
		err = TestEnv.Client.Get(TestEnv.Ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Name + "-service",
		}, service)
		require.Error(t, err, "Service should be deleted")

		// Verify CR is deleted
		err = TestEnv.Client.Get(TestEnv.Ctx, client.ObjectKey{
			Namespace: instance.Namespace,
			Name:      instance.Name,
		}, instance)
		require.Error(t, err, "CR should be deleted")

		// Verify no orphaned resources
		podList := &corev1.PodList{}
		err = TestEnv.Client.List(TestEnv.Ctx, podList, client.InNamespace(instance.Namespace))
		require.NoError(t, err)
		for _, pod := range podList.Items {
			require.NotEqual(t, instance.Name, pod.Labels["app"], "Found orphaned pod")
		}

		// Verify no orphaned configmaps
		configMapList := &corev1.ConfigMapList{}
		err = TestEnv.Client.List(TestEnv.Ctx, configMapList, client.InNamespace(instance.Namespace))
		require.NoError(t, err)
		for _, cm := range configMapList.Items {
			require.NotEqual(t, instance.Name, cm.Labels["app"], "Found orphaned configmap")
		}
	})
}
