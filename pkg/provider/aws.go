package provider

import (
	"context"

	"github.com/spf13/pflag"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/kmsiface"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/secretsmanageriface"
)

const (
	ProviderAWS = "AWS"
)

var (
	AWSFlagSet *pflag.FlagSet
)

var _ Provider = &AWSProvider{}

type AWSProvider struct {
	kmsClient     kmsiface.ClientAPI
	secretsClient secretsmanageriface.ClientAPI
}

func init() {
	AWSFlagSet = pflag.NewFlagSet("aws", pflag.ExitOnError)

	AWSFlagSet.String("aws-region", "eu-central-1", "The AWS region to use")

	newProviderFuncs[ProviderAWS] = func() (Provider, error) {
		return NewAWSProvider()
	}
}

func NewAWSProvider() (*AWSProvider, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}

	region, err := AWSFlagSet.GetString("aws-region")
	if err != nil {
		return nil, err
	}
	cfg.Region = region

	provider := &AWSProvider{
		kmsClient:     kms.New(cfg),
		secretsClient: secretsmanager.New(cfg),
	}

	return provider, nil
}

func (p *AWSProvider) HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, ciphertext := range cr.Spec.Data {
		req := p.kmsClient.DecryptRequest(&kms.DecryptInput{
			CiphertextBlob: ciphertext,
		})

		resp, err := req.Send(ctx)
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
		req := p.secretsClient.GetSecretValueRequest(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		})

		resp, err := req.Send(ctx)
		if err != nil {
			return nil, err
		}

		data[key] = []byte(aws.StringValue(resp.SecretString))
	}

	return data, nil
}
