// state.go - Katzenpost non-voting authority server state.
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

package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"github.com/op/go-logging"
)

var (
	errGone   = errors.New("authority: Request is too far in the past")
	errNotYet = errors.New("authority: Document is not ready yet")
)

type state struct {
	sync.WaitGroup
	sync.RWMutex

	s   *Server
	log *logging.Logger

	authorizedMixes     map[[eddsa.PublicKeySize]byte]bool
	authorizedProviders map[[eddsa.PublicKeySize]byte]string

	haltCh chan interface{}
}

func (s *state) halt() {
	close(s.haltCh)
	s.Wait()

	// XXX: Persist the state to disk.
}

func (s *state) worker() {
	defer func() {
		s.log.Debugf("Halting worker.")
		s.Done()
	}()

	for {
		select {
		case <-s.haltCh:
			s.log.Debugf("Termianting gracefully.")
			return
		}
	}
}

func (s *state) isDescriptorAuthorized(desc *pki.MixDescriptor) bool {
	var tmp [eddsa.PublicKeySize]byte
	copy(tmp[:], desc.IdentityKey.Bytes())

	switch desc.Layer {
	case 0:
		return s.authorizedMixes[tmp]
	case pki.LayerProvider:
		name, ok := s.authorizedProviders[tmp]
		if !ok {
			return false
		}
		return name == desc.Name
	default:
		return false
	}
}

func (s *state) onDescriptorUpload(desc *pki.MixDescriptor, epoch uint64) error {
	// XXX: Write me.
	return fmt.Errorf("state: not implemented yet")
}

func (s *state) documentForEpoch(epoch uint64) ([]byte, error) {
	// XXX: Write me.
	return nil, fmt.Errorf("state: not implemented yet")
}

func newState(s *Server) *state {
	st := new(state)
	st.s = s
	st.log = s.logBackend.GetLogger("state")
	st.haltCh = make(chan interface{})

	// Initialize the authorized peer tables.
	st.authorizedMixes = make(map[[eddsa.PublicKeySize]byte]bool)
	for _, v := range st.s.cfg.Mixes {
		var tmp [eddsa.PublicKeySize]byte
		copy(tmp[:], v.IdentityKey.Bytes())
		st.authorizedMixes[tmp] = true
	}
	st.authorizedProviders = make(map[[eddsa.PublicKeySize]byte]string)
	for _, v := range st.s.cfg.Mixes {
		var tmp [eddsa.PublicKeySize]byte
		copy(tmp[:], v.IdentityKey.Bytes())
		st.authorizedProviders[tmp] = v.Identifier
	}

	// XXX: Initialize the persistence store and restore state.

	st.Add(1)
	go st.worker()
	return st
}
