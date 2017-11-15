// document.go - Katzenpost Non-voting authority document s11n.
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

package s11n

import (
	"encoding/json"
	"fmt"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"gopkg.in/square/go-jose.v2"
)

const documentVersion = "nonvoting-document-v0"

type document struct {
	// Version uniquely identifies the document format as being for the
	// non-voting authority so that it can be rejected when unexpectedly
	// received or if the version changes.
	Version string

	pki.Document
}

// SignDocument signs and serializes the document with the provided signing key.
func SignDocument(signingKey *eddsa.PrivateKey, base *pki.Document) (string, error) {
	d := new(document)
	d.Document = *base
	d.Version = documentVersion

	// Serialize the document.
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

// VerifyAndParseDocument verifies the signautre and deserializes the document.
func VerifyAndParseDocument(b []byte, publicKey *eddsa.PublicKey, epoch uint64) (*pki.Document, error) {
	signed, err := jose.ParseSigned(string(b))
	if err != nil {
		return nil, err
	}

	// Sanity check the signing algorithm and number of signatures, and
	// validate the signature with the provided public key.
	if len(signed.Signatures) != 1 {
		return nil, fmt.Errorf("nonvoting: Expected 1 signature, got: %v", len(signed.Signatures))
	}
	alg := signed.Signatures[0].Header.Algorithm
	if alg != "EdDSA" {
		return nil, fmt.Errorf("nonvoting: Unsupported signature algorithm: '%v'", alg)
	}
	payload, err := signed.Verify(*publicKey.InternalPtr())
	if err != nil {
		return nil, err
	}

	// Parse the payload.
	d := new(document)
	if err = json.Unmarshal(payload, d); err != nil {
		return nil, err
	}

	// Ensure the document is well formed.
	if d.Version != documentVersion {
		return nil, fmt.Errorf("nonvoting: Invalid Document Version: '%v'", d.Version)
	}
	if d.Epoch != epoch {
		return nil, fmt.Errorf("nonvoting: Invalid Document Epoch: '%v'", d.Epoch)
	}
	// XXX: More checks for well-formedness.

	return &d.Document, nil
}
