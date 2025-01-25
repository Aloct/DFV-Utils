package internAPIUtils

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/awnumar/memguard"
)

// get or set key byte slice via JSON
func GetKeyFromJSON(r *http.Request) (*memguard.Enclave, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var b64Encoded string
	err = json.Unmarshal(body, &b64Encoded)
	if err != nil {
		return nil, err
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Encoded)
	if err != nil {
		return nil, err
	}

	enclave := memguard.NewEnclave(decodedBytes)

	return enclave, nil
}

func SetKeyAsJSON(w http.ResponseWriter, key *memguard.Enclave) error {
	w.Header().Set("Content-Type", "application/json")

	lockedBuffer, err  := key.Open()
	if err != nil {
		return err
	}
	defer lockedBuffer.Destroy()

	b64Encoded := base64.StdEncoding.EncodeToString(lockedBuffer.Bytes())

	json.NewEncoder(w).Encode(b64Encoded)

	return nil
}