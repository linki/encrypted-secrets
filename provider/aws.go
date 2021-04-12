package provider

import (
	"context"
	"flag"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/api/v1alpha1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	ProviderAWS = "AWS"
)

var (
	region string
)

var _ Provider = &AWSProvider{}

type AWSProvider struct {
	kmsClient     *kms.Client
	secretsClient *secretsmanager.Client
}

func init() {
	flag.StringVar(&region, "aws-region", "eu-central-1", "The AWS region to use")

	newProviderFuncs[ProviderAWS] = func(ctx context.Context) (Provider, error) {
		return NewAWSProvider(ctx)
	}
}

func NewAWSProvider(ctx context.Context) (*AWSProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	provider := &AWSProvider{
		kmsClient:     kms.NewFromConfig(cfg),
		secretsClient: secretsmanager.NewFromConfig(cfg),
	}

	return provider, nil
}

func (p *AWSProvider) HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, ciphertext := range cr.Spec.Data {
		resp, err := p.kmsClient.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: ciphertext,
		})
		if err != nil {
			return nil, err
		}

		data[key] = resp.Plaintext
	}

	return data, nil
}

func (p *AWSProvider) HandleManagedSecret(ctx context.Context, cr *k8sv1alpha1.ManagedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, secretName := range cr.Spec.Data {
		resp, err := p.secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		})
		if err != nil {
			return nil, err
		}

		data[key] = []byte(*resp.SecretString)
	}

	return data, nil
}
