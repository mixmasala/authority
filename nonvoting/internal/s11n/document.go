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
	"fmt"

	"github.com/katzenpost/core/crypto/eddsa"
	"github.com/katzenpost/core/pki"
)

// SignDocument signs and serializes the document with the provided signing key.
func SignDocument(signingKey *eddsa.PrivateKey, base *pki.Document) (string, error) {
	return "", fmt.Errorf("s11n: not implemented yet")
}
