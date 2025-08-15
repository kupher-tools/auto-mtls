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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// AutomtlsReconciler reconciles a Automtls object
type AutomtlsReconciler struct {
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

// SetupWithManager sets up the controller with the Manager.
func (r *AutomtlsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetAnnotations()["auto-mtls.kupher.io/enabled"] == "true"
		})).
		Complete(r)
}

// 1. Get service with specific annoation
// 2. Create a Cert , which intern create secret
// 3. Once secret created, patch deployment to mount tls secret

func (r *AutomtlsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling auto-mtls", "name", req.Name, "namespace", req.Namespace)
	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		log.Error(err, "unable to fetch Service")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := r.enablemTLS(ctx, svc, log)
	if err != nil {
		log.Error(err, "Failed to enable mTLS for service", "service", svc.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AutomtlsReconciler) enablemTLS(ctx context.Context, svc *corev1.Service, log logr.Logger) error {
	// Implementation for enabling server TLS

	// Create Server cert and corresponding TLS secret
	if err := r.createServerCert(ctx, svc, log); err != nil {
		log.Error(err, "Failed to create certificate for service", "service", svc.Name)
		return err
	}

	// Create CA cert TLS secret
	if err := r.createCACertSecret(ctx, svc, log); err != nil {
		log.Error(err, "Failed to create CA cert secret for service", "service", svc.Name)
		return err
	}

	//mount Ca Cert and Server keys
	if err := r.mountMTLSCerts(ctx, svc, log); err != nil {
		log.Error(err, "Failed to create CA cert secret for service", "service", svc.Name)
		return err
	}
	log.Info("Successfully mounted mTLS certificates for service", "service", svc.Name)
	return nil

}

func (r *AutomtlsReconciler) mountMTLSCerts(ctx context.Context, svc *corev1.Service, log logr.Logger) error {
	// Implementation for mounting mTLS certificates into the deployment
	deploy, err := r.findDeploymentForSvc(ctx, svc)
	if err != nil {
		log.Error(err, "Failed to find deployment for service", "service", svc.Name)
		return err
	}
	if deploy == nil {
		log.Info("No deployment found for service", "service", svc.Name)
		return nil // Nothing to do if no deployment found
	} else {
		err = mountSecrets(ctx, r.Client, deploy, svc.Name)

		if err != nil {
			log.Error(err, "Failed to patch deployment with server certificate", "deployment", deploy.Name, "service", svc.Name)
			return err
		}

		if err != nil {
			log.Error(err, "Failed to mount server certificate", "service", svc.Name)
			return err
		}

		log.Info("Successfully mounted server certificate to deployment", "deployment", deploy.Name, "service", svc.Name)
		return nil
	}

}

func (r *AutomtlsReconciler) createCACertSecret(ctx context.Context, svc *corev1.Service, log logr.Logger) error {
	caCertSecret := &corev1.Secret{}

	err := r.Get(ctx, types.NamespacedName{
		Name:      "auto-mtls-ca-cert",
		Namespace: svc.Namespace,
	}, caCertSecret)

	if err == nil {
		// Secret already exists — skip
		log.Info("Secret already exist, so skipping")
		return nil
	} else {
		//Create secret for CA cert in namespace
		src := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      "auto-mtls-cluster-ca-cert-secret",
			Namespace: "cert-manager",
		}, src); err != nil {
			log.Error(err, "failed to get source CA secret")
			return err
		}

		caData, ok := src.Data["ca.crt"]
		if !ok {
			log.Error(err, "Source secret missing ca.crt")
			return fmt.Errorf("source secret missing ca.crt")

		}
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "auto-mtls-ca-cert",
				Namespace: svc.Namespace,
			},
			Data: map[string][]byte{
				"ca.crt": caData,
			},
			Type: corev1.SecretTypeOpaque,
		}

		if err := r.Create(ctx, newSecret); err != nil {
			return fmt.Errorf("failed to create secret in %s: %w", svc.Namespace, err)
		}

		fmt.Println("Created CA secret in", svc.Namespace)

	}
	return nil
}

func (r *AutomtlsReconciler) findDeploymentForSvc(ctx context.Context, svc *corev1.Service) (*appsv1.Deployment, error) {
	// List all Services in the Deployment's namespace
	var deployList appsv1.DeploymentList
	if err := r.List(ctx, &deployList, client.InNamespace(svc.Namespace)); err != nil {
		return nil, err
	}

	// Get labels from Deployment's Pod template
	svcLabels := svc.Spec.Selector

	// Find matching Service
	for _, deploy := range deployList.Items {
		if selectorMatches(svcLabels, deploy.Spec.Template.Labels) {
			return &deploy, nil
		}
	}

	return nil, nil
}

