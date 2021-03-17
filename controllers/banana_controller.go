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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	fruitscomv1 "github.com/i-sergienko/banana-operator-golang/api/v1"
)

// BananaReconciler reconciles a Banana object
type BananaReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const BananaFinalizer = "bananas.fruits.com/finalizer"

// +kubebuilder:rbac:groups=fruits.com,resources=bananas,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fruits.com,resources=bananas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fruits.com,resources=bananas/finalizers,verbs=update

// Compares the state specified by the Banana object against the actual cluster state, and then
// performs operations to make the cluster state reflect the state specified by the user.
func (r *BananaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("banana", req.NamespacedName)

	// Retrieve the Banana being updated
	banana := &fruitscomv1.Banana{}
	err := r.Get(ctx, req.NamespacedName, banana)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Banana not found: ignoring resource.", "namespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to retrieve Banana", "namespacedName", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if banana.GetDeletionTimestamp() == nil {
		// If deletion timestamp is not present, the resource must have been created or updated
		// Resource processing is performed in `handleCreateOrUpdate`
		return ctrl.Result{}, r.handleCreateOrUpdate(&ctx, banana, &log)
	} else {
		// If deletion timestamp is there, the resource must have been deleted
		// Additional cleanup is performed in `handleDelete`
		return ctrl.Result{}, r.handleDelete(&ctx, banana, &log)
	}
}

func (r *BananaReconciler) handleCreateOrUpdate(ctx *context.Context, banana *fruitscomv1.Banana, log *logr.Logger) error {
	if !controllerutil.ContainsFinalizer(banana, BananaFinalizer) {
		// If the finalizer is not yet present, add it
		controllerutil.AddFinalizer(banana, BananaFinalizer)
		err := r.Update(*ctx, banana)

		if err != nil {
			(*log).Error(err, "Failed to add finalizer", "bananaResource", banana)
			return err
		}
	} else if banana.Spec.Color != banana.Status.Color {
		// If spec.color != status.color, we need to "paint" the Banana resource
		// Simulate work. In a real app you'd do your useful work here - e.g. call external API, create k8s objects, etc.
		err := r.processBanana(banana, log)

		if err != nil {
			(*log).Error(err, "Failed to process Banana", "bananaResource", banana)
			return err
		}

		(*log).Info("Updating Banana Status.", "bananaResource", banana)
		err = r.Status().Update(context.Background(), banana)

		if err != nil {
			(*log).Error(err, "Failed to update Banana status", "bananaResource", banana)
			return err
		}
	}

	return nil
}

func (r *BananaReconciler) handleDelete(ctx *context.Context, banana *fruitscomv1.Banana, log *logr.Logger) error {
	(*log).Info("Banana is being deleted", "bananaResource", banana)
	if controllerutil.ContainsFinalizer(banana, BananaFinalizer) {
		// Run cleanup logic (external API calls, etc.)
		if err := r.cleanUpBanana(banana, log); err != nil {
			return err
		}

		// Remove the finalizer if cleanup was successful. Once the finalizer is removed, k8s will delete the resource from etcd.
		controllerutil.RemoveFinalizer(banana, BananaFinalizer)
		err := r.Update(*ctx, banana)

		if err != nil {
			(*log).Error(err, "Failed remove finalizer", "bananaResource", banana)
			return err
		}
	}
	return nil
}

func (r *BananaReconciler) processBanana(banana *fruitscomv1.Banana, log *logr.Logger) error {
	(*log).Info("Painting Banana", "bananaResource", banana)
	// Pretend that painting the Banana takes 3 seconds - e.g. external API calls take that much
	time.Sleep(3 * time.Second)
	banana.Status.Color = banana.Spec.Color
	(*log).Info("Banana painted successfully", "bananaResource", banana)
	return nil
}

func (r *BananaReconciler) cleanUpBanana(banana *fruitscomv1.Banana, log *logr.Logger) error {
	(*log).Info("Cleaning up Banana", "bananaResource", banana)
	time.Sleep(3 * time.Second) // pretend that some external API calls or other cleanup take 3 seconds
	(*log).Info("Banana cleaned up successfully", "bananaResource", banana)
	return nil
}

func (r *BananaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fruitscomv1.Banana{}).
		Complete(r)
}
