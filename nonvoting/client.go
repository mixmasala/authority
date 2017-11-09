// client.go - Katzenpost Non-voting authority client.
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

package nonvoting

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/log"
	"github.com/katzenpost/core/pki"
	"github.com/katzenpost/core/utils"
	"github.com/op/go-logging"
	"golang.org/x/net/context/ctxhttp"
)

const (
	v0postBase    = "v0/post/"
	v0getBase     = "v0/get/"
	joseMimeType  = "application/jose"
	clientTimeout = 30 * time.Second
)

var httpClient = &http.Client{Timeout: clientTimeout}

// ClientConfig is a nonvoting authority pki.Client instance.
type ClientConfig struct {
	// LogBackend is the `core/log` Backend instance to use for logging.
	LogBackend *log.Backend

	// Address is the authority's address to connect to for posting and
	// fetching documents.
	Address string

	// PublicKey is the authority's public key to use when validating documents.
	PublicKey *eddsa.PublicKey
}

func (cfg *ClientConfig) validate() error {
	if cfg.LogBackend == nil {
		return fmt.Errorf("nonvoting/client: LogBackend is mandatory")
	}
	if err := utils.EnsureAddrIPPort(cfg.Address); err != nil {
		return fmt.Errorf("nonvoting/client: Invalid Address: %v", err)
	}
	if cfg.PublicKey == nil {
		return fmt.Errorf("nonvoting/client: PublicKey is mandatory")
	}
	return nil
}

type client struct {
	cfg *ClientConfig
	log *logging.Logger
}

func (c *client) Post(ctx context.Context, epoch uint64, signingKey *eddsa.PrivateKey, d *pki.MixDescriptor) error {
	c.log.Debugf("Post(ctx, %d, %v, %v)", epoch, signingKey.PublicKey(), d)

	// Ensure that the descriptor we are about to post is well formed.
	if err := isDescriptorWellFormed(d, epoch); err != nil {
		return err
	}

	// Make a serialized + signed + serialized descriptor.
	signed, err := signDescriptor(signingKey, d)
	if err != nil {
		return err
	}
	c.log.Debugf("Signed descriptor: '%v'", signed)

	// Post it to the right place.
	u := postURLForEpoch(c.cfg.Address, epoch)
	c.log.Debugf("Posting descriptor to: %v", u)

	r := bytes.NewReader([]byte(signed))
	resp, err := ctxhttp.Post(ctx, httpClient, u, joseMimeType, r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted, http.StatusNoContent, http.StatusNotModified:
		return nil
	default:
		// TODO: The authority rejected the POST for some reason, the
		// right thing to do is to probably return an error indicating
		// that the server should give up trying to upload a descriptor
		// for this epoch.
		//
		// See: https://github.com/Katzenpost/server/issues/11
		return fmt.Errorf("nonvoting/client: Post() rejected by authority: %v", resp.StatusCode)
	}

	// NOTREACHED
}

func (c *client) Get(ctx context.Context, epoch uint64) (*pki.Document, error) {
	c.log.Debugf("Get(ctx, %d)", epoch)

	// Download the document.
	u := getURLForEpoch(c.cfg.Address, epoch)
	c.log.Debugf("Getting document from: %v", u)

	resp, err := ctxhttp.Get(ctx, httpClient, u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Validate the document.

	return nil, fmt.Errorf("nonvoting/client: Get() is unimplemented")
}

// NewClient constructs a new pki.Client instance.
func NewClient(cfg *ClientConfig) (pki.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nonvoting/client: cfg is mandatory")
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	c := new(client)
	c.cfg = cfg
	c.log = cfg.LogBackend.GetLogger("pki/nonvoting/client")

	return c, nil
}

func postURLForEpoch(addr string, epoch uint64) string {
	u := &url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   fmt.Sprintf("%v%v", v0postBase, epoch),
	}
	return u.String()
}

func getURLForEpoch(addr string, epoch uint64) string {
	u := &url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   fmt.Sprintf("%v%v", v0getBase, epoch),
	}
	return u.String()
}
