package encryptUtils

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	wrapperUtils "github.com/Aloct/DFV-Utils/dataHandling/storageWrapper"
	internAPIUtils "github.com/Aloct/DFV-Utils/internAPIUtils/apiConfig"
	"github.com/awnumar/memguard"
)

var serviceMasterSalt = []byte("serviceMasterSalt1234")
var client = &http.Client{}

type KEKRefs struct {
	Manager     string
	DB          string
	Algorithm   string
	Scope       string
	CachingType string
}

type DEKRefs struct {
	DB          string
	Algorithm   string
	Scope       string
	InnerScope  string
	CachingType string
}

type DEKCombKEK struct {
	KeyMappingType string
	Scope          string
	DEKInfos       *DEKRefs
	KEKInfos       *KEKRefs
}

type PartHandler func(io.Reader) (interface{}, error)

func CreateDEKRefs(db, algorithm, scope, innerScope, cachingType string) *DEKRefs {
	return &DEKRefs{
		DB:          db,
		Algorithm:   algorithm,
		Scope:       scope,
		InnerScope:  innerScope,
		CachingType: cachingType,
	}
}

func CreateKEKRefs(manager, db, algorithm, scope, cachingType string) *KEKRefs {
	return &KEKRefs{
		Manager:     manager, //container:port
		DB:          db,
		Algorithm:   algorithm,
		Scope:       scope,
		CachingType: cachingType,
	}
}

func CreateDEKCombKEK(keyMappingType, scope string, dekRefs *DEKRefs, kekRefs *KEKRefs) (*DEKCombKEK, error) {
	return &DEKCombKEK{
		KeyMappingType: keyMappingType,
		Scope:          scope,
		DEKInfos:       dekRefs,
		KEKInfos:       kekRefs,
	}, nil
}

