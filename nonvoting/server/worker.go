// worker.go - Katzenpost non-voting authority server worker.
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
	"fmt"
	"sync"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"github.com/op/go-logging"
)

type worker struct {
	sync.WaitGroup
	sync.RWMutex

	s   *Server
	log *logging.Logger

	authorizedMixes     map[[eddsa.PublicKeySize]byte]bool
	authorizedProviders map[[eddsa.PublicKeySize]byte]string

	haltCh chan interface{}
}

func (w *worker) halt() {
	close(w.haltCh)
	w.Wait()
}

func (w *worker) doWork() {
	defer func() {
		w.log.Debugf("Halting worker.")
		w.Done()
	}()

	for {
		select {
		case <-w.haltCh:
			w.log.Debugf("Termianting gracefully.")
			return
		}
	}
}

func (w *worker) isDescriptorAuthorized(desc *pki.MixDescriptor) bool {
	var tmp [eddsa.PublicKeySize]byte
	copy(tmp[:], desc.IdentityKey.Bytes())

	switch desc.Layer {
	case 0:
		return w.authorizedMixes[tmp]
	case pki.LayerProvider:
		name, ok := w.authorizedProviders[tmp]
		if !ok {
			return false
		}
		return name == desc.Name
	default:
		return false
	}
}

func (w *worker) onDescriptorUpload(desc *pki.MixDescriptor, epoch uint64) error {
	// XXX: Write me.
	return nil
}

func (w *worker) documentForEpoch(epoch uint64) ([]byte, error) {
	// XXX: Write me.
	return nil, fmt.Errorf("worker: not implemented yet")
}

func newWorker(s *Server) *worker {
	w := new(worker)
	w.s = s
	w.log = s.logBackend.GetLogger("worker")
	w.haltCh = make(chan interface{})

	// Initialize the authorized peer tables.
	w.authorizedMixes = make(map[[eddsa.PublicKeySize]byte]bool)
	for _, v := range w.s.cfg.Mixes {
		var tmp [eddsa.PublicKeySize]byte
		copy(tmp[:], v.IdentityKey.Bytes())
		w.authorizedMixes[tmp] = true
	}
	w.authorizedProviders = make(map[[eddsa.PublicKeySize]byte]string)
	for _, v := range w.s.cfg.Mixes {
		var tmp [eddsa.PublicKeySize]byte
		copy(tmp[:], v.IdentityKey.Bytes())
		w.authorizedProviders[tmp] = v.Identifier
	}

	w.Add(1)
	go w.doWork()
	return w
}
