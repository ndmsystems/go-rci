package rci

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	rciApi "github.com/tdx/go-rci/api"
	logApi "github.com/tdx/go/api/log"
)

type svc struct {
	tag       string
	log       logApi.Logger
	path      string
	file2hook map[string]string
	mu        sync.RWMutex
	hooks     map[string]*rciApi.Hook
}

// New ...
func New(
	log logApi.Logger,
	name, path string,
	filesCommands bool) rciApi.Service {

	s := &svc{
		log:       log,
		tag:       "[RCI " + name + "]:",
		path:      filepath.Join(path, name, "rci"),
		file2hook: make(map[string]string),
		hooks:     make(map[string]*rciApi.Hook),
	}

	s.addBuiltInHooks()

	if filesCommands {
		go s.run(s.log)
	}

	return s
}

//
func (s *svc) run(log logApi.Logger) {

	log.Info().Println(s.tag, "find in path:", s.path)

	if err := os.MkdirAll(s.path, 0700); err != nil {
		log.Error().Println(s.tag, "create dir", s.path, "failed:", err)
	}

	for {
		updated := 0

		// update hooks under s.path
		err := filepath.Walk(
			s.path,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				var (
					update bool
					oldCmd *rciApi.Hook
				)

				if Hook, ok := s.file2hook[path]; ok {
					// lock read free for updating routine
					oldCmd = s.hooks[Hook]
				}

				if oldCmd != nil {
					if oldCmd.Size != info.Size() ||
						oldCmd.ModTime != info.ModTime() {

						update = true
					}
				} else {
					update = true
				}

				log.Info().Println(s.tag,
					path, info.Size(), info.ModTime(), update)

				if !update {
					return nil
				}

				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				cmd := new(rciApi.Hook)
				if err = json.Unmarshal(data, &cmd); err != nil {
					return err
				}

				cmd.Size = info.Size()
				cmd.ModTime = info.ModTime()
				cmd.FileName = path

				s.file2hook[path] = cmd.Hook

				s.mu.Lock()
				s.hooks[cmd.Hook] = cmd
				s.mu.Unlock()

				updated++

				return nil
			})

		if err != nil {
			log.Error().Println(s.tag, "walk", s.path, "failed:", err)
		}

		if updated > 0 {
			log.Info().Println(s.tag, "updated hooks:", updated)
		}

		if deleted := s.chkDeleted(); deleted > 0 {
			log.Info().Println(s.tag, "deleted hooks:", deleted)
			s.delete()
		}

		time.Sleep(time.Minute)
	}
}

//
func (s *svc) chkDeleted() int {
	deleted := 0

	for path, hook := range s.file2hook {
		if !fileExists(path) {
			if cmd, ok := s.hooks[hook]; ok {
				deleted++
				cmd.Deleted = true
			}
		}
	}

	return deleted
}

//
func (s *svc) delete() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for hook, cmd := range s.hooks {
		if cmd.Deleted {
			delete(s.hooks, hook)
			delete(s.file2hook, cmd.FileName)
		}
	}
}

//
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
