package provider

import (
	"context"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

const (
	ProviderGCP = "GCP"
)

func HandleEncryptedSecret_GCP(cr *k8sv1alpha1.EncryptedSecret) ([]byte, error) {
	ctx := context.Background()
	c, err := kms.NewKeyManagementClient(ctx)
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

	return resp.GetPlaintext(), nil
}

func HandleManagedSecret_GCP(cr *k8sv1alpha1.ManagedSecret) ([]byte, error) {
	ctx := context.Background()
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: cr.Spec.SecretName,
	}
	resp, err := c.AccessSecretVersion(ctx, req)
	if err != nil {
		panic(err)
	}

	return resp.GetPayload().GetData(), nil
}
