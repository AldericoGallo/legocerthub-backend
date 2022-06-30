package acme

import (
	"crypto"
	"encoding/json"
	"net/http"
	"strings"
)

// NewAccountPayload is the payload used to post to ACME newAccount
type NewAccountPayload struct {
	TosAgreed bool     `json:"termsOfServiceAgreed"`
	Contact   []string `json:"contact"`
}

// LE response to account data post/update
type AcmeAccountResponse struct {
	Contact   []string `json:"contact"`
	CreatedAt string   `json:"createdAt"`
	Status    string   `json:"status"`
	Location  *string  `json:"-"` // omit because it is in the header
	// -- also available but not in use
	// JsonWebKey jsonWebKey `json:"key"`
	// Orders     string     `json:"orders"`
	// InitialIP  string     `json:"initialIp"`
}

// Account response decoder
func unmarshalAccountResponse(bodyBytes []byte, headers http.Header) (response AcmeAccountResponse, err error) {
	// try to decode an error
	var errorResponse AcmeErrorResponse
	err = json.Unmarshal(bodyBytes, &errorResponse)
	if err == nil {
		// return error if acme response was an error
		return AcmeAccountResponse{}, errorResponse.Error()
	} else {
		// if error didn't decode, decode generally
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			return AcmeAccountResponse{}, err
		}
	}

	// kid isn't part of the JSON response, add it from the header.
	// ACME only returns this if not posting with kid, so have some logic
	// to set it to null if it isn't returned from the server
	if headers.Get("Location") != "" {
		response.Location = new(string)
		*response.Location = headers.Get("Location")
	} else {
		response.Location = nil
	}

	return response, nil
}

// CreatedAt() returns the created at time in unix format. If there is an error
// converting, return 0
func (response *AcmeAccountResponse) CreatedAtUnix() (int, error) {
	time, err := acmeToUnixTime(response.CreatedAt)
	if err != nil {
		return 0, err
	}

	return time, nil
}

// Email() returns an email address from the first string in the Contact slice.
// Any other contacts are discarded.
func (response *AcmeAccountResponse) Email() string {
	// if contacts are empty, email is blank
	if len(response.Contact) == 0 {
		return ""
	}

	return strings.TrimPrefix(response.Contact[0], "mailto:")
}

// NewAccount posts a secure message to the NewAccount URL of the directory
func (service *Service) NewAccount(payload NewAccountPayload, privateKey crypto.PrivateKey) (response AcmeAccountResponse, err error) {
	// Create ACME accountKey
	// Register account should never use kid, it must always use JWK
	var accountKey AccountKey
	accountKey.Key = privateKey
	accountKey.Kid = "" // no-op, just a reminder

	// post new-account
	bodyBytes, headers, err := service.postToUrlSigned(payload, service.dir.NewAccount, accountKey)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	// unmarshal response
	response, err = unmarshalAccountResponse(bodyBytes, headers)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	return response, nil
}

// UpdateAccountPayload is the payload used to update ACME accounts
type UpdateAccountPayload struct {
	Contact []string `json:"contact"`
}

// UpdateAccount posts a secure message to the kid of the account
// initially support only exists to update the email address
// TODO: key rotation and account deactivation
func (service *Service) UpdateAccount(payload UpdateAccountPayload, accountKey AccountKey) (response AcmeAccountResponse, err error) {

	// post account update
	bodyBytes, headers, err := service.postToUrlSigned(payload, accountKey.Kid, accountKey)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	// unmarshal response
	response, err = unmarshalAccountResponse(bodyBytes, headers)
	if err != nil {
		return AcmeAccountResponse{}, err
	}

	return response, nil
}