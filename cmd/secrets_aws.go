package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type awsSecretsProvider struct {
	client *secretsmanager.Client
}

func newAWSSecretsProvider(cfg map[string]string) (*awsSecretsProvider, error) {
	ctx := context.Background()

	// Load AWS configuration options from the provided config map
	var opts []func(*config.LoadOptions) error

	// Region configuration
	if region := cfg["region"]; region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	// Custom endpoint configuration (useful for testing or local development)
	if endpoint := cfg["endpoint"]; endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}

	// Profile configuration
	if profile := cfg["profile"]; profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// Try loading from config first
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		// If default config fails, try using provided credentials
		if accessKey, secretKey := cfg["access_key_id"], cfg["secret_access_key"]; accessKey != "" && secretKey != "" {
			staticCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				cfg["session_token"], // Optional session token
			))

			opts = append(opts, config.WithCredentialsProvider(staticCreds))
			awsCfg, err = config.LoadDefaultConfig(ctx, opts...)
			if err != nil {
				return nil, fmt.Errorf("unable to load AWS config with provided credentials: %w", err)
			}
		} else {
			return nil, fmt.Errorf("unable to load AWS config and no credentials provided: %w", err)
		}
	}

	client := secretsmanager.NewFromConfig(awsCfg)
	return &awsSecretsProvider{client: client}, nil
}

func (p *awsSecretsProvider) GetSecret(ctx context.Context, path string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &path,
	}

	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get AWS secret: %w", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("no secret string found")
}

func (p *awsSecretsProvider) ListSecrets(ctx context.Context, path string) ([]string, error) {
	var secrets []string
	input := &secretsmanager.ListSecretsInput{}

	paginator := secretsmanager.NewListSecretsPaginator(p.client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list AWS secrets: %w", err)
		}

		for _, secret := range output.SecretList {
			if secret.Name != nil && strings.HasPrefix(*secret.Name, path) {
				secrets = append(secrets, *secret.Name)
			}
		}
	}

	return secrets, nil
}
