// config.go - Katzenpost Directory Authority server configuration.
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

// Package config provides the Katzenpost Directory Authority server configuration.
package config

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/pelletier/go-toml"
)

// DirectoryDescriptor describe another Directory Authority server
type DirectoryDescriptor struct {
	// IdentityKey is the Directory Authority server's identity (signing) key.
	IdentityKey *eddsa.PublicKey

	// Address is a network address string
	// e.g. "127.0.0.1:8080"
	Address string

	// BaseURL in the URL prefix
	// (must not end in /)
	BaseURL string
}

// TomlConfig is the TOML configuration struct for the Directory Authority server instance
type TomlConfig struct {
	// IdentityPrivateKey is the Directory Authority server's identity (signing) key.
	IdentityPrivateKey string

	// Address is a network address string
	// e.g. "127.0.0.1:8080"
	Address string

	// BaseURL in the URL prefix
	// (must not end in /)
	BaseURL string

	// DataDir is the filepath where
	// this server stores directory and consensus files
	DataDir string

	// Peers is our set of peer directory authority servers
	// which we will vote amongst.
	Peers []DirectoryDescriptor
}

// Config is the deserialized configuration struct for configuring
// our dir-auth server instance
type Config struct {
	// IdentityPrivateKey is the Directory Authority server's identity (signing) key.
	IdentityPrivateKey *eddsa.PrivateKey

	// Address is a network address string
	// e.g. "127.0.0.1:8080"
	Address string

	// BaseURL in the URL prefix
	// (must not end in /)
	BaseURL string

	// DataDir is the filepath where
	// this server stores directory and consensus files
	DataDir string

	// Peers is our set of peer directory authority servers
	// which we will vote amongst.
	Peers []DirectoryDescriptor
}

// Load loads a configuration from a byte slice
func Load(b []byte) (*Config, error) {
	tomlCfg := new(TomlConfig)
	if err := toml.Unmarshal(b, tomlCfg); err != nil {
		return nil, err
	}
	cfg := new(Config)
	cfg.Address = tomlCfg.Address
	cfg.BaseURL = tomlCfg.BaseURL
	cfg.DataDir = tomlCfg.DataDir
	cfg.Peers = tomlCfg.Peers
	raw, err := base64.StdEncoding.DecodeString(string(tomlCfg.IdentityPrivateKey))
	if err != nil {
		return nil, err
	}
	cfg.IdentityPrivateKey = new(eddsa.PrivateKey)
	err = cfg.IdentityPrivateKey.FromBytes(raw)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFile loads, parses and validates the provided file and returns the
// Config.
func LoadFile(f string) (*Config, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return Load(b)
}
