package provider

import (
	"context"
	"fmt"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
)

var (
	providers = map[string]Provider{}
)

type Provider interface {
	HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) ([]byte, error)
	HandleManagedSecret(ctx context.Context, cr *k8sv1alpha1.ManagedSecret) ([]byte, error)
}

func ProviderFor(provider string) (Provider, error) {
	if provider, ok := providers[provider]; !ok {
		return nil, fmt.Errorf("Provider not found: %s", provider)
	}

	return providers[provider], nil
}
