package internAPIUtils

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/awnumar/memguard"
)

// get or set key byte slice via JSON
func GetEnclaveFromJSON(r *http.Request) (*memguard.Enclave, string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", err
	}
	defer r.Body.Close()

	var b64Encoded string
	err = json.Unmarshal(body, &b64Encoded)
	if err != nil {
		return nil, "", err
	}

	splitted := strings.Split(b64Encoded, ";")

	keyPart, add := splitted[0], splitted[1]

	decodedBytes, err := base64.StdEncoding.DecodeString(keyPart)
	if err != nil {
		return nil, "", err
	}

	enclave := memguard.NewEnclave(decodedBytes)

	return enclave, add, nil
}

func SetEnclaveAsJSON(w http.ResponseWriter, key *memguard.Enclave, add string) error {
	w.Header().Set("Content-Type", "application/json")

	lockedBuffer, err  := key.Open()
	if err != nil {
		return err
	}
	defer lockedBuffer.Destroy()

	b64Encoded := base64.StdEncoding.EncodeToString(lockedBuffer.Bytes())

	b64Encoded = b64Encoded + ";" + add

	json.NewEncoder(w).Encode(b64Encoded)

	return nil
}