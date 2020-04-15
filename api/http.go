package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gitlab.com/Syfract/Xerac/gimulator/auth"
	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gitlab.com/Syfract/Xerac/gimulator/simulator"
)

const (
	readBufSize  = 1024
	writeBufSize = 1024
)

type Manager struct {
	sync.Mutex
	router    *mux.Router
	simulator *simulator.Simulator
	auth      *auth.Auth
}

func NewManager(sim *simulator.Simulator, auth *auth.Auth) *Manager {
	m := Manager{
		Mutex:     sync.Mutex{},
		simulator: sim,
		auth:      auth,
	}
	m.route()

	return &m
}

func (m *Manager) ListenAndServe(bind string) error {
	if m.router == nil {
		m.route()
	}
	return http.ListenAndServe(bind, m.router)
}

func (m *Manager) route() {
	r := mux.NewRouter()
	r.HandleFunc("/register", m.handleRegister).Methods("POST")
	r.HandleFunc("/get", m.handleGet).Methods("POST")
	r.HandleFunc("/find", m.handleFind).Methods("POST")
	r.HandleFunc("/set", m.handleSet).Methods("POST")
	r.HandleFunc("/delete", m.handleDelete).Methods("POST")
	r.HandleFunc("/watch", m.handleWatch).Methods("POST")
	r.HandleFunc("/socket", m.handleSocket).Methods("GET")
	m.router = r
}

func (m *Manager) handleGet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Get")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.HandleRequest(w, r, auth.Get, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	result, err := m.simulator.Get(obj.Key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (m *Manager) handleFind(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Find")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.HandleRequest(w, r, auth.Find, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	objectList, err := m.simulator.Find(obj.Key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(objectList); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (m *Manager) handleSet(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Set")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.HandleRequest(w, r, auth.Set, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	if err := m.simulator.Set(obj); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (m *Manager) handleDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Delete")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.HandleRequest(w, r, auth.Delete, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	if err := m.simulator.Delete(obj.Key); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (m *Manager) handleWatch(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Watch")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.HandleRequest(w, r, auth.Watch, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	m.simulator.Watch(obj.Key, cli.GetChan())
	w.WriteHeader(http.StatusOK)
}

func (m *Manager) handleRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Register")
	var (
		obj object.Object
		cli auth.Client
	)
	status, msg := m.auth.RegisterNewClient(w, r, &cli, &obj)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: cli.GetToken(),
	})
	w.WriteHeader(http.StatusOK)
}

func (m *Manager) handleSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------Socket")
	cookie, err := r.Cookie("token")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	cli, status, msg := m.auth.GetClientWithToken(token)
	if status != http.StatusAccepted {
		http.Error(w, msg, status)
		return
	}

	conn, err := websocket.Upgrade(w, r, w.Header(), readBufSize, writeBufSize)
	if err != nil {
		http.Error(w, "can not open socket", http.StatusInternalServerError)
		return
	}

	cli.SetConn(conn)
}
