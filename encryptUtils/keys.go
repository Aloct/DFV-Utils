package encryptUtils

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	"github.com/awnumar/memguard"
)

var serviceMasterSalt = "authTempSalt1234"

type KEKRefs struct {
	Manager       string
	DB         string
	Algorithm string
	Scope string
	CachingType string
}

type DEKRefs struct {
	DB         string
	Algorithm string
	Scope string
	InnerScope string
	CachingType string
}

type DEKCombKEK struct {
	KeyMappingType string
	Scope string
	DEKInfos *DEKRefs
	KEKInfos *KEKRefs
}

type keyFetcher interface {
	GetKey(id, idType string, stringToKey interface{}) (any, error)
	SetKey(id, individualrelation, keyrelation, version string, key any, d *time.Duration, keyToString interface{}) error

	GetData(query string, values []any) (any, error)
	SetData(query string, values []any, n *time.Duration) error
}

type responseCreator interface {
	NewStdResponse(context string, data interface{}) interface{}
	NewKEKRegister(kekdb, scope, id, userBlind string) interface{}
	NewKEKIdentifier(kekdb, id, idType string) interface{}
	SetKeyMetaForMultipartReq(w *multipart.Writer, key *memguard.Enclave, metadata interface{}) (*multipart.Writer, error)
}

func CreateDEKRefs(db, algorithm, scope, innerScope, cachingType string) *DEKRefs {
	return &DEKRefs{
		DB:         db,
		Algorithm: algorithm,
		Scope: scope,
		InnerScope: innerScope,
		CachingType: cachingType,
	}
}

func CreateKEKRefs(manager, db, algorithm, scope, cachingType string) *KEKRefs {
	return &KEKRefs{
		Manager: 	 manager, //container:port
		DB:         db,
		Algorithm: algorithm,
		Scope: scope,
		CachingType: cachingType,
	}
}

func CreateDEKCombKEK(keyMappingType, scope string, dekRefs *DEKRefs, kekRefs *KEKRefs, resCreate responseCreator) (*DEKCombKEK, error) {
	return &DEKCombKEK{
		KeyMappingType: keyMappingType,
		Scope: scope,
		DEKInfos: dekRefs,
		KEKInfos: kekRefs,
	}, nil
}

// register new KEK
func (dc *DEKCombKEK) RegisterNewKEK(resCreate responseCreator, defaultDEKs... DEKRefs) error {
	var kek *memguard.Enclave
	var err error
	switch dc.KEKInfos.Algorithm {
	case "AES":
		kek, err = CreateAESKey(32)
		if err != nil {
			return err
		}
	}

	userBlind, err := CreateUserBlind(serviceMasterSalt, dc.Scope, "0", "KEK")
	if err != nil {
		return err
	}

	// keyString, err := KeyToString(keyBuf.Bytes())
	// if err != nil {
	// 	return err
	// }

	bodyMultiData := &bytes.Buffer{}
	w := multipart.NewWriter(bodyMultiData)

	subRequestPart, err := w.CreateFormField("subRequests")

	subRequests := make([]interface{}, len(defaultDEKs))
	for i := range defaultDEKs {
		subRequests = append(subRequests, resCreate.NewStdResponse("createDEKRefs", defaultDEKs[i]))
	}

	subRequestsJSON, err := json.Marshal(subRequests)
	if err != nil {
		return err
	}

	subRequestPart.Write(subRequestsJSON)

	kekRegister := resCreate.NewKEKRegister(dc.KEKInfos.DB, dc.Scope, "", userBlind)
	w, err = resCreate.SetKeyMetaForMultipartReq(w, kek, kekRegister)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bodyMultiData)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(10 * time.Second)
		req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bodyMultiData)
		if err != nil {
			return err
		}

		req.Header.Add("Content-Type", "application/json")
		res, err = client.Do(req)
		if err != nil {
			return err
		}
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New("failed to add key comb to auth manager")
	}

	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return err
	// }
	// defer res.Body.Close()

	// resData := make(map[string]interface{})
	// err = json.Unmarshal(body, &resData)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// register new DEK under a KEK => DEK-KEK-reference stored
