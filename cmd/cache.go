package cmd

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/maypok86/otter"
)

// Result cache used across the application for ephemeral results (SARIF, goss outputs, etc)
var (
	ResultCache otter.CacheWithVariableTTL[string, string]
)

// initialiseCache initializes the ResultCache. Keep this name because other files call it.
func initialiseCache() {
	var err error
	// Initialize cache
	ResultCache, err = otter.MustBuilder[string, string](50).
		WithVariableTTL().
		CollectStats().
		Build()
	if err != nil {
		panic(err)
	}
}

// storeResultInCache stores a string value in the ResultCache under key
func storeResultInCache(key, value string) {
	// Default TTL 1 hour for results; adjust as needed
	ResultCache.Set(key, value, time.Hour)
}

// loadResultFromCache loads a string value from ResultCache and returns found boolean
func loadResultFromCache(key string) (string, bool) {
	return ResultCache.Get(key)
}

/*
 Policy cache
 - single canonical place for storing policy -> path mappings
 - ensures keys are normalized consistently (path separator, case on windows, trailing slash for directories)
 - callers may call StorePolicyInCache(key, policy) (we detect isDirectory automatically)
*/

// in-memory policy cache (maps normalized key -> Policy)
var (
	policyCache      = make(map[string]Policy)
	policyCacheMutex sync.RWMutex
)

// normalizeCacheKey returns a canonical form for a path to be used as a cache key.
// Steps:
//   - filepath.Clean
//   - filepath.ToSlash (so Windows backslashes become forward slashes)
//   - strip leading "./"
//   - on windows lowercase for case-insensitive matching
func normalizeCacheKey(p string) string {
	if p == "" {
		return ""
	}
	// Clean and normalize separators
	p = filepath.Clean(p)
	p = filepath.ToSlash(p)

	// Strip leading "./"
	if strings.HasPrefix(p, "./") {
		p = strings.TrimPrefix(p, "./")
	}

	// On Windows do case-insensitive matching
	if runtime.GOOS == "windows" {
		p = strings.ToLower(p)
	}

	return p
}

// normalizeDirectoryKey returns normalized directory key with trailing slash.
func normalizeDirectoryKey(p string) string {
	k := normalizeCacheKey(p)
	if k == "" {
		return k
	}
	if !strings.HasSuffix(k, "/") {
		k = k + "/"
	}
	return k
}

// StorePolicyInCache stores a policy using a normalized key. If isDirectory is true
// the key will be the directory form (trailing slash).
func StorePolicyInCache(path string, policy Policy, isDirectory bool) {
	var key string
	if isDirectory {
		key = normalizeDirectoryKey(path)
	} else {
		key = normalizeCacheKey(path)
	}

	policyCacheMutex.Lock()
	defer policyCacheMutex.Unlock()
	policyCache[key] = policy
}

// LoadPolicyFromCache retrieves a policy from cache by path.
// It tries an exact file match then a directory fallback.
func LoadPolicyFromCache(path string) (Policy, bool) {
	key := normalizeCacheKey(path)

	policyCacheMutex.RLock()
	p, ok := policyCache[key]
	policyCacheMutex.RUnlock()
	if ok {
		return p, true
	}

	// directory fallback using GetDirectory (keeps same behavior but normalized)
	dir := GetDirectory(path)
	dirKey := normalizeDirectoryKey(dir)
	if dirKey != "" {
		policyCacheMutex.RLock()
		p, ok = policyCache[dirKey]
		policyCacheMutex.RUnlock()
		if ok {
			return p, true
		}
	}

	return Policy{}, false
}

// PolicyExistsInCache checks for existence using normalized forms (file and directory).
func PolicyExistsInCache(path string) bool {
	key := normalizeCacheKey(path)
	policyCacheMutex.RLock()
	_, ok := policyCache[key]
	policyCacheMutex.RUnlock()
	if ok {
		return true
	}

	dirKey := normalizeDirectoryKey(path)
	if dirKey != "" {
		policyCacheMutex.RLock()
		_, ok = policyCache[dirKey]
		policyCacheMutex.RUnlock()
		if ok {
			return true
		}
	}

	return false
}

// DeletePolicyFromCache removes a policy by normalized file key.
func DeletePolicyFromCache(path string) {
	key := normalizeCacheKey(path)
	policyCacheMutex.Lock()
	delete(policyCache, key)
	policyCacheMutex.Unlock()
}

// ClearAllPoliciesFromCache clears the policy cache
func ClearAllPoliciesFromCache() {
	policyCacheMutex.Lock()
	policyCache = make(map[string]Policy)
	policyCacheMutex.Unlock()
}

// ListAllPolicyCacheKeys returns a slice of normalized keys present in the cache
func ListAllPolicyCacheKeys() []string {
	policyCacheMutex.RLock()
	defer policyCacheMutex.RUnlock()
	keys := make([]string, 0, len(policyCache))
	for k := range policyCache {
		keys = append(keys, k)
	}
	return keys
}

// LoadAllPoliciesFromCache returns a map copy of all policies
func LoadAllPoliciesFromCache() map[string]Policy {
	policyCacheMutex.RLock()
	defer policyCacheMutex.RUnlock()
	out := make(map[string]Policy, len(policyCache))
	for k, v := range policyCache {
		out[k] = v
	}
	return out
}

// UpdatePolicyInCache updates or inserts a policy; uses isDirectory to determine key form.
func UpdatePolicyInCache(path string, policy Policy, isDirectory bool) {
	StorePolicyInCache(path, policy, isDirectory)
}
