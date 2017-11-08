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
	"context"
	"fmt"
	"net"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/log"
	"github.com/katzenpost/core/pki"
	"github.com/op/go-logging"
)

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
	if _, _, err := net.SplitHostPort(cfg.Address); err != nil {
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
	c.log.Debugf("Post(ctx, %d, %v, d)", epoch, signingKey.PublicKey())
	return fmt.Errorf("nonvoting/client: Post() is unimplemented")
}

func (c *client) Get(ctx context.Context, epoch uint64) (*pki.Document, error) {
	c.log.Debugf("Get(ctx, %d)", epoch)
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
