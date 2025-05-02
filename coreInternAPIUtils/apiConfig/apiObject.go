package apiConfig

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/awnumar/memguard"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type API struct {
	Router *mux.Router
    Port string
    Protected bool
}

func NewAPI(port string, protected bool) *API {
    return &API{
        Router: mux.NewRouter(),
        Port: port,
        Protected: protected, 
    }
}

func (s *API) AddRoute(path string, method string, handlerFunc http.Handler, givenMiddlewares ...func(http.Handler) http.Handler) {
    middlewares := append([]func(http.Handler) http.Handler{Auth}, givenMiddlewares...)
    var handler http.Handler = handlerFunc
    for _, middleware := range middlewares {
        handler = middleware(handler)
    }

    s.Router.Handle(path, handler).Methods(method)
}

func (s *API) Start() error {
    if (s.Protected) {
        memguard.CatchInterrupt()
        defer memguard.Purge()
        
        err := godotenv.Load(".env")
        if err != nil {
            return err
        }

        var cert tls.Certificate

        if (os.Getenv("TLS_cert") != "" || os.Getenv("TLS_prv") != "") {
            certFile := memguard.NewBufferFromBytes([]byte(os.Getenv("TLS_cert")))
            prvKey := memguard.NewBufferFromBytes([]byte(os.Getenv("TLS_prv")))
    
            cert, err = tls.LoadX509KeyPair(string(certFile.Bytes()), string(prvKey.Bytes()))
            if err != nil {
                return err
            }

            certFile.Destroy()
            prvKey.Destroy()
        } else {
            return errors.New("no TLS certificate or/and TLS private key provided")
        }

        tlsConfig := &tls.Config{
            Certificates: []tls.Certificate{cert},
            MinVersion: tls.VersionTLS13,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            },
            PreferServerCipherSuites: true,
        }

        server := &http.Server{
            Addr: ":"+s.Port,
            Handler: s.Router,
            TLSConfig: tlsConfig,
        }

        fmt.Println(fmt.Printf("Starting TLS-secure API on port %s", s.Port))
        err = server.ListenAndServeTLS("", "")
        if err != nil {
            return err
        }

    } else {
        fmt.Println(fmt.Printf("Starting API on port %s", s.Port))
        err := http.ListenAndServe(":"+s.Port, s.Router)
        if err != nil {
            return err
        }
    }

    return nil
}