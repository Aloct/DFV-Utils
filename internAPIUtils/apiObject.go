package internAPIUtils

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func (s *API) Start() {
    fmt.Println("Starting API on port %s", s.Port)
    err := http.ListenAndServe(":"+s.Port, s.Router)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        log.Fatalf("Server failed to start: %v", err)
    }
}