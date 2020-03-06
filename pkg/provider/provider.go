package provider

import (
	"context"
	"fmt"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
)

var (
	providers        = map[string]Provider{}
	newProviderFuncs = map[string]func() (Provider, error){}
)

type Provider interface {
	HandleEncryptedSecret(ctx context.Context, cr *k8sv1alpha1.EncryptedSecret) (map[string][]byte, error)
	HandleManagedSecret(ctx context.Context, cr *k8sv1alpha1.ManagedSecret) (map[string][]byte, error)
}

func ProviderFor(providerName string) (Provider, error) {
	if provider, ok := providers[providerName]; ok {
		return provider, nil
	}

	providerFunc, ok := newProviderFuncs[providerName]
	if !ok {
		return nil, fmt.Errorf("Provider not found: %s", providerName)
	}

	provider, err := providerFunc()
	if err != nil {
		return nil, err
	}

	providers[providerName] = provider
	return provider, nil
}
