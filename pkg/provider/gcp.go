package provider

import (
	"context"
	"log"

	"github.com/spf13/pflag"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

const (
	ProviderGCP = "GCP"
)

var (
	GCPFlagSet *pflag.FlagSet
)

var _ Provider = &GCPProvider{}

type GCPProvider struct {
	kmsClient     *kms.KeyManagementClient
	secretsClient *secretmanager.Client
}

func init() {
	GCPFlagSet = pflag.NewFlagSet("gcp", pflag.ExitOnError)

	provider, err := NewGCPProvider()
	if err != nil {
		log.Fatal(err)
	}
	providers[ProviderGCP] = provider
}

func NewGCPProvider() (*GCPProvider, error) {
	ctx := context.Background()

	kmsClient, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	secretsClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	provider := &GCPProvider{
		kmsClient:     kmsClient,
		secretsClient: secretsClient,
	}

	return provider, nil
}

func (p *GCPProvider) HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) ([]byte, error) {
	req := &kmspb.DecryptRequest{
		Name:       cr.Spec.KeyID,
		Ciphertext: cr.Spec.Ciphertext,
	}
	resp, err := p.kmsClient.Decrypt(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetPlaintext(), nil
}

func (p *GCPProvider) HandleManagedSecret(ctx context.Context, cr *k8sv1alpha1.ManagedSecret) ([]byte, error) {
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: cr.Spec.SecretName,
	}
	resp, err := p.secretsClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.GetPayload().GetData(), nil
}
