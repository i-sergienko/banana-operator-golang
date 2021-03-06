/*
Copyright 2021.

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

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	fruitscomv1 "github.com/i-sergienko/banana-operator-golang/api/v1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = fruitscomv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Banana lifecycle", func() {
	It("Before we create a Banana, there aren't any", func() {
		bananas := fruitscomv1.BananaList{}

		err := k8sClient.List(context.Background(), &bananas, client.InNamespace("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(len(bananas.Items)).To(BeEquivalentTo(0))
	})

	It("A newly created Banana is not painted before processing", func() {
		banana := fruitscomv1.Banana{
			Spec: fruitscomv1.BananaSpec{Color: "yellow"},
		}
		banana.Name = "yellow-banana"
		banana.Namespace = "default"

		err := k8sClient.Create(context.Background(), &banana)
		Expect(err).NotTo(HaveOccurred())

		bananas := fruitscomv1.BananaList{}
		err = k8sClient.List(context.Background(), &bananas, client.InNamespace("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(len(bananas.Items)).To(BeEquivalentTo(1))
		Expect(bananas.Items[0].Name).To(BeEquivalentTo("yellow-banana"))
		Expect(bananas.Items[0].Spec.Color).To(BeEquivalentTo("yellow"))
		Expect(bananas.Items[0].Status.Color).NotTo(BeEquivalentTo("yellow"))
	})

	It("New Bananas are painted by the controller", func() {
		time.Sleep(5 * time.Second)

		banana := fruitscomv1.Banana{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "default",
			Name:      "yellow-banana",
		}, &banana)
		Expect(err).NotTo(HaveOccurred())
		Expect(banana.Name).To(BeEquivalentTo("yellow-banana"))
		Expect(banana.Spec.Color).To(BeEquivalentTo("yellow"))
		Expect(banana.Status.Color).To(BeEquivalentTo("yellow"))
	})

	It("Deleted bananas go through cleanup logic", func() {
		banana := fruitscomv1.Banana{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "default",
			Name:      "yellow-banana",
		}, &banana)
		Expect(err).NotTo(HaveOccurred())
		Expect(banana.GetDeletionTimestamp()).To(BeNil())

		err = k8sClient.Delete(context.Background(), &banana)
		Expect(err).NotTo(HaveOccurred())

		err = k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "default",
			Name:      "yellow-banana",
		}, &banana)
		Expect(err).NotTo(HaveOccurred())
		Expect(controllerutil.ContainsFinalizer(&banana, BananaFinalizer)).To(BeTrue())
		Expect(banana.GetDeletionTimestamp()).NotTo(BeNil())

		time.Sleep(5 * time.Second)
		err = k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "default",
			Name:      "yellow-banana",
		}, &banana)
		Expect(err).To(HaveOccurred())
		Expect(errors.IsNotFound(err)).To(BeTrue())
	})
})
