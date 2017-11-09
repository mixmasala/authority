package main

import (
	"bytes"
	"fmt"
	"github.com/katzenpost/core/pki"
	"net/http"
)

type StateMachine interface {
	SetState()
	Advance()
}

type Scheduler interface {
	Get()
}

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

type Config struct {
	// BaseURL in the URL prefix
	// (must not end in /)
	BaseURL string
	// DataDir is the filepath where
	// this server stores directory and consensus files
	DataDir string
	// NetAddr is a network address string
	// e.g. "127.0.0.1:8080"
	NetAddr string
}

// Server implements our Directory Authority mix network
// consensus document service.
type Server struct {
	config *Config
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

func (s *Server) Start() {
	http.HandleFunc(fmt.Sprintf("%s/upload/", s.config.BaseURL), uploadHandler)
	http.HandleFunc(fmt.Sprintf("%s/consensus/", s.config.BaseURL), consensusHandler)
	http.ListenAndServe(s.config.NetAddr, nil)
}

func New(config *Config) *Server {
	s := Server{
		config: config,
	}
	return &s
}

func main() {
	cfg := Config{
		BaseURL: "/B",
		DataDir: "/home/user/non-critical/gopath/src/github.com/katzenpost/authority/server",
		NetAddr: "127.0.0.1:8080",
	}
	s := New(&cfg)
	s.Start()
}
