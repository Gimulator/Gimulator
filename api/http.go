package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gitlab.com/Syfract/Xerac/gimulator/auth"
	"gitlab.com/Syfract/Xerac/gimulator/object"
	"gitlab.com/Syfract/Xerac/gimulator/simulator"
)

const (
	readBufSize  = 1024
	writeBufSize = 1024
	tokenKey     = "token"
)

type Manager struct {
	sync.Mutex
	router    *mux.Router
	simulator *simulator.Simulator
	auth      *auth.Auth
	log       *logrus.Entry
	clients   map[string]*client
}

func NewManager(sim *simulator.Simulator, auth *auth.Auth) *Manager {
	m := &Manager{
		Mutex:     sync.Mutex{},
		simulator: sim,
		auth:      auth,
		log:       logrus.WithField("Entity", "http"),
		clients:   make(map[string]*client),
	}
	m.route()

	return m
}

func (m *Manager) ListenAndServe(bind string) error {
	m.log.Info("Start to Listen and Serve")
	if m.router == nil {
		m.route()
	}
	return http.ListenAndServe(bind, m.router)
}

func (m *Manager) route() {
	m.log.Info("Start to set router")
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
	log := m.log.WithField("Request", "get")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Authorize(cli.cred.Role, auth.Get, obj.Key)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Unauthorize action")
		http.Error(w, msg, status)
		return
	}

	result, err := m.simulator.Get(obj.Key)
	if err != nil {
		log.WithError(err).Debug("Error in get result from simulator")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.WithError(err).Debug("Can not encode result of simulator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (m *Manager) handleFind(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "find")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Authorize(cli.cred.Role, auth.Get, obj.Key)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Unauthorize action")
		http.Error(w, msg, status)
		return
	}

	objectList, err := m.simulator.Find(obj.Key)
	if err != nil {
		log.WithError(err).Debug("Error in get result from simulator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(objectList); err != nil {
		log.WithError(err).Debug("Can not encode result of simulator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (m *Manager) handleSet(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "set")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Authorize(cli.cred.Role, auth.Get, obj.Key)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Unauthorize action")
		http.Error(w, msg, status)
		return
	}

	if err := m.simulator.Set(obj); err != nil {
		log.WithError(err).Debug("Error in get result from simulator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (m *Manager) handleDelete(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "delete")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Authorize(cli.cred.Role, auth.Get, obj.Key)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Unauthorize action")
		http.Error(w, msg, status)
		return
	}

	if err := m.simulator.Delete(obj.Key); err != nil {
		log.WithError(err).Debug("Error in get result from simulator")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (m *Manager) handleWatch(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "watch")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Authorize(cli.cred.Role, auth.Get, obj.Key)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Unauthorize action")
		http.Error(w, msg, status)
		return
	}

	if err := m.simulator.Watch(obj.Key, cli.GetChan()); err != nil {
		log.WithError(err).Debug("Error in get result from simulator")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (m *Manager) handleRegister(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "register")
	log.Info("Start to handle request")

	var cred *Credential
	if status, msg := decodeJSONBody(w, r, &cred); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	if status, msg := m.auth.Authenticate(cred.Role, cred.Password); status != http.StatusOK {
		log.WithField("status", status).WithField("message", msg).Debug("Fail to authenticate")
		http.Error(w, msg, status)
		return
	}

	cli, status, msg := m.getClientWithUsername(cred.Username)
	if status != http.StatusOK {
		if cli, status, msg = m.registerNewClient(cred); status != http.StatusOK {
			http.Error(w, msg, status)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: cli.GetToken(),
	})
	w.WriteHeader(http.StatusAccepted)
}

func (m *Manager) handleSocket(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "socket")
	log.Info("Start to handle request")

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Invalid token")
		http.Error(w, msg, status)
		return
	}

	log.Info("Start to upgrade connection")
	conn, err := websocket.Upgrade(w, r, w.Header(), readBufSize, writeBufSize)
	if err != nil {
		log.WithError(err).Debug("Can not upgrade connection")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cli.Reconcile(conn)
}

func (m *Manager) validateToken(r *http.Request) (*client, int, string) {
	cookie, err := r.Cookie(tokenKey)
	if err != nil {
		return nil, http.StatusUnauthorized, fmt.Sprintf("Invalid token")
	}
	token := cookie.Value

	cli, status, msg := m.getClientWithToken(token)
	if status != http.StatusOK {
		return nil, http.StatusUnauthorized, msg
	}

	if cli.token != token {
		return nil, http.StatusUnauthorized, "Invalid token"
	}
	return cli, http.StatusOK, ""
}

func (m *Manager) getClientWithToken(token string) (*client, int, string) {
	m.Lock()
	defer m.Unlock()
	m.log.Info("Start to get a client")

	if cli, exists := m.clients[token]; exists {
		return cli, http.StatusOK, ""
	}
	return nil, http.StatusNotFound, "There is no client with the specified cookie"
}

func (m *Manager) getClientWithUsername(username string) (*client, int, string) {
	for _, c := range m.clients {
		if c.cred.Username == username {
			return c, http.StatusOK, ""
		}
	}
	return nil, http.StatusNotFound, "User not found"
}

func (m *Manager) registerNewClient(cred *Credential) (*client, int, string) {
	m.Lock()
	defer m.Unlock()
	m.log.Info("Start to create a new client")

	token, status := newCookie()
	if status != http.StatusOK {
		return nil, status, "Can not generate new cookie"
	}

	client := NewClient(cred, token)
	m.clients[token] = client

	return client, http.StatusOK, ""
}
