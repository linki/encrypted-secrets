package provider

import (
	"context"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/api/v1alpha1"
)

const (
	ProviderIdentity = "Identity"
)

var _ Provider = &IdentityProvider{}

type IdentityProvider struct{}

func init() {
	newProviderFuncs[ProviderIdentity] = func(ctx context.Context) (Provider, error) {
		return NewIdentityProvider(ctx)
	}
}

func NewIdentityProvider(_ context.Context) (*IdentityProvider, error) {
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
