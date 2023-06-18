package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/perfectmak/k4indie/api/v1alpha1"
	"github.com/perfectmak/k4indie/internal/controller/resolvers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileDeployment(
	ctx context.Context,
	req reconcile.Request,
	appToReconcile *operatorsv1alpha1.Application,
) (*reconcile.Result, error) {
	log := log.FromContext(ctx)
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
					Type:   typeDeploymentAvailable,
					Status: metav1.ConditionFalse,
					Reason: "ReconcileError",
					Message: fmt.Sprintf(
						"Failed to create deployment: %s",
						err.Error(),
					),
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

	return r.setApplicationReconciled(ctx, req, appToReconcile, log)
}

func (r *ApplicationReconciler) createDeployment(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
) (*appsv1.Deployment, error) {
	log := log.FromContext(ctx)

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
	labels := map[string]string{
		"app.kubernetes.io/instance":   fmt.Sprintf("application-%s", appToReconcile.Name),
		"app.kubernetes.io/version":    appToReconcile.Spec.Runtime.Image.Tag(),
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
						Image:           appToReconcile.Spec.Runtime.Image.String(),
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

	return nil
}
