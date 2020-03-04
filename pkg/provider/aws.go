package provider

import (
	"context"
	"log"

	"github.com/spf13/pflag"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

const (
	ProviderAWS = "AWS"
)

var (
	AWSFlagSet *pflag.FlagSet
)

var _ Provider = &AWSProvider{}

type AWSProvider struct {
	kmsClient     kmsiface.KMSAPI
	secretsClient secretsmanageriface.SecretsManagerAPI
}

func init() {
	AWSFlagSet = pflag.NewFlagSet("aws", pflag.ExitOnError)

	AWSFlagSet.String("aws-region", "eu-central-1", "The AWS region to use")

	region, err := AWSFlagSet.GetString("aws-region")
	if err != nil {
		log.Fatal(err)
	}

	provider, err := NewAWSProvider(region)
	if err != nil {
		log.Fatal(err)
	}
	providers[ProviderAWS] = provider
}

func NewAWSProvider(region string) (*AWSProvider, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	config := &aws.Config{
		Region: aws.String(region),
	}

	provider := &AWSProvider{
		kmsClient:     kms.New(sess, config),
		secretsClient: secretsmanager.New(sess, config),
	}

	return provider, nil
}

func (p *AWSProvider) HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, ciphertext := range cr.Spec.Data {
		req := &kms.DecryptInput{
			CiphertextBlob: ciphertext,
		}
		resp, err := p.kmsClient.DecryptWithContext(ctx, req)
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
		req := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		}
		resp, err := p.secretsClient.GetSecretValueWithContext(ctx, req)
		if err != nil {
			return nil, err
		}

		data[key] = []byte(aws.StringValue(resp.SecretString))
	}

	return data, nil
}
