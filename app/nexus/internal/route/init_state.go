//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state"
	"golang.org/x/crypto/pbkdf2"
)

func updateStateForInit(password string, adminTokenBytes, salt []byte) {
	iterationCount := env.Pbkdf2IterationCount()
	hashLength := env.ShaHashLength()
	passwordHash := pbkdf2.Key(
		[]byte(password), salt,
		iterationCount, hashLength, sha256.New,
	)

	state.SetAdminToken("spike." + string(adminTokenBytes))
	state.SetAdminCredentials(
		hex.EncodeToString(passwordHash),
		hex.EncodeToString(salt),
	)
}