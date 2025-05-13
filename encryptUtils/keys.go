package encryptUtils

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	coreutils "github.com/Aloct/DFV-Utils/coreUtils"
	wrapperUtils "github.com/Aloct/DFV-Utils/dataHandling/storageWrapper"
	proto "github.com/Aloct/DFV-Utils/internAPIUtils/proto"
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
	// InnerScope  string
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
		// InnerScope:  innerScope,
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
func (dc *DEKCombKEK) RegisterNewKEK(pool wrapperUtils.DBPool, publicDB wrapperUtils.MySQLWrapper, grpcClient proto.KeyManagerClient, defaultDEKs ...*proto.DEKDefaultRegistration) error {
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

	stream, err := grpcClient.RegisterKEK(context.Background())
	if err != nil {
		return err
	}

	err = stream.Send(&proto.KEKAndDefaultDEKs{
		Data: &proto.KEKAndDefaultDEKs_Kek{
			Kek: &proto.KEKRegistration{
				Scope: dc.Scope,
				IdParams: &proto.DEKGetter{
					InnerScope: "",
					KekDb: dc.KEKInfos.DB,
					UserBlind: userBlind,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	resp, err := stream.Recv()
	if err == io.EOF {
		return errors.New("stream closed unexpectedly by server")
	}
	statusResp, ok := resp.RegisterResult.(*proto.RegisterResponse_Status)
	if !ok || statusResp.Status.StatusCode != 7001 {
		return errors.New("failed to register KEK")
	}

	// bodyMultiData := &bytes.Buffer{}
	// w := multipart.NewWriter(bodyMultiData)

	// subRequests := make([]internAPIUtils.StdRequest, len(defaultDEKs))
	// for i := range defaultDEKs {
	// 	subRequests = append(subRequests, internAPIUtils.NewStdRequest("createDEKRefs", defaultDEKs[i]))
	// }
	// w, err = internAPIUtils.SetSubRequestsForMultipartReq(w, subRequests)
	// if err != nil {
	// 	return err
	// }

	// kekRegister := internAPIUtils.NewKEKRegister(dc.KEKInfos.DB, dc.Scope, "", userBlind)
	// w, err = internAPIUtils.SetKeyMetaForMultipartReq(w, kek, kekRegister)
	// if err != nil {
	// 	return err
	// }

	// err = w.Close()
	// if err != nil {
	// 	return err
	// }

	// req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bodyMultiData)
	// if err != nil {
	// 	return err
	// }

	// req.Header.Add("Content-Type", "application/json")

	// res, err := client.Do(req)
	// if err != nil {
	// 	time.Sleep(10 * time.Second)
	// 	req, err = http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bodyMultiData)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	req.Header.Add("Content-Type", "application/json")
	// 	res, err = client.Do(req)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	if defaultDEKs != nil {
		// handlers := internAPIUtils.NewHandlerMap("subRequests")
		// resData, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
		// if err != nil {
		// 	return err
		// }

		// subResponses, ok := resData["subRequests"].(internAPIUtils.GroupResponse)
		// if !ok {
		// 	return errors.New("invalid response format")
		// }

		for _, dekObj := range defaultDEKs {
			err := stream.Send(&proto.KEKAndDefaultDEKs{
				Data: &proto.KEKAndDefaultDEKs_Dek{
					Dek: &proto.DEKDefaultRegistration{
						InnerScope: dekObj.InnerScope,
						Scope: dc.DEKInfos.Scope,
					},
				},
			})
			if err != nil {
				return err
			}
		}

		// for _, stdRespRaw := range subResponses.Subs {
		// 	var stdResp internAPIUtils.StdResponse
		// 	err := json.Unmarshal(stdRespRaw.([]byte), &stdResp)
		// 	if err != nil {
		// 		return err
		// 	}

		// 	dekObj, ok := stdResp.Data.(internAPIUtils.DEKIdentifier)
		// 	if !ok {
		// 		return errors.New("invalid DEK object format")
		// 	}

		// 	dc.setNewDEKFromSet(dekObj, dekDB, kek, publicDB)
		// }
	}

	err = stream.CloseSend()
	if err != nil {
		return err
	}

	if defaultDEKs != nil {
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

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return errors.New(err.Error())
			}

			result, ok := resp.RegisterResult.(*proto.RegisterResponse_DekResult)
			if !ok {
				return errors.New("unexpcted return for default DEK registration")
			}

			err = dc.setNewDEKFromSet(result.DekResult, dekDB, kek, publicDB)
			if err != nil {
				return err
			}
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

func (dc *DEKCombKEK) setNewDEKFromSet(keyMetadata *proto.DEKBlindResult, dekDB *wrapperUtils.MySQLWrapper, kek *memguard.Enclave, publicDB wrapperUtils.MySQLWrapper) error {
	dekKEKBlind := keyMetadata.KekBlind

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
	dekDB.SetKey(keyMetadata.Id, relationBlind, "v1", dc.DEKInfos.Scope, keyMetadata.InnerScope, encodedDEK, nil)

	return nil
}

// register new DEK under a KEK => DEK-KEK-reference stored
func (dc *DEKCombKEK) RegisterNewDEK(userRef, innerScope string, dp wrapperUtils.DBPool, publicDB wrapperUtils.MySQLWrapper, grpcClient proto.KeyManagerClient) error {
	// body := internAPIUtils.NewStdRequest("createDEKRefs", dekReg)
	// serialized, err := json.Marshal(body)
	// if err != nil {
	// 	return err
	// }

	// req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/registerKEK", dc.KEKInfos.Manager), bytes.NewBuffer(serialized))
	// if err != nil {
	// 	return err
	// }

	// res, err := client.Do(req)
	// if err != nil {
	// 	return err
	// }

	// handlers := internAPIUtils.NewHandlerMap("key", "keyMetadata")
	// parts, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
	// if err != nil {
	// 	return err
	// }

	// kek, ok := parts["key"].(*memguard.Enclave)
	// if !ok {
	// 	return errors.New("unexpcted type of KEK while registering DEK")
	// }

	// // DEKIdentifier
	// dekMetaRaw, ok := parts["keyMetadata"].([]byte)
	// if !ok {
	// 	return errors.New("unexpected type of metadata while registering DEK")
	// }
	// var dekMeta internAPIUtils.DEKIdentifier
	// err = json.Unmarshal(dekMetaRaw, &dekMeta)
	// if err != nil {
	// 	return err
	// }
	userBlind, err := CreateUserBlind(serviceMasterSalt, dc.Scope, userRef, "KEK")
	if err != nil {
		return err
	}

	dekMeta, err := grpcClient.RegisterDEK(context.Background(), &proto.DEKRegistration{
		KekId: &proto.KEKGetter{
			KekDb: dc.KEKInfos.DB,
			UserBlind: userBlind,
		},
		DekId: &proto.DEKDefaultRegistration{
			InnerScope: innerScope,
			Scope: dc.Scope,
		},
	})
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

	dc.setNewDEKFromSet(dekMeta.DekId, dbWrapper, memguard.NewEnclave(dekMeta.Kek), publicDB)

	return nil
}

// retrieve decrypted DEK to handle Data
func (dc *DEKCombKEK) GetDEK(innerScope, userRef string, dp wrapperUtils.DBPool, grpcClient proto.KeyManagerClient) (*memguard.Enclave, error) {
	userBlind, err := CreateUserBlind(serviceMasterSalt, dc.Scope, userRef, "KEK")
	if err != nil {
		return nil, err
	}

	// getter := internAPIUtils.NewDEKGetter(internAPIUtils.NewKEKBlindedID(dc.KEKInfos.DB, userBlind), dc.DEKInfos.InnerScope)
	// serialized, err := json.Marshal(internAPIUtils.NewStdRequest("decryptDEK", getter))
	// if err != nil {
	// 	return nil, err
	// }

	// req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/getDEK", dc.KEKInfos.Manager), bytes.NewBuffer(serialized))
	// if err != nil {
	// 	return nil, err
	// }
	// res, err := client.Do(req)
	// if err != nil {
	// 	return nil, err
	// }
	// defer res.Body.Close()

	// handlers := internAPIUtils.NewHandlerMap("key", "keyMetadata")
	// data, err := internAPIUtils.ProcessMultipartResponse(res, handlers)
	// if err != nil {
	// 	return nil, err
	// }

	// // DEKIdentifier
	// dekMetaRaw, ok := data["keyMetadata"].([]byte)
	// if !ok {
	// 	return nil, errors.New("unexpected type of metadata while retrieving DEK")
	// }
	// var dekMeta internAPIUtils.DEKIdentifier
	// err = json.Unmarshal(dekMetaRaw, &dekMeta)
	// if err != nil {
	// 	return nil, err
	// }
	res, err := grpcClient.DecryptKEKAndGetReference(context.Background(), &proto.DEKGetter{
		InnerScope: innerScope,
		KekDb: dc.KEKInfos.DB,
		UserBlind: userBlind,
	})
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
	dek, err := dbWrapper.GetKey(res.KekBlind, "relation")
	if err != nil {
		return nil, err
	}

	// kek, ok := data["key"].(*memguard.Enclave)
	// if !ok {
	// 	return nil, errors.New("unexpected type of KEK while retrieving DEK")
	// }

	decryptedDEK, err := AesDecryption(dek.([]byte), memguard.NewEnclave(res.Kek))
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
