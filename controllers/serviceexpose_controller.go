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
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	serviceexposev1alpha1 "github.com/takumakume/service-expose-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const managedByServiceExposeLabelValue = "serviceexpose"

//const serviceExposeFinalizer = "service-exposer.github.io/finalizer"

// ServiceExposeReconciler reconciles a ServiceExpose object
type ServiceExposeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=service-expose.takumakume.github.io,resources=serviceexposes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=service-expose.takumakume.github.io,resources=serviceexposes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=service-expose.takumakume.github.io,resources=serviceexposes/finalizers,verbs=update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceExpose object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ServiceExposeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	se := &serviceexposev1alpha1.ServiceExpose{}
	err := r.Get(ctx, req.NamespacedName, se)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("ServiceExpose resource not found. Ignoring since object must be deleted.")
			return r.stop()
		}
		return r.stopWithError(err)
	}

	currentIngress, err := r.getIngress(ctx, se)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.createIngress(ctx, se); err != nil {
				r.Log.Error(err, fmt.Sprintf("create new ingress error. ServiceExpose:%+v", se))
				se.Status.Ready = corev1.ConditionFalse
				if statusUpdateError := r.Client.Status().Update(ctx, se); statusUpdateError != nil {
					r.Log.Error(statusUpdateError, "status update error.")
				}
				return r.stopWithError(err)
			}
			se.Status.Ready = corev1.ConditionTrue
			if statusUpdateError := r.Client.Status().Update(ctx, se); statusUpdateError != nil {
				r.Log.Error(statusUpdateError, "status update error.")
			}
			return r.requeue()
		}
		return r.stopWithError(err)
	}

	if needsUpdateIngress(currentIngress, se) {
		if !isManagedByServiceExpose(currentIngress) {
			se.Status.Ready = corev1.ConditionFalse
			if statusUpdateError := r.Client.Status().Update(ctx, se); statusUpdateError != nil {
				r.Log.Error(statusUpdateError, "status update error.")
			}
			return r.stopWithError(fmt.Errorf("This ingress is out of control. Ingress/%s", currentIngress.Name))
		}
		if err := r.updateIngress(ctx, se); err != nil {
			r.Log.Error(err, fmt.Sprintf("update ingress error. ServiceExpose:%+v", se))
			se.Status.Ready = corev1.ConditionFalse
			if statusUpdateError := r.Client.Status().Update(ctx, se); statusUpdateError != nil {
				r.Log.Error(statusUpdateError, "status update error.")
			}
			return r.stopWithError(err)
		}
		se.Status.Ready = corev1.ConditionTrue
		if statusUpdateError := r.Client.Status().Update(ctx, se); statusUpdateError != nil {
			r.Log.Error(statusUpdateError, "status update error.")
		}
	}

	return r.requeue()
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceExposeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serviceexposev1alpha1.ServiceExpose{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}

func (r *ServiceExposeReconciler) requeue() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

func (r *ServiceExposeReconciler) stopWithError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func (r *ServiceExposeReconciler) stop() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *ServiceExposeReconciler) getIngress(ctx context.Context, se *serviceexposev1alpha1.ServiceExpose) (*networkingv1.Ingress, error) {
	ingress := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: se.Status.IngressName, Namespace: se.Namespace}, ingress); err != nil {
		return nil, err
	}

	return ingress, nil
}

func (r *ServiceExposeReconciler) createIngress(ctx context.Context, se *serviceexposev1alpha1.ServiceExpose) error {
	ingress := generateIngress(se)

	if err := ctrl.SetControllerReference(se, ingress, r.Scheme); err != nil {
		return err
	}

	r.Log.Info(fmt.Sprintf("creating new ingress. ingress:%+v", ingress))
	if err := r.Create(ctx, ingress); err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("created new ingress: %+v", ingress))

	se.Status.IngressHost = ingress.Spec.Rules[0].Host
	se.Status.IngressName = ingress.ObjectMeta.Name
	if err := r.Client.Status().Update(ctx, se); err != nil {
		return err
	}

	return nil
}

func (r *ServiceExposeReconciler) updateIngress(ctx context.Context, se *serviceexposev1alpha1.ServiceExpose) error {
	ingress := generateIngress(se)

	if err := ctrl.SetControllerReference(se, ingress, r.Scheme); err != nil {
		return err
	}

	r.Log.Info(fmt.Sprintf("updating ingress. ingress:%+v", ingress))
	if err := r.Update(ctx, ingress); err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("update ingress: %+v", ingress))

	se.Status.IngressHost = ingress.Spec.Rules[0].Host
	se.Status.IngressName = ingress.ObjectMeta.Name
	if err := r.Client.Status().Update(ctx, se); err != nil {
		return err
	}

	return nil
}

func needsUpdateIngress(currentIngress *networkingv1.Ingress, se *serviceexposev1alpha1.ServiceExpose) bool {
	switch {
	case len(currentIngress.Spec.Rules) != 1:
		return true
	case currentIngress.Spec.Rules[0].Host != generateIngresHost(se):
		return true
	case len(currentIngress.Spec.Rules[0].HTTP.Paths) != 1:
		return true
	case !reflect.DeepEqual(currentIngress.Spec.Rules[0].HTTP.Paths[0].Backend, se.Spec.Backend):
		return true
	case currentIngress.Spec.Rules[0].HTTP.Paths[0].Path != se.Spec.Path:
		return true
	case *currentIngress.Spec.Rules[0].HTTP.Paths[0].PathType != se.Spec.PathType:
		return true
	case len(currentIngress.Spec.TLS) > 0 != se.Spec.TLSEnabled:
		return true
	case !reflect.DeepEqual(currentIngress.Annotations, se.Spec.Annotations):
		return true
	}

	if se.Spec.TLSEnabled {
		switch {
		case len(currentIngress.Spec.TLS) != 1:
			return true
		case len(currentIngress.Spec.TLS[0].Hosts) != 1:
			return true
		case currentIngress.Spec.TLS[0].Hosts[0] != generateIngresHost(se):
			return true
		case currentIngress.Spec.TLS[0].SecretName != se.Spec.TLSSecretName:
			return true
		}
	}

	return false
}

func isManagedByServiceExpose(ingress *networkingv1.Ingress) bool {
	if ingress.Labels["app.kubernetes.io/managed-by"] != managedByServiceExposeLabelValue {
		return false
	}
	return true
}

func generateIngresName(se *serviceexposev1alpha1.ServiceExpose) string {
	// TODO: support backend type
	return se.Spec.Backend.Service.Name
}

func generateIngresHost(se *serviceexposev1alpha1.ServiceExpose) string {
	return fmt.Sprintf("%s.%s.%s", generateIngresName(se), se.Namespace, se.Spec.Domain)
}

func generateIngress(se *serviceexposev1alpha1.ServiceExpose) *networkingv1.Ingress {
	ingressName := generateIngresName(se)
	ingressHost := generateIngresHost(se)

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressName,
			Namespace:   se.Namespace,
			Annotations: se.Spec.Annotations,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": managedByServiceExposeLabelValue,
			},
		},
		Spec: networkingv1.IngressSpec{
			// TODO: support ingressClassName
			Rules: []networkingv1.IngressRule{
				{
					Host: ingressHost,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     se.Spec.Path,
									PathType: &se.Spec.PathType,
									Backend:  se.Spec.Backend,
								},
							},
						},
					},
				},
			},
		},
	}
	if se.Spec.TLSEnabled {
		ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{ingressHost},
				SecretName: se.Spec.TLSSecretName,
			},
		}
	}

	return ingress
}