func (dc *DEKCombKEK) RegisterNewDEK(individualRef string) error {
	var dek *memguard.Enclave
	var err error
	switch dc.KEKInfos.Algorithm {
	case "AES":
		dek, err = CreateAESKey(32)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// retrieve decrypted DEK to handle Data
func (dc *DEKCombKEK) GetDEK(id, idType string, dbF keyFetcher, stringToKey interface{}, resCreate responseCreator) (*memguard.Enclave, error) {
	if (idType != "keyrelation" && idType != "individualrelation" && idType != "uniqueid") {
		return nil, fmt.Errorf("invalid id type")
	}
	
	// get KEK from decryptKEK()
	stringToKeyC, ok := stringToKey.(func(keyRaw any) ([]byte, error))
	if !ok {
		return nil, errors.New("invalid stringToKey function provided")
	}

	// get DEK and retrieve information to get related KEK 
	dekRaw, err := dbF.GetKey(id, idType, stringToKeyC)
	if err != nil {
		return nil, err
	}

	kek, err := dc.getKEK(id, idType, dbF, stringToKeyC, resCreate)
	if err != nil {
		return nil, err
	}

	dek, err := AesDecryption(dekRaw.([]byte), kek)
	if err != nil {
		return nil, err
	}

	return dek, nil
}

// retrieve KEK from manager to de/encrypt DEK
func (dc *DEKCombKEK) getKEK(id, idType string, dbF keyFetcher, stringToKey func(keyRaw any) ([]byte, error), resCreate responseCreator) (*memguard.Enclave, error) {
	// check if KEK is cached
	var kek *memguard.Enclave

	keyRaw, err := dbF.GetKey(id, idType, stringToKey)
	if err != nil {
		return nil, err
	}

	// keyString, err := KeyToString(keyRaw.([]byte))
	// if err != nil {
	// 	return nil, err
	// }

	jsonData, err := json.Marshal(resCreate.NewStdResponse("getKEK", resCreate.NewKEKIdentifier(keyString, dc.KEKInfos.DB)))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/decryptKEK", dc.KEKInfos.Manager), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusFound {
		return nil, errors.New("failed to decrypt KEK in auth manager")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	kek, err = func() (*memguard.Enclave, error) {
		var b64Encoded string
		err = json.Unmarshal(body, &b64Encoded)
		if err != nil {
			return nil, err
		}

		splitted := strings.Split(b64Encoded, ";")
		b64Raw := splitted[0]

		b64Decoded, err := base64.StdEncoding.DecodeString(b64Raw)
		if err != nil {
			return nil, err
		}

		return memguard.NewEnclave(b64Decoded), nil
	}()

	if err != nil {
		return nil, err
	}

	return kek, nil
}

func CreateAESKey(keySize int) (*memguard.Enclave, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 && keySize != 64 {
		return nil, errors.New("no valid key byte slice provided")
	}

	keyRaw := make([]byte, keySize)
	if _, err := rand.Read(keyRaw); err != nil {
		return nil, errors.New("failed to create key")
	}

	key := memguard.NewEnclave(keyRaw)
	coreutils.ToZero(keyRaw)

	return key, nil
}

func CreateECCKey(publicDBStore keyFetcher) (*memguard.Enclave, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	publicDBStore.SetData("INSERT INTO kstore (id, k_val, created_at, last_swap) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON DUPLICATE KEY UPDATE k_val = VALUES(k_val), created_at = CURRENT_TIMESTAMP, last_swap = CURRENT_TIMESTAMP", []any{"1", hex.EncodeToString(public)}, nil)

	return memguard.NewEnclave(private), nil
}