package cmd

import (
	"os"
	"strings"
)

// FilterPoliciesByAnyTags filters policies that match any of the provided tags
func FilterPoliciesByAnyTags(policies []Policy, tags []string) []Policy {
	if len(tags) == 0 {
		return policies
	}

	var filteredPolicies []Policy
	for _, policy := range policies {
		for _, tag := range tags {
			if containsTag(policy.Metadata.Tags, strings.TrimSpace(tag)) {
				filteredPolicies = append(filteredPolicies, policy)
				break
			}
		}
	}
	return filteredPolicies
}

// FilterPoliciesByAllTags filters policies that match all of the provided tags
func FilterPoliciesByAllTags(policies []Policy, tags []string) []Policy {
	if len(tags) == 0 {
		return policies
	}

	var filteredPolicies []Policy
	for _, policy := range policies {
		if containsAllTags(policy.Metadata.Tags, tags) {
			filteredPolicies = append(filteredPolicies, policy)
		}
	}
	return filteredPolicies
}

// FilterPoliciesByEnvironment filters policies that match the specified environment
func FilterPoliciesByEnvironment(policies []Policy, environment string) []Policy {
	if environment == "" {
		return policies // No filtering if no environment is specified
	}

	var filteredPolicies []Policy
	for _, policy := range policies {
		for _, enforcement := range policy.Enforcement {
			if enforcement.Environment == "all" || strings.EqualFold(enforcement.Environment, environment) {
				filteredPolicies = append(filteredPolicies, policy)
				break
			}
		}
	}
	return filteredPolicies
}

// containsTag checks if a slice of strings contains a specific tag
func containsTag(slice []string, tag string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, tag) {
			return true
		}
	}
	return false
}

// containsAllTags checks if a slice of strings contains all the specified tags
func containsAllTags(slice []string, tags []string) bool {
	for _, tag := range tags {
		if !containsTag(slice, strings.TrimSpace(tag)) {
			return false
		}
	}
	return true
}

// DetectEnvironment detects the current environment based on environment variables
func DetectEnvironment() string {
	envVars := []string{
		"ENV",
		"ENVIRONMENT",
		"NODE_ENV",
		"RAILS_ENV",
		"RACK_ENV",
		"DJANGO_ENV",
		"DJANGO_SETTINGS_MODULE",
		"FLASK_ENV",
		"APP_ENV",
		"SYMFONY_ENV",
		"SPRING_PROFILES_ACTIVE",
		"ASPNETCORE_ENVIRONMENT",
		"GO_ENV",
		"GIN_MODE",
		"MIX_ENV",
		"REACT_APP_ENV",
		"VUE_APP_ENV",
		"NG_ENV",
		"RUST_ENV",
		"PLACK_ENV",
		"KOTLIN_ENV",
		"RUST_ENV",
		"SPRING_PROFILES_ACTIVE",
	}

	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			return value
		}
	}

	return ""
}
