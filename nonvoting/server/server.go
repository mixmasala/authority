// server.go - Katzenpost non-voting authority server.
// Copyright (C) 2017  Yawning Angel.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package server implements the Katzenpost non-voting authority server.
//
// The non-voting authority server is intended to be a stop gap for debugging
// and testing and is likely only suitable for very small networks where the
// lack of distributed trust and or quality of life features is a non-issue.
package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/katzenpost/authority/nonvoting/server/config"
	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/crypto/rand"
	"github.com/katzenpost/core/log"
	"github.com/op/go-logging"
)

// Server is a non-voting authority server instance.
type Server struct {
	sync.WaitGroup

	cfg *config.Config

	identityKey *eddsa.PrivateKey

	logBackend *log.Backend
	log        *logging.Logger

	worker    *worker
	listeners []*http.Server

	fatalErrCh chan error
	haltedCh   chan interface{}
	haltOnce   sync.Once
}

func (s *Server) initDataDir() error {
	const dirMode = os.ModeDir | 0700
	d := s.cfg.Authority.DataDir

	// Initialize the data directory, by ensuring that it exists (or can be
	// created), and that it has the appropriate permissions.
	if fi, err := os.Lstat(d); err != nil {
		// Directory doesn't exist, create one.
		if !os.IsNotExist(err) {
			return fmt.Errorf("authority: failed to stat() DataDir: %v", err)
		}
		if err = os.Mkdir(d, dirMode); err != nil {
			return fmt.Errorf("authority: failed to create DataDir: %v", err)
		}
	} else {
		if !fi.IsDir() {
			return fmt.Errorf("authority: DataDir '%v' is not a directory", d)
		}
		if fi.Mode() != dirMode {
			return fmt.Errorf("authority: DataDir '%v' has invalid permissions '%v'", d, fi.Mode())
		}
	}

	return nil
}

func (s *Server) initLogging() error {
	p := s.cfg.Logging.File
	if !s.cfg.Logging.Disable && s.cfg.Logging.File != "" {
		if !filepath.IsAbs(p) {
			p = filepath.Join(s.cfg.Authority.DataDir, p)
		}
	}

	var err error
	s.logBackend, err = log.New(p, s.cfg.Logging.Level, s.cfg.Logging.Disable)
	if err == nil {
		s.log = s.logBackend.GetLogger("authority")
	}
	return err
}

func (s *Server) initListener(addr string) (*http.Server, error) {
	const (
		readTimeout  = 30 * time.Second
		writeTimeout = 30 * time.Second
	)

	l := new(http.Server)
	l.Addr = addr
	l.Handler = s
	l.ReadTimeout = readTimeout
	l.WriteTimeout = writeTimeout
	l.ErrorLog = s.logBackend.GetGoLogger("httpd", "ERROR")
	return l, l.ListenAndServe()
}

// Shutdown cleanly shuts down a given Server instance.
func (s *Server) Shutdown() {
	s.haltOnce.Do(func() { s.halt() })
}

func (s *Server) halt() {
	s.log.Notice("Starting graceful shutdown.")

	// Halt the listeners.
	for idx, l := range s.listeners {
		if l != nil {
			l.Close()
		}
		s.listeners[idx] = nil
	}

	// Wait for all the connections to terminate.
	s.Wait()

	// Halt the worker.
	if s.worker != nil {
		s.worker.halt()
		s.worker = nil
	}

	// XXX: Persist state to disk.

	s.identityKey.Reset()
	close(s.fatalErrCh)

	s.log.Notice("Shutdown complete.")
	close(s.haltedCh)
}

// New returns a new Server instance parameterized with the specific
// configuration.
func New(cfg *config.Config) (*Server, error) {
	s := new(Server)
	s.cfg = cfg
	s.fatalErrCh = make(chan error)
	s.haltedCh = make(chan interface{})

	// Do the early initialization and bring up logging.
	if err := s.initDataDir(); err != nil {
		return nil, err
	}
	if err := s.initLogging(); err != nil {
		return nil, err
	}

	s.log.Notice("Katzenpost is still pre-alpha.  DO NOT DEPEND ON IT FOR STRONG SECURITY OR ANONYMITY.")
	if s.cfg.Logging.Level == "DEBUG" {
		s.log.Warning("Unsafe Debug logging is enabled.")
	}

	// Initialize the authority identity key.
	var err error
	identityPrivateKeyFile := filepath.Join(s.cfg.Authority.DataDir, "identity.private.pem")
	identityPublicKeyFile := filepath.Join(s.cfg.Authority.DataDir, "identity.public.pem")
	if s.identityKey, err = eddsa.Load(identityPrivateKeyFile, identityPublicKeyFile, rand.Reader); err != nil {
		s.log.Errorf("Failed to initialize identity: %v", err)
		return nil, err
	}
	s.log.Noticef("Authority identity public key is: %s", s.identityKey.PublicKey())

	// Past this point, failures need to call s.Shutdown() to do cleanup.
	isOk := false
	defer func() {
		if !isOk {
			s.Shutdown()
		}
	}()

	// Start the fatal error watcher.
	go func() {
		err, ok := <-s.fatalErrCh
		if !ok {
			return
		}
		s.log.Warningf("Shutting down due to error: %v", err)
		s.Shutdown()
	}()

	// XXX: Initialize the persistence store and restore state.

	// Start up the worker.
	s.worker = newWorker(s)

	// Start up the listeners.
	for _, v := range s.cfg.Authority.Addresses {
		l, err := s.initListener(v)
		if err != nil {
			s.log.Errorf("Failed to start listener '%v': %v", v, err)
			continue
		}
		s.listeners = append(s.listeners, l)
	}
	if len(s.listeners) == 0 {
		s.log.Errorf("Failed to start all listeners.")
		return nil, fmt.Errorf("authority: failed to start all listeners")
	}

	isOk = true
	return s, nil
}
