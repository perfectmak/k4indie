/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/perfectmak/k4indie/api/v1alpha1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

var (
	typeDeploymentAvailable = "DeploymentAvailable"
	typeDeploymentDegraded  = "DeploymentDegraded"
)

//+kubebuilder:rbac:groups=operators.k4indie.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operators.k4indie.io,resources=applications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operators.k4indie.io,resources=applications/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	appToReconcile := &operatorsv1alpha1.Application{}
	err := r.Get(ctx, req.NamespacedName, appToReconcile)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("application resource not found. ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get application")
		return ctrl.Result{}, err
	}

	if appToReconcile.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}

	result, err := r.reconcileDeployment(ctx, req, appToReconcile)
	if err != nil {
		return ctrl.Result{}, err
	}
	if result != nil {
		return *result, nil
	}

	result, err = r.reconcileService(ctx, req, appToReconcile)
	if err != nil {
		return ctrl.Result{}, err
	}
	if result != nil {
		return *result, nil
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) setApplicationReconciled(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
	log logr.Logger,
) (*reconcile.Result, error) {
	// Re-fetch the Resource before update the status
	// so that we have the latest state of the resource on the cluster and we will avoid
	// raise the issue "the object has been modified, please apply
	// your changes to the latest version and try again" which would re-trigger the reconciliation
	if err := r.Get(ctx, req.NamespacedName, appToReconcile); err != nil {
		log.Error(err, "failed to re-fetch application")
		return &reconcile.Result{}, err
	}

	meta.SetStatusCondition(
		&appToReconcile.Status.Conditions,
		metav1.Condition{
			Type:   typeDeploymentAvailable,
			Status: metav1.ConditionTrue,
			Reason: "Reconciled",
			Message: fmt.Sprintf(
				"Deployment for custom resource (%s) with %d replicas created successfully",
				appToReconcile.Name,
				appToReconcile.Spec.Replicas,
			),
		})

	if err := r.Status().Update(ctx, appToReconcile); err != nil {
		log.Error(err, "failed to update application status")
		return nil, err
	}

	return nil, nil
}

func (r *ApplicationReconciler) setApplicationReconcileError(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
	log logr.Logger,
	err error,
) (*reconcile.Result, error) {
	// Re-fetch the Resource before update the status
	// so that we have the latest state of the resource on the cluster and we will avoid
	// raise the issue "the object has been modified, please apply
	// your changes to the latest version and try again" which would re-trigger the reconciliation
	if err := r.Get(ctx, req.NamespacedName, appToReconcile); err != nil {
		log.Error(err, "failed to re-fetch application")
		return &reconcile.Result{}, err
	}

	meta.SetStatusCondition(
		&appToReconcile.Status.Conditions,
		metav1.Condition{
			Type:   typeDeploymentAvailable,
			Status: metav1.ConditionFalse,
			Reason: "ReconcileError",
			Message: fmt.Sprintf(
				"Failed to update the application resource (%s): (%s)",
				appToReconcile.Name,
				err,
			),
		})

	if err := r.Status().Update(ctx, appToReconcile); err != nil {
		log.Error(err, "failed to update application status")
		return nil, err
	}

	return nil, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1alpha1.Application{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
