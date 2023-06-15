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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	"github.com/perfectmak/k4indie/internal/controller/resolvers"
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
	log := log.FromContext(ctx).WithValues("application", req.NamespacedName.String())

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

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) reconcileDeployment(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
) (*reconcile.Result, error) {
	log := log.FromContext(ctx).WithValues("application", req.NamespacedName.String())
	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, req.NamespacedName, deploy)

	deployment := &appsv1.Deployment{}
	if err != nil && apierrors.IsNotFound(err) {
		deployment, err = r.createDeployment(ctx, appToReconcile)
		if err != nil {
			log.Error(err, "failed to create deployment")

			meta.SetStatusCondition(
				&appToReconcile.Status.Conditions,
				metav1.Condition{
					Type:    typeDeploymentAvailable,
					Status:  metav1.ConditionFalse,
					Reason:  "ReconcileError",
					Message: "Failed to create deployment",
				},
			)

			if err := r.Status().Update(ctx, appToReconcile); err != nil {
				log.Error(err, "failed to update application status")
				return nil, err
			}

			return nil, err
		}

		// Re=enqueue the request to check the status of the deployment
		return &reconcile.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "failed to get deployment")
		return nil, err
	}

	err = r.updateDeploymentSpec(ctx, appToReconcile, deployment, log)
	if err != nil {
		return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
	}

	if err := r.Update(ctx, deployment); err != nil {
		log.Error(err, "failed to update deployment")

		return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
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
		log.Error(err, "failed to update Memcached status")
		return nil, err
	}

	return nil, nil
}

func (r *ApplicationReconciler) createDeployment(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
) (*appsv1.Deployment, error) {
	log := log.FromContext(ctx).WithValues("application", appToReconcile.Name)

	deployment, err := r.buildDeployment(ctx, appToReconcile)
	if err != nil {
		return nil, err
	}

	log.Info(
		"creating deployment",
		"deployment.name", deployment.Name,
		"deployment.namespace", deployment.Namespace,
	)
	if err := r.Create(ctx, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *ApplicationReconciler) buildDeployment(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
) (*appsv1.Deployment, error) {
	imageTag := appToReconcile.Spec.Runtime.Image
	labels := map[string]string{
		"app.kubernetes.io/instance":   fmt.Sprintf("application-%s", appToReconcile.Name),
		"app.kubernetes.io/version":    imageTag,
		"app.kubernetes.io/part-of":    "k4indie-operator",
		"app.kubernetes.io/created-by": "controller-manager",
	}
	resourcesRequired, err := resolvers.GetResourcesForRuntimeSize(
		appToReconcile.Spec.Runtime.Size,
	)
	if err != nil {
		return nil, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appToReconcile.Name,
			Namespace: appToReconcile.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &appToReconcile.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &[]bool{true}[0],
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{{
						Image:           imageTag,
						Name:            "application",
						ImagePullPolicy: corev1.PullIfNotPresent,
						SecurityContext: &corev1.SecurityContext{
							RunAsNonRoot:             &[]bool{true}[0],
							AllowPrivilegeEscalation: &[]bool{false}[0],
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
						Ports:     appToReconcile.Spec.Endpoints.AsContainerPorts(),
						Command:   appToReconcile.Spec.LaunchCommand,
						Resources: resourcesRequired,
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(appToReconcile, deployment, r.Scheme); err != nil {
		return nil, err
	}
	return deployment, nil
}

func (r *ApplicationReconciler) updateDeploymentSpec(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
	deployment *appsv1.Deployment,
	log logr.Logger,
) error {
	newDeployment, err := r.buildDeployment(ctx, appToReconcile)
	if err != nil {
		return err
	}

	newDeployment.DeepCopyInto(deployment)

	// // Reconcile Runtime
	// if appToReconcile.Spec.Runtime.Image != applicationContainer.Image {
	// 	log.Info(
	// 		"updating deployment image",
	// 		"old-image", deployment.Spec.Template.Spec.Containers[0].Image,
	// 		"new-image", appToReconcile.Spec.Runtime.Image,
	// 	)
	// 	applicationContainer.Image = appToReconcile.Spec.Runtime.Image
	// }

	// resourceRequirements, err := resolvers.GetResourcesForRuntimeSize(appToReconcile.Spec.Runtime.Size)
	// if err != nil {
	// 	return err
	// }
	// resourceRequirements.DeepCopyInto(&applicationContainer.Resources)

	// // Reconcile Endpoints
	// applicationContainer.Ports = appToReconcile.Spec.Endpoints.AsContainerPorts()

	// // Reconcile Launch Commands

	// // Reconcile Replicas
	// if appToReconcile.Spec.Replicas != *deployment.Spec.Replicas {
	// 	log.Info(
	// 		"updating deployment replicas",
	// 		"old-replicas", deployment.Spec.Replicas,
	// 		"new-replicas", appToReconcile.Spec.Replicas,
	// 	)
	// 	deployment.Spec.Replicas = &appToReconcile.Spec.Replicas
	// }

	return nil
}

func (r *ApplicationReconciler) setApplicationReconcileError(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
	log logr.Logger,
	err error,
) (*reconcile.Result, error) {
	// Re-fetch the memcached Custom Resource before update the status
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
		log.Error(err, "failed to update Memcached status")
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
