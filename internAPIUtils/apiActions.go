package internAPIUtils

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/awnumar/memguard"
)

// get or set key byte slice via JSON
func getKeyFromJSON(r *http.Request) (error, *memguard.Enclave) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err, nil
	}
	defer r.Body.Close()

	var b64Encoded string
	err = json.Unmarshal(body, &b64Encoded)
	if err != nil {
		return err, nil
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(b64Encoded)
	if err != nil {
		return err, nil
	}

	enclave := memguard.NewEnclave(decodedBytes)

	return nil, enclave
}

func setKeyAsJSON(w http.ResponseWriter, key *memguard.Enclave) error {
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