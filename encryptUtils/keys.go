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
	"net/http"
	"strings"
	"time"

	"github.com/awnumar/memguard"
)

type kekRefs struct {
	KEKID         string
	Manager       string
	KEKDB         string
	KEKCache      string
	Cached        time.Time
	CacheDuration time.Duration
}

type DEKCombKEK struct {
	Dek   []byte
	KekDB kekRefs
}

type keyFetcher interface {
	GetKey(id string, stringToKey interface{}) (any, error)
	SetKey(id string, key any, d *time.Duration, keyToString interface{}) error

	GetData(query string, values []any) (any, error)
	SetData(query string, values []any, n *time.Duration) error
}

// dek is encrypted via the kek which is referenced by kekRef, passed dek must be unencrypted
func CreateDEKCombKEK(dek *memguard.Enclave, kekDB string, kekCache string, cacheDuration time.Duration, managerC string, managerP int) (*DEKCombKEK, error) {
	kek, err := CreateAESKey(32)
	if err != nil {
		return nil, err
	}

	dekEncrypted, err := AesEncryption(dek, kek)
	if err != nil {
		return nil, err
	}

	keyBuf, err := kek.Open()
	if err != nil {
		return nil, err
	}
	defer keyBuf.Destroy()

	jsonData, err := json.Marshal(base64.StdEncoding.EncodeToString(keyBuf.Bytes()) + ";" + kekDB)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%d/registerKEK", managerC, managerP), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		time.Sleep(10 * time.Second)
		res, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.New("failed to add key comb to auth manager")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resData := make(map[string]interface{})
	err = json.Unmarshal(body, &resData)
	if err != nil {
		return nil, err
	}

	return &DEKCombKEK{
		Dek: dekEncrypted,
		KekDB: kekRefs{
			KEKID:         resData["data"].(string),
			Manager:       fmt.Sprintf("%s:%d", managerC, managerP),
			KEKDB:         kekDB,
			KEKCache:      kekCache,
			Cached:        time.Now().Add(-cacheDuration),
			CacheDuration: cacheDuration,
		},
	}, nil
}

// https://developer.ibm.com/tutorials/docker-dev-db/

func (dc *DEKCombKEK) GetDEK(dbF keyFetcher, cacheF keyFetcher, keyToString interface{}, stringToKey interface{}) (*memguard.Enclave, error) {
	// get KEK from decryptKEK()
	keyToStringC := keyToString.(func(keyRaw any) (string, error))
	stringToKeyC := stringToKey.(func(keyRaw any) ([]byte, error))

	kek, err := dc.decryptKEK(dbF, cacheF, keyToStringC, stringToKeyC)
	if err != nil {
		return nil, err
	}

	dek, err := AesDecryption(dc.Dek, kek)
	if err != nil {
		return nil, err
	}

	return dek, nil
}

func (dc *DEKCombKEK) decryptKEK(dbF keyFetcher, cacheF keyFetcher, keyToString func(keyRaw any) (string, error), stringToKey func(keyRaw any) ([]byte, error)) (*memguard.Enclave, error) {
	// check if KEK is cached
	var kek *memguard.Enclave

	if time.Since(dc.KekDB.Cached) < dc.KekDB.CacheDuration {
		kekRaw, err := cacheF.GetKey(dc.KekDB.KEKID, stringToKey)
		if err != nil {
			return nil, err
		}

		kek = kekRaw.(*memguard.Enclave)
	} else {
		keyRaw, err := dbF.GetKey(dc.KekDB.KEKID, stringToKey)
		if err != nil {
			return nil, err
		}

		// process key and decrypt via manager
		jsonData, err := json.Marshal(base64.StdEncoding.EncodeToString(keyRaw.([]byte)))
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/decryptKEK", dc.KekDB.Manager), bytes.NewBuffer(jsonData))
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

			enclave := memguard.NewEnclave([]byte(b64Raw))

			// cache key to increase decryption speed
			cacheF.SetKey(dc.KekDB.KEKID, enclave, &dc.KekDB.CacheDuration, keyToString)
			dc.KekDB.Cached = time.Now()

			b64Decoded, err := base64.StdEncoding.DecodeString(b64Raw)
			if err != nil {
				return nil, err
			}

			return memguard.NewEnclave(b64Decoded), nil
		}()

		if err != nil {
			return nil, err
		}
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
	toZero(keyRaw)

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
