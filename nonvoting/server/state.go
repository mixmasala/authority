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
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/epochtime"
	"github.com/katzenpost/core/pki"
	"github.com/op/go-logging"
)

var (
	errGone   = errors.New("authority: Request is too far in the past")
	errNotYet = errors.New("authority: Document is not ready yet")
)

type descriptor struct {
	desc *pki.MixDescriptor
	raw  []byte
}

type state struct {
	sync.WaitGroup
	sync.RWMutex

	s   *Server
	log *logging.Logger

	authorizedMixes     map[[eddsa.PublicKeySize]byte]bool
	authorizedProviders map[[eddsa.PublicKeySize]byte]string

	documents   map[uint64][]byte
	descriptors map[uint64]map[[eddsa.PublicKeySize]byte]*descriptor

	haltCh chan interface{}

	bootstrapEpoch uint64
}

func (s *state) halt() {
	close(s.haltCh)
	s.Wait()

	// XXX: Gracefully close the persistence store.
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
	pk := desc.IdentityKey.ByteArray()

	switch desc.Layer {
	case 0:
		return s.authorizedMixes[pk]
	case pki.LayerProvider:
		name, ok := s.authorizedProviders[pk]
		if !ok {
			return false
		}
		return name == desc.Name
	default:
		return false
	}
}

func (s *state) onDescriptorUpload(rawDesc []byte, desc *pki.MixDescriptor, epoch uint64) error {
	// Note: Caller ensures that the epoch is the current epoch +- 1.
	pk := desc.IdentityKey.ByteArray()

	s.Lock()
	defer s.Unlock()

	// Get the public key -> descriptor map for the epoch.
	m, ok := s.descriptors[epoch]
	if !ok {
		m = make(map[[eddsa.PublicKeySize]byte]*descriptor)
		s.descriptors[epoch] = m
	}

	// Check for redundant uploads.
	if d, ok := m[pk]; ok {
		// If the descriptor changes, then it will be rejected to prevent
		// nodes from reneging on uploads.
		if !bytes.Equal(d.raw, rawDesc) {
			return fmt.Errorf("state: Node %v: Conflicting descriptor for epoch %v", desc.IdentityKey, epoch)
		}

		// Redundant uploads that don't change are harmless.
		return nil
	}

	// Ok, this is a new descriptor.
	if s.documents[epoch] != nil {
		// If there is a document already, the descriptor is late, and will
		// never appear in a document, so reject it.
		return fmt.Errorf("state: Node %v: Late descriptor upload for for epoch %v", desc.IdentityKey, epoch)
	}

	// Store the raw descriptor and the parsed struct.
	d := new(descriptor)
	d.desc = desc
	d.raw = rawDesc
	m[pk] = d // XXX: Persist d.raw to disk.

	// XXX: Kick the worker.

	s.log.Debugf("Node %v: Sucessfully submitted descriptor for epoch %v.", desc.IdentityKey, epoch)

	return nil
}

func (s *state) documentForEpoch(epoch uint64) ([]byte, error) {
	const generationDeadline = 45 * time.Minute

	s.RLock()
	defer s.RUnlock()

	// If we have a serialized document, return it.
	if d, ok := s.documents[epoch]; ok {
		return d, nil
	}

	// Otherwise, return an error based on the time.
	now, _, till := epochtime.Now()
	switch epoch {
	case now:
		// Check to see if we are doing a bootstrap, and it's possible that
		// we may decide to publish a document at some point ignoring the
		// standard schedule.
		if now == s.bootstrapEpoch {
			return nil, errNotYet
		}

		// We missed the deadline to publish a descriptor for the current
		// epoch, so we will never be able to service this request.
		return nil, errGone
	case now + 1:
		// If it's past the time by which we should have generated a document
		// then we will never be able to service this.
		if till < generationDeadline {
			return nil, errGone
		}
		return nil, errNotYet
	default:
		if epoch < now {
			// Requested epoch is in the past, and it's not in the cache.
			// We will never be able to satisfy this request.
			return nil, errGone
		}
		return nil, fmt.Errorf("state: Request for invalid epoch: %v", epoch)
	}

	// NOTREACHED
}

func newState(s *Server) *state {
	st := new(state)
	st.s = s
	st.log = s.logBackend.GetLogger("state")
	st.haltCh = make(chan interface{})

	// Initialize the authorized peer tables.
	st.authorizedMixes = make(map[[eddsa.PublicKeySize]byte]bool)
	for _, v := range st.s.cfg.Mixes {
		pk := v.IdentityKey.ByteArray()
		st.authorizedMixes[pk] = true
	}
	st.authorizedProviders = make(map[[eddsa.PublicKeySize]byte]string)
	for _, v := range st.s.cfg.Mixes {
		pk := v.IdentityKey.ByteArray()
		st.authorizedProviders[pk] = v.Identifier
	}

	// XXX: Initialize the persistence store and restore state.
	st.documents = make(map[uint64][]byte)
	st.descriptors = make(map[uint64]map[[eddsa.PublicKeySize]byte]*descriptor)

	// Do a "rapid" bootstrap where we will generate and publish a Document
	// for the current epoch regardless of time iff:
	//
	//  * We do not have a persisted Document for the epoch.
	//  * (Checked in worker) *All* nodes publish a descriptor.
	//
	// This could be relaxed a bit, but it's primarily intended for debugging.
	epoch, _, _ := epochtime.Now()
	if _, ok := st.documents[epoch]; !ok {
		st.bootstrapEpoch = epoch
	}

	st.Add(1)
	go st.worker()
	return st
}
