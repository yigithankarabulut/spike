//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package route

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"golang.org/x/crypto/pbkdf2"

	"github.com/spiffe/spike/app/nexus/internal/state"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/internal/net"
)

// routeAdminLogin handles HTTP requests for administrator authentication using PBKDF2-SHA256
// password hashing. It validates the provided password against stored credentials and issues
// a JWT token upon successful authentication.
//
// The function implements the following security measures:
//   - PBKDF2-SHA256 password hashing with 600,000 iterations (OWASP recommended minimum)
//   - Constant-time password comparison using crypto/hmac.Equal
//   - Salted password hashing
//   - JWT token-based authentication
//
// Authentication Process:
//  1. Reads and validates the request body containing the password
//  2. Retrieves stored admin credentials (password hash and salt)
//  3. Decodes the stored salt and password hash from hex format
//  4. Generates a new hash from the provided password using PBKDF2
//  5. Performs constant-time comparison of password hashes
//  6. Issues a signed JWT token upon successful authentication
//
// Parameters:
//   - w: http.ResponseWriter to write the HTTP response
//   - r: *http.Request containing the incoming HTTP request
//   - audit: *log.AuditEntry for logging audit information
//
// Returns:
//   - error: nil if authentication succeeds, or an error if:
//   - Request body cannot be read or parsed
//   - Salt or password hash cannot be decoded
//   - Password is invalid
//   - Admin token is not set
//   - JWT token cannot be signed
//   - Response body cannot be marshaled
//
// Request Body:
//
//	{
//	  "password": "admin_password"
//	}
//
// Response Status Codes:
//   - 200 OK: Successfully authenticated
//   - 400 Bad Request: Invalid request body
//   - 401 Unauthorized: Invalid password
//   - 500 Internal Server Error: Server-side errors
//
// Response Body on Success:
//
//	{
//	  "token": "signed_jwt_token"
//	}
//
// Response Body on Error:
//
//	{
//	  "err": "error_code"
//	}
//
// Security Notes:
//   - Uses PBKDF2-SHA256 with 600,000 iterations for password hashing
//   - Output hash length is 32 bytes (256 bits)
//   - Implements constant-time comparison to prevent timing attacks
func routeAdminLogin(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("routeAdminLogin", "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = "login"

	// TODO: signature should be `w, r` for consistency.
	requestBody := net.ReadRequestBody(r, w)
	if requestBody == nil {
		return errors.New("failed to read request body")
	}

	request := net.HandleRequest[
		reqres.AdminLoginRequest, reqres.AdminLoginResponse](
		requestBody, w,
		reqres.AdminLoginResponse{Err: reqres.ErrBadInput},
	)
	if request == nil {
		return errors.New("failed to parse request body")
	}

	password := request.Password
	creds := state.AdminCredentials()
	passwordHash := creds.PasswordHash
	salt := creds.Salt

	s, err := hex.DecodeString(salt)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding salt",
			"err", err.Error())

		body := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault,
		}, w)
		if body == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, body, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return errors.New("failed to decode salt")
	}

	// TODO: duplication.
	// TODO: make this configurable.
	iterationCount := 600_000 // Minimum OWASP recommendation for PBKDF2-SHA256
	hashLength := 32          // 256 bits output

	ph := pbkdf2.Key(
		[]byte(password), s,
		iterationCount, hashLength, sha256.New,
	)

	b, err := hex.DecodeString(passwordHash)
	if err != nil {
		log.Log().Error("routeAdminLogin",
			"msg", "Problem decoding password hash",
			"err", err.Error())

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "OK")
		return errors.New("failed to decode password hash")
	}

	if !hmac.Equal(ph, b) {
		log.Log().Info("routeAdminLogin", "msg", "Invalid password")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrUnauthorized}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusUnauthorized, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return errors.New("invalid password")
	}

	adminToken := state.AdminToken()
	if adminToken == "" {
		log.Log().Error("routeAdminLogin", "msg", "Admin token not set")

		responseBody := net.MarshalBody(reqres.AdminLoginResponse{
			Err: reqres.ErrServerFault}, w)
		if responseBody == nil {
			return errors.New("failed to marshal response body")
		}

		net.Respond(http.StatusInternalServerError, responseBody, w)
		log.Log().Info("routeAdminLogin", "msg", "unauthorized")
		return errors.New("admin token not set")
	}

	signedToken := net.CreateJwt(adminToken, w)
	if signedToken == "" {
		return errors.New("failed to sign token")
	}

	responseBody := net.MarshalBody(reqres.AdminLoginResponse{
		Token: signedToken,
	}, w)
	if responseBody == nil {
		return errors.New("failed to marshal response body")
	}

	net.Respond(http.StatusOK, responseBody, w)
	log.Log().Info("routeAdminLogin", "msg", "authorized")
	return nil
}