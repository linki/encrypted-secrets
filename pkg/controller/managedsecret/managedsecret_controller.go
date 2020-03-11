package managedsecret

import (
	"context"
	"reflect"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
	"github.com/linki/encrypted-secrets/pkg/provider"
)

const (
	apiVersion = "k8s.linki.space/v1alpha1"
	kind       = "ManagedSecret"
)

var log = logf.Log.WithName("controller_managedsecret")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ManagedSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedSecret{
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		eventRecorder: mgr.GetEventRecorderFor("managedsecret-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Check if primary resource ManagedSecret exists
	dc := discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig())
	exists, err := k8sutil.ResourceExists(dc, apiVersion, kind)
	if err != nil {
		return err
	}
	if !exists {
		log.WithValues("APIVersion", apiVersion, "Kind", kind).Info("CustomResourceDefinition not found")
		return nil
	}

	// Create a new controller
	c, err := controller.New("managedsecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ManagedSecret
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.ManagedSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Secrets and requeue the owner ManagedSecret
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &k8sv1alpha1.ManagedSecret{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManagedSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManagedSecret{}

// ReconcileManagedSecret reconciles a ManagedSecret object
type ReconcileManagedSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	scheme        *runtime.Scheme
	eventRecorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a ManagedSecret object and makes changes based on the state read
// and what is in the ManagedSecret.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Secret as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileManagedSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManagedSecret")

	// Fetch the ManagedSecret instance
	instance := &k8sv1alpha1.ManagedSecret{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Secret object
	secret, err := newSecretForCR(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set ManagedSecret instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists
	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.eventRecorder.Eventf(instance, v1.EventTypeNormal, "SuccessfulCreate", "Created secret: %s", secret.Name)

		// Secret created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if !reflect.DeepEqual(found.Data, secret.Data) {
		reqLogger.Info("Updating existing Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Update(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		r.eventRecorder.Eventf(instance, v1.EventTypeNormal, "SuccessfulUpdate", "Updated secret: %s", secret.Name)

		// Secret updated successfully - don't requeue
		return reconcile.Result{}, nil
	}

	// Secret already exists and is unchanged - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists and is unchanged", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

// newSecretForCR returns a plain old secret with the same name/namespace as the cr containing the decrypted secret value
func newSecretForCR(cr *k8sv1alpha1.ManagedSecret) (*corev1.Secret, error) {
	provider, err := provider.ProviderFor(cr.Spec.Provider)
	if err != nil {
		return nil, err
	}

	data, err := provider.HandleManagedSecret(context.TODO(), cr)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: data,
	}, nil
}
