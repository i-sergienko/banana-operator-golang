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
	"flag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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

func startController() client.Client {
	var scheme = runtime.NewScheme()
	var setupLog = ctrl.Log.WithName("setup")
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "fef4abb1.fruits.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&BananaReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Banana"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Banana")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	go func() {
		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	}()

	return mgr.GetClient()
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

	k8sClient = startController()
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
			Spec: fruitscomv1.BananaSpec{Color: "white"},
		}
		banana.Name = "white-banana"
		banana.Namespace = "default"

		err := k8sClient.Create(context.Background(), &banana)
		Expect(err).NotTo(HaveOccurred())

		bananas := fruitscomv1.BananaList{}
		err = k8sClient.List(context.Background(), &bananas, client.InNamespace("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(len(bananas.Items)).To(BeEquivalentTo(1))
		Expect(bananas.Items[0].Name).To(BeEquivalentTo("white-banana"))
		Expect(bananas.Items[0].Spec.Color).To(BeEquivalentTo("white"))
		Expect(bananas.Items[0].Status.Color).NotTo(BeEquivalentTo("white"))
	})

	It("New Bananas are painted by the controller", func() {
		time.Sleep(6 * time.Second)

		banana := fruitscomv1.Banana{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Namespace: "default",
			Name:      "white-banana",
		}, &banana)
		Expect(err).NotTo(HaveOccurred())
		Expect(banana.Name).To(BeEquivalentTo("white-banana"))
		Expect(banana.Spec.Color).To(BeEquivalentTo("white"))
		Expect(banana.Status.Color).To(BeEquivalentTo("white"))
	})
})
