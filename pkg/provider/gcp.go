package provider

import (
	"context"

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

	newProviderFuncs[ProviderGCP] = func() (Provider, error) {
		return NewGCPProvider()
	}
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

func (p *GCPProvider) HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, ciphertext := range cr.Spec.Data {
		req := &kmspb.DecryptRequest{
			Name:       cr.Spec.KeyID,
			Ciphertext: ciphertext,
		}
		resp, err := p.kmsClient.Decrypt(ctx, req)
		if err != nil {
			return nil, err
		}

		data[key] = resp.GetPlaintext()
	}

	return data, nil
}

func (p *GCPProvider) HandleManagedSecret(ctx context.Context, cr *k8sv1alpha1.ManagedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, secretName := range cr.Spec.Data {
		req := &secretmanagerpb.AccessSecretVersionRequest{
			Name: secretName,
		}
		resp, err := p.secretsClient.AccessSecretVersion(ctx, req)
		if err != nil {
			return nil, err
		}

		data[key] = resp.GetPayload().GetData()
	}

	return data, nil
}
