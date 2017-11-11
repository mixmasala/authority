// server.go - Katzenpost Directory Authority server API
// Copyright (C) 2017  David Stainton.
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

// Package config provides the Katzenpost Directory Authority server API
package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/jonboulle/clockwork"
	"github.com/katzenpost/authority/config"
	"github.com/katzenpost/authority/scheduler"
	"github.com/katzenpost/authority/statemachine"
	"github.com/katzenpost/core/epochtime"
	"github.com/katzenpost/core/pki"
)

type MixServer interface {
	PutDescriptor(descriptor *pki.MixDescriptor) error
	GetConsensus(epoch int64) (*pki.Document, error)
}

type DirectoryServer interface {
	GetProposedDirectory(epoch int64) (*pki.Document, error)
	GetFinalDirectory()
}

type Storage interface {
	PutDirectory(d *pki.Document) error
	GetDirectory(epoch int64) (*pki.Document, error)
}

// Server implements our Directory Authority mix network
// consensus document service.
type Server struct {
	config       *config.Config
	scheduler    *scheduler.EpochScheduler
	statemachine statemachine.StateMachine
	clock        clockwork.Clock
}

// consensusHandler handles requests for consensus documents
func consensusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("REQUEST URI %s\n", r.RequestURI)
}

// uploadHandler handles client uploads
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("output", string(buf.Bytes()))
}

func (s *Server) VoteExchangeHandler() {
	fmt.Println("vote phase")
}

func (s *Server) SignatureExchangeHandler() {
	fmt.Println("signature exchange phase")
}

func (s *Server) Start() error {
	var err error
	t := epochtime.New(s.clock)
	_, elapsed, _ := t.Now()
	s.statemachine, err = statemachine.New(elapsed, s.VoteExchangeHandler, s.SignatureExchangeHandler)
	if err != nil {
		return err
	}
	go s.scheduler.Run(s.statemachine)
	http.HandleFunc(fmt.Sprintf("%s/upload/", s.config.BaseURL), uploadHandler)
	http.HandleFunc(fmt.Sprintf("%s/consensus/", s.config.BaseURL), consensusHandler)
	http.ListenAndServe(s.config.Address, nil)
	return nil
}

func New(cfg *config.Config, ctx context.Context, clock clockwork.Clock, sm statemachine.StateMachine) *Server {
	s := Server{
		config:       cfg,
		scheduler:    scheduler.New(ctx, clock),
		statemachine: sm,
		clock:        clock,
	}
	return &s
}
