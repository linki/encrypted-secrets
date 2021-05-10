package provider

import (
	"context"
	"fmt"
	"sync"

	k8slinkidevv1beta1 "github.com/linki/encrypted-secrets/api/v1beta1"
)

var (
	providers        = map[string]Provider{}
	newProviderFuncs = map[string]NewProviderFunc{}
	lock             = sync.Mutex{}
)

type Provider interface {
	HandleEncryptedSecret(ctx context.Context, cr *k8slinkidevv1beta1.EncryptedSecret) (map[string][]byte, error)
	HandleManagedSecret(ctx context.Context, cr *k8slinkidevv1beta1.ManagedSecret) (map[string][]byte, error)
}

type NewProviderFunc func(ctx context.Context) (Provider, error)

func ProviderFor(ctx context.Context, providerName string) (Provider, error) {
	lock.Lock()
	defer lock.Unlock()

	if provider, ok := providers[providerName]; ok {
		return provider, nil
	}

	providerFunc, ok := newProviderFuncs[providerName]
	if !ok {
		return nil, fmt.Errorf("Provider not found: %s", providerName)
	}

	provider, err := providerFunc(ctx)
	if err != nil {
		return nil, err
	}

	providers[providerName] = provider
	return provider, nil
}
