/*
Copyright 2025.

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
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// AutomtlsReconciler reconciles a Automtls object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=automtls.kupher.io,resources=automtls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=automtls.kupher.io,resources=automtls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=automtls.kupher.io,resources=automtls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Automtls object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling Cert Mount Controller", "name", req.Name, "namespace", req.Namespace)

	caCertSecret := &corev1.Secret{}

	err := r.Get(ctx, types.NamespacedName{
		Name:      "ca-cert",
		Namespace: req.Namespace,
	}, caCertSecret)

	if err == nil {
		// Secret already exists — skip
		fmt.Println("CA secret already exists in", req.Namespace)

	} else {
		//Create secret for CA cert in namespace

		src := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      "auto-mtls-cluster-ca-cert-secret",
			Namespace: "cert-manager",
		}, src); err != nil {
			log.Error(err, "failed to get source CA secret")
			return ctrl.Result{}, err
		}

		caData, ok := src.Data["ca.crt"]
		if !ok {
			log.Error(err, "Source secret missing ca.crt")
			return ctrl.Result{}, fmt.Errorf("source secret missing ca.crt")

		}
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "auto-mtls-ca-cert",
				Namespace: req.Namespace,
			},
			Data: map[string][]byte{
				"ca.crt": caData,
			},
			Type: corev1.SecretTypeOpaque,
		}

		if err := r.Create(ctx, newSecret); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to create secret in %s: %w", req.Namespace, err)
		}

		fmt.Println("Created CA secret in", req.Namespace)

	}

	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		log.Error(err, "Unable to fetch deployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Patch deployment to add auto-mtls-cert volume
	if err := patchDeployment(ctx, r.Client, deployment); err != nil {
		log.Error(err, "Failed to patch deployment with auto-mtls-ca-cert volume", "deployment", deployment.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// patchDeployment adds a volume and volumeMount if missing
func patchDeployment(ctx context.Context, c client.Client, deploy *appsv1.Deployment) error {
	volumeName := "auto-mtls-ca-cert"
	secretName := "auto-mtls-ca-cert"
	patched := deploy.DeepCopy()

	// Check if volume exists, if not append

	foundVol := false
	for _, v := range patched.Spec.Template.Spec.Volumes {
		if v.Name == volumeName {
			fmt.Println("Skipping auto-mtls-ca-cert volume to deployment", "deployment", deploy.Name)
			foundVol = true
			return nil
		}
	}
	if !foundVol {
		patched.Spec.Template.Spec.Volumes = append(patched.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: volumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName, // Secret name spacific to service
					},
				},
			},
		)
	}

	// Add volumeMount to each container if missing
	for i, container := range patched.Spec.Template.Spec.Containers {
		foundMount := false
		for _, vm := range container.VolumeMounts {
			if vm.Name == volumeName {
				foundMount = true
				break
			}
		}
		if !foundMount {
			patched.Spec.Template.Spec.Containers[i].VolumeMounts = append(container.VolumeMounts,
				corev1.VolumeMount{
					Name:      volumeName,
					MountPath: "/etc/ca-cert",
					ReadOnly:  true,
				},
			)
		}
	}

	// Patch the deployment
	return c.Patch(ctx, patched, client.MergeFrom(deploy))
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetAnnotations()["auto-mtls.kupher.io/ca-public-cert"] == "true"
		})).
		Complete(r)
}