// register new KEK
func (dc *DEKCombKEK) RegisterNewKEK(pool wrapperUtils.DBPool, publicDB wrapperUtils.MySQLWrapper, defaultDEKs ...internAPIUtils.DEKRegister) error {
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

	bodyMultiData := &bytes.Buffer{}
	w := multipart.NewWriter(bodyMultiData)

	subRequests := make([]internAPIUtils.StdRequest, len(defaultDEKs))
	for i := range defaultDEKs {
		subRequests = append(subRequests, internAPIUtils.NewStdRequest("createDEKRefs", defaultDEKs[i]))
	}
	w, err = internAPIUtils.SetSubRequestsForMultipartReq(w, subRequests)
	if err != nil {
		return err
	}

	kekRegister := internAPIUtils.NewKEKRegister(dc.KEKInfos.DB, dc.Scope, "", userBlind)
	w, err = internAPIUtils.SetKeyMetaForMultipartReq(w, kek, kekRegister)
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

	if defaultDEKs != nil {
		handlers := internAPIUtils.NewHandlerMap("subRequests")
		resData, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
		if err != nil {
			return err
		}

		subResponses, ok := resData["subRequests"].(internAPIUtils.GroupResponse)
		if !ok {
			return errors.New("invalid response format")
		}

		dekDB, err := pool.NewSQLWrapper(dc.DEKInfos.DB)
		if err != nil {
			return err
		}
		if dekDB.DB == nil {
			err := dekDB.Connect(context.Background(), 2)
			if err != nil {
				return err
			}
		}

		for _, stdRespRaw := range subResponses.Subs {
			var stdResp internAPIUtils.StdResponse
			err := json.Unmarshal(stdRespRaw.([]byte), &stdResp)
			if err != nil {
				return err
			}

			dekObj, ok := stdResp.Data.(internAPIUtils.DEKIdentifier)
			if !ok {
				return errors.New("invalid DEK object format")
			}

			dc.setNewDEKFromSet(dekObj, dekDB, kek, publicDB)
		}
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

func (dc *DEKCombKEK) setNewDEKFromSet(keySet internAPIUtils.DEKIdentifier, dekDB *wrapperUtils.MySQLWrapper, kek *memguard.Enclave, publicDB wrapperUtils.MySQLWrapper) error {
	dekKEKBlind := keySet.KEKBlind

	relationBlind, err := HashBlind(serviceMasterSalt, dekKEKBlind)
	if err != nil {
		return err
	}

	var dek *memguard.Enclave
	switch dc.DEKInfos.Algorithm {
	case "AES":
		dek, err = CreateAESKey(32)
		if err != nil {
			return err
		}
	case "ECC":
		dek, err = CreateECCKey(publicDB)
		if err != nil {
			return err
		}
	}

	encodedDEK, err := AesEncryption(dek, kek)
	if err != nil {
		return err
	}

	// store DEK in right DB
	dekDB.SetKey(keySet.ID, relationBlind, "v1", dc.DEKInfos.Scope, dc.DEKInfos.InnerScope, encodedDEK, nil)

	return nil
}

// register new DEK under a KEK => DEK-KEK-reference stored
func (dc *DEKCombKEK) RegisterNewDEK(dekReg internAPIUtils.DEKRegister, dp wrapperUtils.DBPool, publicDB wrapperUtils.MySQLWrapper) error {
	body := internAPIUtils.NewStdRequest("createDEKRefs", dekReg)
	serialized, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bytes.NewBuffer(serialized))
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	handlers := internAPIUtils.NewHandlerMap("key", "keyMetadata")
	parts, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
	if err != nil {
		return err
	}

	kek, ok := parts["key"].(*memguard.Enclave)
	if !ok {
		return errors.New("unexpcted type of KEK while registering DEK")
	}

	// DEKIdentifier
	dekMetaRaw, ok := parts["keyMetadata"].([]byte)
	if !ok {
		return errors.New("unexpected type of metadata while registering DEK")
	}
	var dekMeta internAPIUtils.DEKIdentifier
	err = json.Unmarshal(dekMetaRaw, &dekMeta)
	if err != nil {
		return err
	}

	dbWrapper, err := dp.NewSQLWrapper(dc.DEKInfos.DB)
	if err != nil {
		return err
	}
	if dbWrapper.DB == nil {
		dbWrapper.Connect(context.Background(), 2)
	}

	dc.setNewDEKFromSet(dekMeta, dbWrapper, kek, publicDB)

	return nil
}

// retrieve decrypted DEK to handle Data
func (dc *DEKCombKEK) GetDEK(innerScope, userRef string, dp wrapperUtils.DBPool) (*memguard.Enclave, error) {
	userBlind, err := CreateUserBlind(serviceMasterSalt, dc.Scope, userRef, "KEK")
	if err != nil {
		return nil, err
	}

	getter := internAPIUtils.NewDEKGetter(internAPIUtils.NewKEKBlindedID(dc.KEKInfos.DB, userBlind), dc.DEKInfos.InnerScope)
	serialized, err := json.Marshal(internAPIUtils.NewStdRequest("decryptDEK", getter))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/getDEK", dc.KEKInfos.Manager), bytes.NewBuffer(serialized))
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	handlers := internAPIUtils.NewHandlerMap("key", "keyMetadata")
	data, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
	if err != nil {
		return nil, err
	}

	// DEKIdentifier
	dekMetaRaw, ok := data["keyMetadata"].([]byte)
	if !ok {
		return nil, errors.New("unexpected type of metadata while retrieving DEK")
	}
	var dekMeta internAPIUtils.DEKIdentifier
	err = json.Unmarshal(dekMetaRaw, &dekMeta)
	if err != nil {
		return nil, err
	}
	dbWrapper, err := dp.NewSQLWrapper(dc.DEKInfos.DB)
	if err != nil {
		return nil, err
	}
	if dbWrapper.DB == nil {
		dbWrapper.Connect(context.Background(), 2)
	}
	dek, err := dbWrapper.GetKey(dekMeta.KEKBlind, "relation")
	if err != nil {
		return nil, err
	}

	kek, ok := data["key"].(*memguard.Enclave)
	if !ok {
		return nil, errors.New("unexpected type of KEK while retrieving DEK")
	}

	decryptedDEK, err := AesDecryption(dek.([]byte), kek)
	if err != nil {
		return nil, err
	}

	return decryptedDEK, nil
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

func CreateECCKey(publicDBStore wrapperUtils.MySQLWrapper) (*memguard.Enclave, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	publicDBStore.SetData("INSERT INTO kstore (id, k_val, created_at, last_swap) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON DUPLICATE KEY UPDATE k_val = VALUES(k_val), created_at = CURRENT_TIMESTAMP, last_swap = CURRENT_TIMESTAMP", []any{"1", hex.EncodeToString(public)}, nil)

	return memguard.NewEnclave(private), nil
}
