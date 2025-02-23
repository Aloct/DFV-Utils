package encryptUtils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/awnumar/memguard"
)

type kekRefs struct {
	KEKID string
	Manager string
	KEKDB string
	KEKCache string
	Cached time.Time
	CacheDuration time.Duration
}

type DEKCombKEK struct {
	Dek []byte
	KekDB kekRefs
}

type keyFetcher interface {
	GetKey(id string) (any, error)
	SetKey(id string, key any, d *time.Duration) error
}

// dek is encrypted via the kek which is referenced by kekRef, passed dek must be unencrypted
func CreateDEKCombKEK(dek *memguard.Enclave, kekDB string, kekCache string, cacheDuration time.Duration, managerC string, managerP int) (*DEKCombKEK, error) {
	kek, err := CreateAESKey(32)
	if err != nil {
		return nil, err
	}

	dekEncrypted, err := AesEncryption(kek, dek)
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
		return nil, err
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

	fmt.Println(resData["data"].(string))
	return &DEKCombKEK{
		Dek: dekEncrypted,
		KekDB: kekRefs{
			KEKID: resData["data"].(string),
			Manager: fmt.Sprintf("%s:%d", managerC, managerP),
			KEKDB: kekDB,
			KEKCache: kekCache,
			Cached: time.Now().Add(-cacheDuration),
			CacheDuration: cacheDuration,
		},
	}, nil
}

// https://developer.ibm.com/tutorials/docker-dev-db/

func (dc *DEKCombKEK) GetDEK(dbF keyFetcher, cacheF keyFetcher) (*memguard.Enclave, error) {
	// get KEK from decryptKEK()
	fmt.Println("GetDEK")
	fmt.Println(dbF)
	kek, err := dc.decryptKEK(dbF, cacheF)
	if err != nil {
		return nil, err
	}

	dek, err := AesDecryption(dc.Dek, kek)
	if err != nil {
		return nil, err
	}

	return dek, nil
}

func (dc *DEKCombKEK) decryptKEK(dbF keyFetcher, cacheF keyFetcher) (*memguard.Enclave, error) {
	// check if KEK is cached
	var kek *memguard.Enclave

	if (time.Since(dc.KekDB.Cached) < dc.KekDB.CacheDuration) {
		kekRaw, err := cacheF.GetKey(dc.KekDB.KEKID)
		if err != nil {
			return nil, err
		}

		kek = kekRaw.(*memguard.Enclave)
	} else {
		fmt.Println(dc.KekDB.KEKID)
		fmt.Println("decryptKEK")
		fmt.Println(dbF)
		keyRaw, err := dbF.GetKey(dc.KekDB.KEKID)
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
	
			// cache key to increase decryption speed
			cacheF.SetKey(dc.KekDB.KEKID, b64Encoded, &dc.KekDB.CacheDuration)
			dc.KekDB.Cached = time.Now()
	
			return memguard.NewEnclave([]byte(base64.StdEncoding.EncodeToString([]byte(b64Encoded)))), nil
		}()

		if err != nil {
			return nil, err
		}
	}

	return kek, nil
}

func CreateAESKey(keySize int) (*memguard.Enclave, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
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

func CreateECCKey() (*memguard.Enclave, error) {
	private, err := CreateAESKey(32)
	if err != nil {
		return nil, err
	}

	return private, nil
}