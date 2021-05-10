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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8slinkidevv1beta1 "github.com/linki/encrypted-secrets/api/v1beta1"
	"github.com/linki/encrypted-secrets/provider"
)

// EncryptedSecretReconciler reconciles a EncryptedSecret object
type EncryptedSecretReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=k8s,resources=encryptedsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s,resources=encryptedsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s,resources=encryptedsecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EncryptedSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *EncryptedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("encryptedsecret", req.NamespacedName)

	// your logic here

	// Fetch the EncryptedSecret instance
	instance := &k8slinkidevv1beta1.EncryptedSecret{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	fmt.Println(instance.Name)

	// Define a new Secret object
	secret, err := newSecretForCR(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Set EncryptedSecret instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if this Secret already exists
	found := &v1.Secret{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		r.Log.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.Client.Create(ctx, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(instance, v1.EventTypeNormal, "SuccessfulCreate", "Created secret: %s", secret.Name)

		// Secret created successfully - don't requeue
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(found.Data, secret.Data) {
		r.Log.Info("Updating existing Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.Client.Update(ctx, secret)
		if err != nil {
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(instance, v1.EventTypeNormal, "SuccessfulUpdate", "Updated secret: %s", secret.Name)

		// Secret updated successfully - don't requeue
		return ctrl.Result{}, nil
	}

	// Secret already exists and is unchanged - don't requeue
	r.Log.Info("Skip reconcile: Secret already exists and is unchanged", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EncryptedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8slinkidevv1beta1.EncryptedSecret{}).
		Owns(&v1.Secret{}).
		Complete(r)
}

// newSecretForCR returns a plain old secret with the same name/namespace as the cr containing the decrypted secret value
func newSecretForCR(ctx context.Context, cr *k8slinkidevv1beta1.EncryptedSecret) (*v1.Secret, error) {
	provider, err := provider.ProviderFor(ctx, cr.Spec.Provider)
	if err != nil {
		return nil, err
	}

	data, err := provider.HandleEncryptedSecret(ctx, cr)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: data,
	}, nil
}
