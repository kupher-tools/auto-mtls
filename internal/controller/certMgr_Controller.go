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
	"time"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	certmanagermetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AutomtlsReconciler reconciles a Automtls object
type CertMgrReconciler struct {
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
func (r *CertMgrReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("cert-mgr")

	log.Info("Starting certificate manager infrastructure reconciliation")

	if err := createSelfSignedIssuer(ctx, r.Client); err != nil {
		return ctrl.Result{}, err
	}

	if err := createCACert(ctx, r.Client); err != nil {
		return ctrl.Result{}, err
	}

	if err := createClusterCAIssuer(ctx, r.Client); err != nil {
		return ctrl.Result{}, err
	}

	return reconcile.Result{RequeueAfter: 10 * time.Second}, nil

}

func createSelfSignedIssuer(ctx context.Context, c client.Client) error {
	log := logf.FromContext(ctx)
	selfSignedIssuer := "auto-mtls-cluster-selfsigned-issuer"

	// Check if it already exists
	existing := &certmanagerv1.ClusterIssuer{}
	err := c.Get(ctx, client.ObjectKey{Name: selfSignedIssuer}, existing)
	if err == nil {
		log.Info("Self-signed issuer already exists, skipping creation", "issuer", selfSignedIssuer)
		return nil
	}

	// Define ClusterIssuer object
	clusterIssuer := &certmanagerv1.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "ClusterIssuer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: selfSignedIssuer,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				SelfSigned: &certmanagerv1.SelfSignedIssuer{},
			},
		},
	}

	// Create if not exists
	if err := c.Create(ctx, clusterIssuer); err != nil {
		log.Error(err, "Failed to create self-signed issuer", "issuer", clusterIssuer.Name)
		return err
	}

	log.Info("Self-signed issuer created successfully", "issuer", clusterIssuer.Name)
	return nil

}

func createCACert(ctx context.Context, c client.Client) error {
	log := logf.FromContext(ctx)
	caCertName := "auto-mtls-cluster-ca-cert"
	caCertNamespace := "cert-manager"
	caCertSecret := "auto-mtls-cluster-ca-cert-secret"
	caCertCommonName := "auto-mtls-cluster-ca"

	existing := &certmanagerv1.Certificate{}
	// Check if it already exists
	err := c.Get(ctx, client.ObjectKey{Name: caCertName, Namespace: caCertNamespace}, existing)
	if err == nil {
		log.Info("CA certificate already exists, skipping creation", "certificate", caCertName, "namespace", caCertNamespace)
		return nil
	}

	// If does not exist, create one
	caCert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caCertName,
			Namespace: caCertNamespace,
		},
		Spec: certmanagerv1.CertificateSpec{
			IsCA:       true,
			SecretName: caCertSecret,
			CommonName: caCertCommonName,
			IssuerRef: certmanagermetav1.ObjectReference{
				Name: "auto-mtls-cluster-selfsigned-issuer",
				Kind: "ClusterIssuer",
			},
		},
	}

	if err := c.Create(ctx, caCert); err != nil {
		log.Error(err, "Failed to create CA certificate", "certificate", caCert.Name, "namespace", caCert.Namespace)
		return err
	}
	log.Info("CA certificate created successfully", "certificate", caCert.Name, "namespace", caCert.Namespace)
	return nil
}

func createClusterCAIssuer(ctx context.Context, c client.Client) error {
	log := logf.FromContext(ctx)
	caIssuer := "auto-mtls-cluster-ca-issuer"

	// Check if it already exists
	existing := &certmanagerv1.ClusterIssuer{}
	err := c.Get(ctx, client.ObjectKey{Name: caIssuer}, existing)
	if err == nil {
		log.Info("CA cluster issuer already exists, skipping creation", "issuer", caIssuer)
		return nil
	}

	// Define ClusterIssuer object
	clusterIssuer := &certmanagerv1.ClusterIssuer{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cert-manager.io/v1",
			Kind:       "ClusterIssuer",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: caIssuer,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: certmanagerv1.IssuerConfig{
				CA: &certmanagerv1.CAIssuer{
					SecretName: "auto-mtls-cluster-ca-cert-secret",
				},
			},
		},
	}

	// Create if not exists
	if err := c.Create(ctx, clusterIssuer); err != nil {
		log.Error(err, "Failed to create CA cluster issuer", "issuer", clusterIssuer.Name)
		return err
	}

	log.Info("CA cluster issuer created successfully", "issuer", clusterIssuer.Name)
	return nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *CertMgrReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		ticker := time.NewTicker(10 * time.Second) // your interval
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log := logf.FromContext(ctx)
				log.V(1).Info("Running periodic certificate manager reconciliation")
				// Call your existing Reconcile logic
				if _, err := r.Reconcile(ctx, ctrl.Request{}); err != nil {
					log.Error(err, "Error in periodic certificate manager reconciliation")
				}
			case <-ctx.Done():
				return nil
			}
		}
	}))

	return nil
}
