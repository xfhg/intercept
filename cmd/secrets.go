package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// SecretProvider defines the interface for secret management providers
type SecretProvider interface {
	GetSecret(ctx context.Context, path string) (string, error)
	ListSecrets(ctx context.Context, path string) ([]string, error)
}

// SecretManager handles secret operations across different providers
type SecretManager struct {
	provider SecretProvider
}

// SecretProviderType represents supported secret manager types
type SecretProviderType string

const (
	AWSSecretsManager    SecretProviderType = "aws"
	AzureKeyVault        SecretProviderType = "azure"
	HashicorpVault       SecretProviderType = "vault"
	EnvironmentVariables SecretProviderType = "env"
)

// NewSecretManager creates a new secret manager instance
func NewSecretManager(providerType SecretProviderType, config map[string]string) (*SecretManager, error) {
	var provider SecretProvider
	var err error

	switch providerType {
	case AWSSecretsManager:
		provider, err = newAWSSecretsProvider(config)
	case AzureKeyVault:
		// provider, err = newAzureKeyVaultProvider(config)
	case HashicorpVault:
		// provider, err = newHashicorpVaultProvider(config)
	case EnvironmentVariables:
		provider = newEnvProvider()
	default:
		return nil, fmt.Errorf("unsupported secret provider type: %s", providerType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize secret provider: %w", err)
	}

	return &SecretManager{provider: provider}, nil
}

// GetSecret retrieves a secret from the configured provider
func (sm *SecretManager) GetSecret(ctx context.Context, path string) (string, error) {
	return sm.provider.GetSecret(ctx, path)
}

// ListSecrets lists available secrets at the given path
func (sm *SecretManager) ListSecrets(ctx context.Context, path string) ([]string, error) {
	return sm.provider.ListSecrets(ctx, path)
}

// Environment Variables Provider
type envProvider struct{}

func newEnvProvider() *envProvider {
	return &envProvider{}
}

func (p *envProvider) GetSecret(_ context.Context, path string) (string, error) {
	value := os.Getenv(path)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not found", path)
	}
	return value, nil
}

func (p *envProvider) ListSecrets(_ context.Context, prefix string) ([]string, error) {
	var secrets []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			secrets = append(secrets, parts[0])
		}
	}
	return secrets, nil
}
