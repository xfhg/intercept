package cmd

import (
	"time"

	"github.com/maypok86/otter"
)

var (
	ResultCache otter.CacheWithVariableTTL[string, string]
)

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

func storeResultInCache(key, value string) {
	ResultCache.Set(key, value, time.Hour)
}

func loadResultFromCache(key string) (string, bool) {
	return ResultCache.Get(key)
}
