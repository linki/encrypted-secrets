package provider

import (
	"context"
	"fmt"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	k8slinkidevv1beta1 "github.com/linki/encrypted-secrets/api/v1beta1"

	"cloud.google.com/go/compute/metadata"

	kms "cloud.google.com/go/kms/apiv1"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

const (
	ProviderGCP = "GCP"
)

const (
	defaultProject = "default-project"
	defaultRegion  = "default-region"
)

var _ Provider = &GCPProvider{}

type GCPProvider struct {
	kmsClient     *kms.KeyManagementClient
	secretsClient *secretmanager.Client
	projectID     string
	region        string
}

func init() {
	newProviderFuncs[ProviderGCP] = func(ctx context.Context) (Provider, error) {
		return NewGCPProvider(ctx)
	}
}

func NewGCPProvider(ctx context.Context) (*GCPProvider, error) {
	var log = logf.Log.WithName("gcp_provider")

	projectID, err := metadata.ProjectID()
	if err != nil {
		log.Info("Failed to auto-detect GCP project")
	}

	region, err := metadata.Zone()
	if err != nil {
		log.Info("Failed to auto-detect GCP region")
	}
	if i := strings.LastIndex(region, "-"); i != -1 {
		region = region[:i]
	}

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
		projectID:     projectID,
		region:        region,
	}

	return provider, nil
}

func (p *GCPProvider) HandleEncryptedSecret(ctx context.Context, cr *k8slinkidevv1beta1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, ciphertext := range cr.Spec.Data {
		req := &kmspb.DecryptRequest{
			Name:       expandKeyID(cr.Spec.KeyID, p.projectID, p.region),
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

func (p *GCPProvider) HandleManagedSecret(ctx context.Context, cr *k8slinkidevv1beta1.ManagedSecret) (map[string][]byte, error) {
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

func expandKeyID(keyID, defaultProject, defaultRegion string) string {
	keyInfo := map[string]string{
		"projects":  defaultProject,
		"locations": defaultRegion,
	}

	keyIDParts := strings.Split(keyID, "/")
	for i := 0; i < len(keyIDParts)-1; i += 2 {
		keyInfo[keyIDParts[i]] = keyIDParts[i+1]
	}

	expandedKeyID := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		keyInfo["projects"], keyInfo["locations"], keyInfo["keyRings"], keyInfo["cryptoKeys"])

	return expandedKeyID
}
