package encryptedsecret

import (
	"bytes"
	"context"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	googlekms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

var log = logf.Log.WithName("controller_encryptedsecret")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new EncryptedSecret Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileEncryptedSecret{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("encryptedsecret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource EncryptedSecret
	err = c.Watch(&source.Kind{Type: &k8sv1alpha1.EncryptedSecret{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Secrets and requeue the owner EncryptedSecret
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &k8sv1alpha1.EncryptedSecret{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileEncryptedSecret implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEncryptedSecret{}

// ReconcileEncryptedSecret reconciles a EncryptedSecret object
type ReconcileEncryptedSecret struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a EncryptedSecret object and makes changes based on the state read
// and what is in the EncryptedSecret.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Secret as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileEncryptedSecret) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling EncryptedSecret")

	// Fetch the EncryptedSecret instance
	instance := &k8sv1alpha1.EncryptedSecret{}
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
	secret := newSecretForCR(instance)

	// Set EncryptedSecret instance as the owner and controller
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

		// Secret created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if bytes.Compare(found.Data["content"], secret.Data["content"]) != 0 {
		reqLogger.Info("Updating existing Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
		err = r.client.Update(context.TODO(), secret)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Secret updated successfully - don't requeue
		return reconcile.Result{}, nil
	}

	// Secret already exists and is unchanged - don't requeue
	reqLogger.Info("Skip reconcile: Secret already exists and is unchanged", "Secret.Namespace", found.Namespace, "Secret.Name", found.Name)
	return reconcile.Result{}, nil
}

// newSecretForCR returns a plain old secret with the same name/namespace as the cr containing the decrypted secret value
func newSecretForCR(cr *k8sv1alpha1.EncryptedSecret) *corev1.Secret {
	var result []byte

	switch cr.Spec.Provider {
	case "AWS":
		var client kmsiface.KMSAPI
		sess := session.Must(session.NewSession())
		client = kms.New(sess, &aws.Config{
			Region: aws.String("eu-central-1"),
		})

		out, err := client.Decrypt(&kms.DecryptInput{
			CiphertextBlob: cr.Spec.Ciphertext,
		})
		if err != nil {
			panic(err)
		}

		result = out.Plaintext
	case "GCP":
		ctx := context.Background()
		c, err := googlekms.NewKeyManagementClient(ctx)
		if err != nil {
			panic(err)
		}
		defer c.Close()

		req := &kmspb.DecryptRequest{
			Name:       cr.Spec.KeyID,
			Ciphertext: cr.Spec.Ciphertext,
		}
		resp, err := c.Decrypt(ctx, req)
		if err != nil {
			panic(err)
		}

		result = resp.GetPlaintext()
	default:
		panic("provider doesn't exist")
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Data: map[string][]byte{
			"content": result,
		},
	}
}
