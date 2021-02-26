package rci

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rciApi "github.com/tdx/go-rci/api"
	logApi "github.com/tdx/go/api/log"
)

type svc struct {
	tag        string
	log        logApi.Logger
	pathGlobal string
	pathLocal  string
	file2hook  map[string]string
	mu         sync.RWMutex
	hooks      map[string]*rciApi.Hook
}

var (
	tenHours = time.Duration(10 * time.Hour)
)

// New ...
func New(
	log logApi.Logger,
	name, pathGlobal, pathLocal string,
	filesCommands bool) rciApi.Service {

	s := &svc{
		log:        log,
		tag:        "[RCI " + name + "]:",
		pathGlobal: pathGlobal,
		pathLocal:  filepath.Join(pathLocal, name, "rci"),
		file2hook:  make(map[string]string),
		hooks:      make(map[string]*rciApi.Hook),
	}

	s.addBuiltInHooks()

	if filesCommands {
		go s.run()
	}

	return s
}

//
func (s *svc) run() {

	s.log.Info().Println(s.tag, "global path::", s.pathGlobal)
	s.log.Info().Println(s.tag, "local path::", s.pathLocal)

	if err := os.MkdirAll(s.pathLocal, 0770); err != nil {
		s.log.Error().Println(s.tag, "create dir", s.pathLocal, "failed:", err)
	}

	pathAsync := filepath.Join(s.pathLocal, "async")
	if err := os.MkdirAll(pathAsync, 0770); err != nil {
		s.log.Error().Println(s.tag, "create dir", pathAsync, "failed:", err)
	}

	for {
		s.walkPath(s.pathGlobal)
		s.walkPath(s.pathLocal)

		time.Sleep(time.Minute)
	}
}

func (s *svc) walkPath(path string) {

	updated := 0

	// update hooks under s.path
	err := filepath.Walk(
		path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// skip async commands dir
			if strings.Contains(path, "async") {
				if err = chkAsync(path, info); err != nil {
					s.log.Error().Println(s.tag, path, err)
				}
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

			s.log.Info().Println(s.tag,
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
		s.log.Error().Println(s.tag, "walk", path, "failed:", err)
	}

	if updated > 0 {
		s.log.Info().Println(s.tag, "updated hooks in", path, updated)
	}

	if deleted := s.chkDeleted(); deleted > 0 {
		s.log.Info().Println(s.tag, "deleted hooks in", path, deleted)
		s.delete()
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