// helper: check if all selector key/values exist in labels
func selectorMatches(labels, selector map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func (r *AutomtlsReconciler) createServerCert(ctx context.Context, service *corev1.Service, log logr.Logger) error {
	namespace := service.Namespace
	svc := service.Name
	caIssuer := "auto-mtls-cluster-ca-issuer"
	certName := svc + "-cert"
	secretName := certName + "-tls"

	existingCert := &certmanagerv1.Certificate{}

	err := r.Get(ctx, types.NamespacedName{
		Name:      certName,
		Namespace: namespace,
	}, existingCert)

	if err == nil {
		// Certificate already exists — nothing to do
		log.Info("Certificate already exists", "name", certName, "namespace", namespace)
		return nil
	}

	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certName,
			Namespace: namespace,
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName:  secretName,
			Duration:    &metav1.Duration{Duration: 8760 * time.Hour}, // 1 year
			RenewBefore: &metav1.Duration{Duration: 720 * time.Hour},  // 30 days
			CommonName:  svc + "." + namespace + ".svc.cluster.local",
			DNSNames: []string{
				svc,
				svc + "." + namespace,
				svc + "." + namespace + ".svc",
				svc + "." + namespace + ".svc.cluster.local",
			},
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: caIssuer,
				Kind: "ClusterIssuer",
			},
			SecretTemplate: &certmanagerv1.CertificateSecretTemplate{
				Annotations: map[string]string{
					"auto-mtls.kupher.io/generated-for": namespace + "/" + svc,
				},
			},
		},
	}
	err = r.Create(ctx, cert)
	if err != nil {
		log.Error(err, "Failed to create certificate", "name", certName, "namespace", namespace)
		return err
	}
	log.Info("Created certificate", "name", certName, "namespace", namespace)
	return nil
}

// patchDeployment adds a volume and volumeMount if missing
func mountSecrets(ctx context.Context, c client.Client, deploy *appsv1.Deployment, svcName string) error {
	serverCertvolumeName := svcName + "-cert-tls"
	patched := deploy.DeepCopy()

	// Check if foundServerCertVol exists, if not append
	foundServerCertVol := false
	for _, v := range patched.Spec.Template.Spec.Volumes {
		if v.Name == serverCertvolumeName {
			fmt.Println("Skipping auto-mtls-cert volume to deployment", "deployment", deploy.Name)
			foundServerCertVol = true
		}
	}
	if !foundServerCertVol {
		patched.Spec.Template.Spec.Volumes = append(patched.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: serverCertvolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: serverCertvolumeName, // Secret name spacific to service
					},
				},
			},
		)
	}

	// Add volumeMount to each container if missing
	for i, container := range patched.Spec.Template.Spec.Containers {
		foundMount := false
		for _, vm := range container.VolumeMounts {
			if vm.Name == serverCertvolumeName {
				foundMount = true
				break
			}
		}
		if !foundMount {
			patched.Spec.Template.Spec.Containers[i].VolumeMounts = append(container.VolumeMounts,
				corev1.VolumeMount{
					Name:      serverCertvolumeName,
					MountPath: "/etc/tls",
					ReadOnly:  true,
				},
			)
		}
	}

	caCertvolumeName := "auto-mtls-ca-cert"
	// Check if foundServerCertVol exists, if not append
	caCertVol := false
	for _, v := range patched.Spec.Template.Spec.Volumes {
		if v.Name == caCertvolumeName {
			fmt.Println("Skipping auto-mtls-cert volume to deployment", "deployment", deploy.Name)
			caCertVol = true
		}
	}
	if !caCertVol {
		patched.Spec.Template.Spec.Volumes = append(patched.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: caCertvolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: caCertvolumeName, // Secret name spacific to service
					},
				},
			},
		)
	}

	// Add volumeMount to each container if missing
	for i, container := range patched.Spec.Template.Spec.Containers {
		foundMount := false
		for _, vm := range container.VolumeMounts {
			if vm.Name == caCertvolumeName {
				foundMount = true
				break
			}
		}
		if !foundMount {
			patched.Spec.Template.Spec.Containers[i].VolumeMounts = append(container.VolumeMounts,
				corev1.VolumeMount{
					Name:      caCertvolumeName,
					MountPath: "/etc/ca",
					ReadOnly:  true,
				},
			)
		}
	}

	// Patch the deployment

	return c.Patch(ctx, patched, client.MergeFrom(deploy))
}
