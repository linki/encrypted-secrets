package provider

import (
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

func init() {
	AWSFlagSet = pflag.NewFlagSet("aws", pflag.ExitOnError)

	AWSFlagSet.String("aws-region", "eu-central-1", "The AWS region to use")
}

func HandleEncryptedSecret_AWS(cr *k8sv1alpha1.EncryptedSecret) ([]byte, error) {
	region, err := AWSFlagSet.GetString("aws-region")
	if err != nil {
		panic(err)
	}

	var client kmsiface.KMSAPI
	sess := session.Must(session.NewSession())
	client = kms.New(sess, &aws.Config{
		Region: aws.String(region),
	})

	out, err := client.Decrypt(&kms.DecryptInput{
		CiphertextBlob: cr.Spec.Ciphertext,
	})
	if err != nil {
		panic(err)
	}

	return out.Plaintext, nil
}

func HandleManagedSecret_AWS(cr *k8sv1alpha1.ManagedSecret) ([]byte, error) {
	var client secretsmanageriface.SecretsManagerAPI
	sess := session.Must(session.NewSession())
	client = secretsmanager.New(sess, &aws.Config{
		Region: aws.String("eu-central-1"),
	})

	out, err := client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: &cr.Spec.SecretName,
	})
	if err != nil {
		panic(err)
	}

	return []byte(aws.StringValue(out.SecretString)), nil
}
