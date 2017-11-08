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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
	"gopkg.in/square/go-jose.v2"
)

const (
	nodeDescriptorVersion = "nonvoting-v0"
	nodeJosePubKeyHdr     = "EdDSAPublicKey"
)

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
	//
	// HACK: Frustratingly enough, the library doesn't appear to support
	// embedding JWKs for EdDSA signatures, so just jam the public key into
	// the header.
	k := jose.SigningKey{
		Algorithm: jose.EdDSA,
		Key:       *signingKey.InternalPtr(),
	}
	sOpts := &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			nodeJosePubKeyHdr: signingKey.PublicKey(),
		},
	}
	signer, err := jose.NewSigner(k, sOpts)
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

func verifyAndParseDescriptor(epoch uint64, b []byte) (*pki.MixDescriptor, error) {
	signed, err := jose.ParseSigned(string(b))
	if err != nil {
		return nil, err
	}

	// The descriptor should be signed by it's own key, but the library
	// doesn't give a convenient way to extract the payload when it hasn't
	// been verified, so reach into the header instead.
	//
	// HACK: This should be an embedded JWK, but it's just Base64 encoded
	// into the JWS Header because of library limitations.
	if len(signed.Signatures) != 1 {
		return nil, fmt.Errorf("nonvoting: Expected 1 signature, got: %v", len(signed.Signatures))
	}
	alg := signed.Signatures[0].Header.Algorithm
	if alg != "EdDSA" {
		return nil, fmt.Errorf("nonvoting: Unsupported signature algorithm: '%v'", alg)
	}
	s, ok := signed.Signatures[0].Header.ExtraHeaders[nodeJosePubKeyHdr]
	if !ok {
		return nil, fmt.Errorf("nonvoting: Failed to find descriptor public key")
	}
	var candidatePk eddsa.PublicKey
	pkStr, ok := s.(string)
	if !ok {
		return nil, fmt.Errorf("nonvoting: Pathologically malfored descriptor public key")
	}
	if err = candidatePk.UnmarshalText([]byte(pkStr)); err != nil {
		return nil, err
	}

	// Verify that the descriptor is signed by the key in the header.
	payload, err := signed.Verify(*candidatePk.InternalPtr())
	if err != nil {
		return nil, err
	}

	// Parse the payload.
	d := new(nodeDescriptor)
	if err = json.Unmarshal(payload, d); err != nil {
		return nil, err
	}

	// Ensure the descriptor is well formed.
	if d.Version != nodeDescriptorVersion {
		return nil, fmt.Errorf("nonvoting: Invalid Descriptor Version: '%v'", d.Version)
	}
	if d.LinkKey == nil {
		return nil, fmt.Errorf("nonvoting: Descriptor missing 'LinkKey'")
	}
	if d.IdentityKey == nil {
		return nil, fmt.Errorf("nonvoting: Descriptor missing 'IdentityKey'")
	}
	// XXX: Addresses
	// XXX: Layer
	// XXX: MixKeys (Use `epoch`).

	// And as the final check, ensure that the key embedded in the descriptor
	// matches the key embedded in the JOSE header, that we used to validate
	// the signature.
	//
	// MixDescriptors returned from this function are essentially known to be
	// "well formed", and correctly self-signed.
	if !bytes.Equal(candidatePk.Bytes(), d.IdentityKey.Bytes()) {
		return nil, fmt.Errorf("nonvoting: Descriptor signing key mismatch")
	}
	return &d.MixDescriptor, nil
}
