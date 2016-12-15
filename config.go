package main

import (
	"encoding/json"
	"log"
	"os"

	"sync"

	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	config *Config
	watchr *fsnotify.Watcher

	configLock *sync.RWMutex
)

func init() {
	configLock = &sync.RWMutex{}
}

//GetConfig Obtains config safely
func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func loadConfig(path string, fail bool) bool {
	configFile, err := os.Open(path)
	if err != nil {
		log.Println("Error opening config file", err.Error())
		if fail {
			os.Exit(1)
		} else {
			return false
		}
	}

	var temp *Config
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&temp); err != nil {
		log.Println("Error parsing config file", err.Error())
		if fail {
			os.Exit(1)
		} else {
			return false
		}
	}

	configLock.Lock()
	config = temp
	configLock.Unlock()
	return true
}

// WatchConfig Adds `path` to watchr and returns a reload channel that application should use to handle config reload.
func WatchConfig(path string) (chan struct{}, chan struct{}) {
	if watchr != nil {
		log.Fatalln("Already watching.")
	}

	var err error
	watchr, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed creating watcher", err.Error())
	}

	loadConfig(path, true)

	reloadChan := make(chan struct{})
	reloadedChan := make(chan struct{})
	isWorking := false

	go func() {
		for {
			select {
			case event := <-watchr.Events:
				if isWorking {
					log.Println("Already in process of reloading, try again later.")
					continue
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					isWorking = true
					go func() {
						log.Println("Config file change has been detected, attempting to parse new config.")
						ok := loadConfig(path, false)
						if !ok {
							log.Println("New config encountered error, continuing with old one.")
						} else {
							log.Println("New config is valid, reloading.")
							reloadChan <- struct{}{}
							// Prevents double reloads, thanks fsnotify...
							<-reloadedChan
							time.Sleep(1 * time.Second)
						}
						isWorking = false
					}()
				}

			case err := <-watchr.Errors:
				log.Println("Error:", err)
			}
		}
	}()

	err = watchr.Add(path)
	if err != nil {
		log.Fatalln("Failed to add config file to watcher", err)
	}

	return reloadChan, reloadedChan
}
