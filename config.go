package main

import (
	"encoding/json"
	"log"
	"os"

	"sync"

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
func WatchConfig(path string) chan bool {
	if watchr != nil {
		log.Fatalln("Already watching.")
	}

	var err error
	watchr, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln("Failed creating watcher", err.Error())
	}

	loadConfig(path, true)

	reloadChan := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watchr.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Config file change has been detected... trying new config...", event.Name)
					ok := loadConfig(path, false)
					if !ok {
						log.Println("New config encountered error, continuing with old one.")
					} else {
						reloadChan <- true
					}
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

	return reloadChan
}
