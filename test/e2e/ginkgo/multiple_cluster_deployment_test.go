/*
Copyright 2023 The KubeStellar Authors.

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

package e2e

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/test/util"
)

const (
	ns = "nginx"
)

var _ = ginkgo.Describe("end to end testing", func() {
	ginkgo.BeforeEach(func(ctx context.Context) {
		// Cleanup the WDS, create 1 deployment and 1 binding policy.
		util.CleanupWDS(ctx, wds, ksWds, ns)
		util.CreateDeployment(ctx, wds, ns, "nginx",
			map[string]string{
				"app.kubernetes.io/name":         "nginx",
				"test.kubestellar.io/test-label": "here",
			})
		util.CreateBindingPolicy(ctx, ksWds, "nginx",
			[]metav1.LabelSelector{
				{MatchLabels: map[string]string{"location-group": "edge"}},
			},
			[]ksapi.DownsyncObjectTest{
				{ObjectSelectors: []metav1.LabelSelector{
					{MatchLabels: map[string]string{"app.kubernetes.io/name": "nginx"}},
				}}})
	})

	ctx := context.Background()

	ginkgo.Context("multiple WECs", func() {
		ginkgo.It("propagates deployment to the WECs while applying CustomTransform", func() {
			util.CreateCustomTransform(ctx, ksWds, "test", "apps", "deployments", `$.metadata.labels["test.kubestellar.io/test-label"]`)
			testLabelAbsent := func(deployment *appsv1.Deployment) string {
				if _, has := deployment.Labels["test.kubestellar.io/test-label"]; has {
					return "it has the 'test.kubestellar.io/test-label' label"
				}
				return ""
			}
			util.ValidateNumDeployments(ctx, wec1, ns, 1, testLabelAbsent)
			util.ValidateNumDeployments(ctx, wec2, ns, 1, testLabelAbsent)
		})

		ginkgo.It("updates objects on the WECs following an update on the WDS", func() {
			patch := []byte(`{"spec":{"replicas": 2}}`)
			_, err := wds.AppsV1().Deployments(ns).Patch(
				ctx, "nginx", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())

			util.ValidateNumDeploymentReplicas(ctx, wec1, ns, 2)
			util.ValidateNumDeploymentReplicas(ctx, wec2, ns, 2)
		})

		ginkgo.It("handles changes in bindingpolicy ObjectSelector", func() {
			ginkgo.By("deletes WEC objects when bindingpolicy ObjectSelector stops matching")
			patch := []byte(`{"spec": {"downsync": [{"objectSelectors": [{"matchLabels": {"app.kubernetes.io/name": "invalid"}}]}]}}`)
			_, err := ksWds.ControlV1alpha1().BindingPolicies().Patch(
				ctx, "nginx", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)

			ginkgo.By("creates WEC objects when bindingpolicy ObjectSelector matches")
			patch = []byte(`{"spec": {"downsync": [{"objectSelectors": [{"matchLabels": {"app.kubernetes.io/name": "nginx"}}]}]}}`)
			_, err = ksWds.ControlV1alpha1().BindingPolicies().Patch(
				ctx, "nginx", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})

		ginkgo.It("handles changes in workload object labels", func() {
			ginkgo.By("deletes WEC objects when workload object labels stop matching")
			patch := []byte(`{"metadata": {"labels": {"app.kubernetes.io/name": "not-me"}}}`)
			_, err := wds.AppsV1().Deployments("nginx").Patch(
				ctx, "nginx", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)

			ginkgo.By("creates WEC objects when workload object labels resume matching")
			patch = []byte(`{"metadata": {"labels": {"app.kubernetes.io/name": "nginx"}}}`)
			_, err = wds.AppsV1().Deployments("nginx").Patch(
				ctx, "nginx", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})

		ginkgo.It("handles multiple bindingpolicies with overlapping matches", func() {
			ginkgo.By("creates a second bindingpolicy with overlapping matches")
			util.CreateBindingPolicy(ctx, ksWds, "nginx-2",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"location-group": "edge"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"app.kubernetes.io/name": "nginx"}},
					}}})
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)

			ginkgo.By("delete the second bindingpolicy")
			err := ksWds.ControlV1alpha1().BindingPolicies().Delete(ctx, "nginx-2", metav1.DeleteOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})

		ginkgo.It("deletes WEC objects when wds deployment is deleted", func() {
			util.DeleteDeployment(ctx, wds, ns, "nginx")
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)
		})

		ginkgo.It("deletes WEC objects when BindingPolicy is deleted", func() {
			err := ksWds.ControlV1alpha1().BindingPolicies().Delete(ctx, "nginx", metav1.DeleteOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)
		})

		ginkgo.It("shards a wrapped workload when needed", func() {
			util.DeleteDeployment(ctx, wds, ns, "nginx")
			ginkgo.DeferCleanup(func() {
				patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"args":["--max-size-wrapped-object=512000","--transport-kubeconfig=/mnt/shared/transport-kubeconfig","--wds-kubeconfig=/mnt/shared/wds-kubeconfig","--wds-name=wds1","-v=4"],"name":"transport-controller"}]}}}}`)
				_, err := coreCluster.AppsV1().Deployments("wds1-system").Patch(ctx, "transport-controller", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
				gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.By("set max size of wrapped object to 1000 bytes")
			patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"args":["--max-size-wrapped-object=1000","--transport-kubeconfig=/mnt/shared/transport-kubeconfig","--wds-kubeconfig=/mnt/shared/wds-kubeconfig","--wds-name=wds1","-v=4"],"name":"transport-controller"}]}}}}`)
			_, err := coreCluster.AppsV1().Deployments("wds1-system").Patch(ctx, "transport-controller", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			var names = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
			for _, i := range names {
				util.CreateDeployment(ctx, wds, ns, i,
					map[string]string{
						"label1": "test1",
					})
			}
			util.CreateBindingPolicy(ctx, ksWds, "multipledep",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{
							"label1": "test1",
						}},
					}}})

			util.ValidateNumDeployments(ctx, wec1, ns, 10)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)

			ginkgo.By("update to bindingpolicy object with sharded wrapped workloads updates deployments")
			patch = []byte(`{"spec": {"clusterSelectors": [{"matchLabels": {"name": "cluster2"}}]}}`)
			_, err = ksWds.ControlV1alpha1().BindingPolicies().Patch(ctx, "multipledep", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 10)

			ginkgo.By("delete of bindingpolicy with sharded wrapped workloads deletes deployments")
			util.DeleteBindingPolicy(ctx, ksWds, "multipledep")
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)
		})

		// 1. Create object 1 with label A and label B, object 2 with label B,
		// and a Placement that matches labels A AND B, and verify that only
		// object 1 matches and is delivered.
		// 2. Patch the cluster selector and expect objects to be removed and downsynced correctly.
		ginkgo.It("downsync objects that fully match on object and cluster selector", func() {
			ginkgo.By("create two deployments and a bindingpolicy that matches only one")
			util.DeleteDeployment(ctx, wds, ns, "nginx")
			util.CreateDeployment(ctx, wds, ns, "one",
				map[string]string{
					"label1": "test1",
				})
			util.CreateDeployment(ctx, wds, ns, "two",
				map[string]string{
					"label1": "test1",
					"label2": "test2",
				})
			util.CreateBindingPolicy(ctx, ksWds, "both-labels",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{
							"label1": "test1",
							"label2": "test2",
						}},
					}}})
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)

			ginkgo.By("patch the cluster selector")
			patch := []byte(`{"spec": {"clusterSelectors": [{"matchLabels": {"name": "cluster2"}}]}}`)
			_, err := ksWds.ControlV1alpha1().BindingPolicies().Patch(
				ctx, "both-labels", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})

		ginkgo.It("handles clusterSelector with MatchLabels and MatchExpressions", func() {
			util.DeleteDeployment(ctx, wds, ns, "nginx")
			util.CreateDeployment(ctx, wds, ns, "one",
				map[string]string{
					"label1": "A",
				})
			util.CreateDeployment(ctx, wds, ns, "two",
				map[string]string{
					"label1": "B",
				})
			util.CreateDeployment(ctx, wds, ns, "three",
				map[string]string{
					"label1": "C",
				})
			util.CreateBindingPolicy(ctx, ksWds, "two-cluster-selectors",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
					{MatchExpressions: []metav1.LabelSelectorRequirement{
						{Key: "name", Operator: metav1.LabelSelectorOpIn, Values: []string{"cluster1", "cluster2", "cluster3"}},
					}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"label1": "A"}},
						{MatchLabels: map[string]string{"label1": "B"}},
					}}})
			util.ValidateNumDeployments(ctx, wec1, ns, 2)
			util.ValidateNumDeployments(ctx, wec2, ns, 2)
		})

		ginkgo.It("downsync based on object labels and object name", func() {
			util.DeleteDeployment(ctx, wds, ns, "nginx")
			util.CreateDeployment(ctx, wds, ns, "one",
				map[string]string{
					"label1": "A",
				})
			util.CreateDeployment(ctx, wds, ns, "two",
				map[string]string{
					"label1": "B",
				})
			util.CreateDeployment(ctx, wds, ns, "three",
				map[string]string{
					"label1": "C",
				})
			util.CreateBindingPolicy(ctx, ksWds, "test-name-and-label",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"location-group": "edge"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectNames: []string{"three"}},
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"label1": "C"}},
					}}})
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})
	})

	ginkgo.Context("singleton status testing", func() {
		ginkgo.It("sets (or deletes) singleton status when a singleton bindingpolicy/deployment is created (or deleted)", func() {
			util.DeleteDeployment(ctx, wds, ns, "nginx") // we don't have to delete nginx
			util.CreateDeployment(ctx, wds, ns, "nginx-singleton",
				map[string]string{
					"app.kubernetes.io/name": "nginx-singleton",
				})
			util.CreateBindingPolicy(ctx, ksWds, "nginx-singleton",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"app.kubernetes.io/name": "nginx-singleton"}},
					}}})
			patch := []byte(`{"spec":{"wantSingletonReportedState": true}}`)
			_, err := ksWds.ControlV1alpha1().BindingPolicies().Patch(
				ctx, "nginx-singleton", types.MergePatchType, patch, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)
			util.ValidateSingletonStatus(ctx, wds, ns, "nginx-singleton")
			patch_again := []byte(`{"spec":{"clusterSelectors":[{"matchLabels":{"name":"CelestialNexus"}}]}}`)
			_, err = ksWds.ControlV1alpha1().BindingPolicies().Patch(
				ctx, "nginx-singleton", types.MergePatchType, patch_again, metav1.PatchOptions{})
			gomega.ExpectWithOffset(1, err).NotTo(gomega.HaveOccurred())
			util.ValidateNumDeployments(ctx, wec1, ns, 0)
			util.ValidateNumDeployments(ctx, wec2, ns, 0)
			util.ValidateSingletonStatusZeroValue(ctx, wds, ns, "nginx-singleton")
		})
	})

	ginkgo.Context("object cleaning", func() {
		ginkgo.It("properly starts a service", func() {
			util.CreateService(ctx, wds, ns, "hello-service", "hello-service")
			util.CreateBindingPolicy(ctx, ksWds, "hello-service",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"app.kubernetes.io/name": "hello-service"}},
					}}})
			util.ValidateNumServices(ctx, wec1, ns, 1)
		})
		ginkgo.It("properly starts a job with metadata.generateName", func() {
			util.CreateJob(ctx, wds, ns, "hello-job", "hello-job")
			util.CreateBindingPolicy(ctx, ksWds, "hello-job",
				[]metav1.LabelSelector{
					{MatchLabels: map[string]string{"name": "cluster1"}},
				},
				[]ksapi.DownsyncObjectTest{
					{ObjectSelectors: []metav1.LabelSelector{
						{MatchLabels: map[string]string{"app.kubernetes.io/name": "hello-job"}},
					}}})
			util.ValidateNumJobs(ctx, wec1, ns, 1)
		})
	})

	ginkgo.Context("resiliency testing", func() {
		ginkgo.It("survives WDS coming down", func() {
			util.DeletePods(ctx, coreCluster, "wds1-system", "kubestellar")
			util.DeletePods(ctx, coreCluster, "wds1-system", "transport")
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
			util.Expect1PodOfEach(ctx, coreCluster, "wds1-system", "kubestellar-controller-manager", "transport-controller")
		})

		ginkgo.It("survives kubeflex coming down", func() {
			util.DeletePods(ctx, coreCluster, "kubeflex-system", "")
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
			util.Expect1PodOfEach(ctx, coreCluster, "kubeflex-system", "kubeflex-controller-manager", "postgres-postgresql-0")
		})

		ginkgo.It("survives ITS vcluster coming down", func() {
			util.DeletePods(ctx, coreCluster, "its1-system", "vcluster")
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)
		})

		ginkgo.It("survives everything coming down", func() {
			ginkgo.By("kill as many pods as possible")
			util.DeletePods(ctx, coreCluster, "wds1-system", "kubestellar")
			util.DeletePods(ctx, coreCluster, "wds1-system", "transport")
			util.DeletePods(ctx, coreCluster, "kubeflex-system", "")
			util.DeletePods(ctx, coreCluster, "its1-system", "vcluster")
			util.ValidateNumDeployments(ctx, wec1, ns, 1)
			util.ValidateNumDeployments(ctx, wec2, ns, 1)

			ginkgo.By("test that a new deployment still gets downsynced")
			util.CreateDeployment(ctx, wds, ns, "nginx-2",
				map[string]string{
					"app.kubernetes.io/name": "nginx",
				})
			util.ValidateNumDeployments(ctx, wec1, ns, 2)
			util.ValidateNumDeployments(ctx, wec2, ns, 2)
			util.Expect1PodOfEach(ctx, coreCluster, "wds1-system", "kubestellar-controller-manager", "transport-controller")
			util.Expect1PodOfEach(ctx, coreCluster, "kubeflex-system", "kubeflex-controller-manager", "postgres-postgresql-0")
			util.Expect1PodOfEach(ctx, coreCluster, "its1-system", "vcluster")
		})
	})

})
