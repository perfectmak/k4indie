package controller

import (
	"context"

	operatorsv1alpha1 "github.com/perfectmak/k4indie/api/v1alpha1"
	"github.com/perfectmak/k4indie/internal/controller/resolvers"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ApplicationReconciler) reconcileIngress(
	ctx context.Context,
	req ctrl.Request,
	appToReconcile *operatorsv1alpha1.Application,
) (*reconcile.Result, error) {
	log := log.FromContext(ctx)
	ingress := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ingress)

	domainEndpoint := resolvers.EndpointsWithDomains(&appToReconcile.Spec.Endpoints)
	noEndpointsDefined := len(domainEndpoint) == 0

	if err != nil && apierrors.IsNotFound(err) {
		if noEndpointsDefined {
			return nil, nil
		}

		ingress, err = r.createIngress(ctx, appToReconcile, domainEndpoint)
	} else if err != nil {
		log.Error(err, "failed to get existing ingress")
		return nil, err
	}

	if noEndpointsDefined {
		log.Info("deleting ingress")
		err := r.Delete(ctx, ingress)
		if err != nil {
			return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
		}

		return nil, nil
	}

	log.Info("updating ingress")
	err = r.updateIngressSpec(ctx, appToReconcile, domainEndpoint, ingress)
	if err != nil {
		return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
	}

	if err := r.Update(ctx, ingress); err != nil {
		log.Error(err, "failed to update service")
		return r.setApplicationReconcileError(ctx, req, appToReconcile, log, err)
	}

	return nil, nil
}

func (r *ApplicationReconciler) createIngress(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
	endpoints []operatorsv1alpha1.ApplicationEndpoint,
) (*networkingv1.Ingress, error) {
	log := log.FromContext(ctx)

	ingress, err := r.buildIngress(ctx, appToReconcile, endpoints)
	if err != nil {
		return nil, err
	}

	log.Info("creating ingress")

	err = r.Create(ctx, ingress)
	if err != nil {
		return nil, err
	}

	return ingress, nil
}

func (r *ApplicationReconciler) buildIngress(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
	endpoints []operatorsv1alpha1.ApplicationEndpoint,
) (*networkingv1.Ingress, error) {
	labels := resolvers.MergeDefaultLabels(
		appToReconcile.Labels,
		map[string]string{
			"app.kubernetes.io/instance":  appToReconcile.Name,
			"kubernetes.io/ingress.class": "k4indie-ingress",
		})

	ingressRules := resolvers.BuildIngressRules(appToReconcile.Name, endpoints)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appToReconcile.Name,
			Namespace: appToReconcile.Namespace,
			Labels:    labels,
		},
		Spec: networkingv1.IngressSpec{
			Rules: ingressRules,
		},
	}

	return ingress, nil
}

func (r *ApplicationReconciler) updateIngressSpec(
	ctx context.Context,
	appToReconcile *operatorsv1alpha1.Application,
	endpoints []operatorsv1alpha1.ApplicationEndpoint,
	ingress *networkingv1.Ingress,
) error {
	newIngress, err := r.buildIngress(ctx, appToReconcile, endpoints)
	if err != nil {
		return err
	}

	newIngress.DeepCopyInto(ingress)

	return nil
}
