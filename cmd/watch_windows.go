//go:build windows && amd64
// +build windows,amd64

package cmd

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/segmentio/ksuid"
)

const eventCacheDuration = 2000 * time.Millisecond // Configurable duration

type cachedEvent struct {
	event fsnotify.Event
	timer *time.Timer
}

func watchPaths(paths ...string) {
	if len(paths) < 1 {
		log.Fatal().Msg("must specify at least one path to watch")
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Msgf("creating a new watcher: %s", err)
	}
	defer w.Close()

	// Start listening for events.
	switch observeMode {
	case "last":
		go watchLoopLastEvent(w, paths)
	case "first":
		go watchLoopFirstEvent(w, paths)
	case "all":
		go watchLoopAllEvents(w, paths)
	default:
		go watchLoopFirstEvent(w, paths)
	}

	// Add all paths from the commandline.
	for _, p := range paths {
		err = w.Add(p)
		if err != nil {
			log.Fatal().Msgf("Failed to watch %q: %s", p, err)
		}
		log.Debug().Str("Observe", p).Msg("Watching Path")
	}

	log.Debug().Msg("Path Watcher Ready")
	<-make(chan struct{}) // Block forever
}

func watchLoopAllEvents(w *fsnotify.Watcher, watchedPaths []string) {
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Error().Msgf("Watcher error: %s", err)

		case event, ok := <-w.Events:
			if !ok {
				return
			}

			log.Debug().Msgf("Watcher caught [%s] on [%s]", event.Op.String(), event.Name)

			// Process the event
			processEvent(event)

			// Re-add the watch for the file if it was removed or renamed
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				for _, path := range watchedPaths {
					if path == event.Name {
						// Wait a short time for the file to be recreated/renamed
						time.Sleep(100 * time.Millisecond)
						if err := w.Add(path); err != nil {
							log.Error().Msgf("Failed to re-add watch for %s: %s", path, err)
						} else {
							log.Debug().Msgf("Re-added watch for %s", path)
						}
						break
					}
				}
			}
		}
	}
}

func watchLoopFirstEvent(w *fsnotify.Watcher, watchedPaths []string) {
	eventCache := make(map[string]time.Time)
	var cacheMutex sync.Mutex

	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Error().Msgf("Watcher error: %s", err)

		case event, ok := <-w.Events:
			if !ok {
				return
			}

			log.Debug().Msgf("Watcher caught [%s] on [%s]", event.Op.String(), event.Name)

			// Check if we should process this event
			cacheMutex.Lock()
			lastEventTime, exists := eventCache[event.Name]
			now := time.Now()
			if !exists || now.Sub(lastEventTime) > eventCacheDuration {
				// Process the event and update the cache
				eventCache[event.Name] = now
				cacheMutex.Unlock()

				// Process the event in a goroutine to avoid blocking
				go processEvent(event)

				// Clean up old cache entries
				go cleanEventCache(&eventCache, &cacheMutex)
			} else {
				cacheMutex.Unlock()
				log.Debug().Msgf("Ignored duplicate event for [%s] within cache duration", event.Name)
			}

			// Re-add the watch for the file if it was removed or renamed
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				for _, path := range watchedPaths {
					if path == event.Name {
						// Wait a short time for the file to be recreated/renamed
						time.Sleep(100 * time.Millisecond)
						if err := w.Add(path); err != nil {
							log.Error().Msgf("Failed to re-add watch for %s: %s", path, err)
						} else {
							log.Debug().Msgf("Re-added watch for %s", path)
						}
						break
					}
				}
			}
		}
	}
}

func watchLoopLastEvent(w *fsnotify.Watcher, watchedPaths []string) {
	eventCache := make(map[string]*cachedEvent)
	var cacheMutex sync.Mutex

	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Error().Msgf("Watcher error: %s", err)

		case event, ok := <-w.Events:
			if !ok {
				return
			}

			log.Debug().Msgf("Watcher caught [%s] on [%s]", event.Op.String(), event.Name)

			cacheMutex.Lock()
			if cached, exists := eventCache[event.Name]; exists {
				// Stop the existing timer
				cached.timer.Stop()
				// Update the cached event
				cached.event = event
				// Reset the timer
				cached.timer.Reset(eventCacheDuration)
			} else {
				// Create a new timer for this event
				timer := time.AfterFunc(eventCacheDuration, func() {
					processLastEvent(event.Name, &eventCache, &cacheMutex)
				})
				eventCache[event.Name] = &cachedEvent{event: event, timer: timer}
			}
			cacheMutex.Unlock()

			// Re-add the watch for the file if it was removed or renamed
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				for _, path := range watchedPaths {
					if path == event.Name {
						// Wait a short time for the file to be recreated/renamed
						time.Sleep(100 * time.Millisecond)
						if err := w.Add(path); err != nil {
							log.Error().Msgf("Failed to re-add watch for %s: %s", path, err)
						} else {
							log.Debug().Msgf("Re-added watch for %s", path)
						}
						break
					}
				}
			}
		}
	}
}

func processLastEvent(path string, cache *map[string]*cachedEvent, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	if cached, exists := (*cache)[path]; exists {
		// Process the most recent event
		processEvent(cached.event)
		// Remove the event from the cache
		delete(*cache, path)
	}
}

func cleanEventCache(cache *map[string]time.Time, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()

	now := time.Now()
	for k, v := range *cache {
		if now.Sub(v) > eventCacheDuration {
			delete(*cache, k)
		}
	}
}

func processEvent(e fsnotify.Event) {
	log.Debug().Str("fs", e.Name).Msg(e.String())

	// Normalize event path so Windows backslashes and relative paths match cached keys
	eventPath := normalizeCacheKey(e.Name)

	policy, ok := LoadPolicyFromCache(eventPath)

	log.Info().Msgf("Processing event [%s] on [%s]", e.Op.String(), e.Name)

	// If we didn't find a policy for the exact file, try the directory key
	if !ok {
		directoryCheck := GetDirectory(e.Name)
		dirKey := normalizeDirectoryKey(directoryCheck)
		log.Debug().Str("directory", directoryCheck).Msg(e.String())
		policy, ok = LoadPolicyFromCache(dirKey)
	}

	if !ok {
		log.Debug().Str("path", e.Name).Msg("No policy found in cache for event")
		return
	}

	// If this is a runtime policy, run goss-based processing
	if policy.Type == "runtime" {
		// Generate a run id for this observe action
		policy.RunID = ksuid.New().String()

		// Prepare file list (the event file)
		filePaths := []string{e.Name}

		// Launch runtime processing in a goroutine so the watcher stays responsive
		go func(pol Policy, paths []string) {
			log.Info().Str("policy", pol.ID).Msgf("Triggering runtime policy for paths: %v", paths)
			if err := ProcessRuntimeType(pol, gossPath, "", paths, true); err != nil {
				log.Error().Err(err).Str("policy", pol.ID).Msg("Runtime processing failed")
			} else {
				log.Info().Str("policy", pol.ID).Msg("Runtime processing completed")
			}
		}(policy, filePaths)

		return
	}

	// For non-runtime types, leave existing behavior unchanged (no-op here).
	// If you want to trigger audit/action for other policy types on Windows,
	// add the logic here.
	log.Debug().Str("policy", policy.ID).Str("type", policy.Type).Msg("Non-runtime observe event — no action taken on Windows by default")
}
