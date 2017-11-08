// s11n.go - Katzenpost Non-voting authority serialization routines.
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
	"encoding/json"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"gopkg.in/square/go-jose.v2"
)

const nodeDescriptorVersion = "nonvoting-v0"

type nodeDescriptor struct {
	// Version uniquely identifies the descriptor format as being for the
	// non-voting authority so that it can be rejected when unexpectedly
	// posted to, or received from an authority, or if the version changes.
	Version string

	pki.MixDescriptor
}

func signDescriptor(signingKey *eddsa.PrivateKey, base *pki.MixDescriptor) (string, error) {
	d := new(nodeDescriptor)
	d.MixDescriptor = *base
	d.Version = nodeDescriptorVersion

	// Serialize the descriptor.
	payload, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	// Sign the descriptor.
	k := jose.SigningKey{
		Algorithm: jose.EdDSA,
		Key:       *signingKey.InternalPtr(),
	}
	signer, err := jose.NewSigner(k, nil)
	if err != nil {
		return "", err
	}
	signed, err := signer.Sign(payload)
	if err != nil {
		return "", err
	}

	// Serialize the key, descriptor and signature.
	return signed.CompactSerialize()
}
