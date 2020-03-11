package provider

import (
	"context"

	"github.com/spf13/pflag"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
)

const (
	ProviderIdentity = "Identity"
)

var (
	IdentityFlagSet *pflag.FlagSet
)

var _ Provider = &IdentityProvider{}

type IdentityProvider struct{}

func init() {
	IdentityFlagSet = pflag.NewFlagSet("identity", pflag.ExitOnError)

	newProviderFuncs[ProviderIdentity] = func() (Provider, error) {
		return NewIdentityProvider()
	}
}

func NewIdentityProvider() (*IdentityProvider, error) {
	return &IdentityProvider{}, nil
}

func (p *IdentityProvider) HandleEncryptedSecret(_ context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, secretValue := range cr.Spec.Data {
		data[key] = secretValue
	}

	return data, nil
}

func (p *IdentityProvider) HandleManagedSecret(_ context.Context, cr *k8sv1alpha1.ManagedSecret) (map[string][]byte, error) {
	data := map[string][]byte{}

	for key, secretName := range cr.Spec.Data {
		data[key] = []byte(secretName)
	}

	return data, nil
}
