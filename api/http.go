package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/object"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
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
		log:       logrus.WithField("entity", "http"),
		clients:   make(map[string]*client),
	}
	m.route()

	return m
}

func (m *Manager) ListenAndServe(bind string) error {
	m.log.Info("starting to listen and serve")
	defer m.log.Info("end of listening and serving")

	if m.router == nil {
		m.route()
	}

	if err := http.ListenAndServe(bind, m.router); err != nil {
		m.log.WithError(err).Error("could not listen and serve")
		return err
	}
	return nil
}

func (m *Manager) route() {
	m.log.Info("starting to set router")

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
	m.handleRequest(w, r, object.MethodGet)
}

func (m *Manager) handleFind(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, object.MethodFind)
}

func (m *Manager) handleSet(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, object.MethodSet)
}

func (m *Manager) handleDelete(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, object.MethodDelete)
}

func (m *Manager) handleWatch(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, object.MethodWatch)
}

func (m *Manager) handleRequest(w http.ResponseWriter, r *http.Request, method object.Method) {
	log := m.log.WithField("method", method)

	cli, err := m.tokenToClient(r)
	if err != nil {
		log.WithError(err).Error("could not validate token")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	log = log.WithField("client-id", cli.id)

	var obj *object.Object
	if msg := decodeJSONBody(w, r, &obj); msg != "" {
		log.WithField("message", msg).Error("could not decode the body of request")
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	log = log.WithField("asked-object", obj.String())
	log.Debug("starting to handle request")

	if err := m.auth.Auth(cli.id, method, obj); err != nil {
		log.WithError(err).Error("could not auth the request")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := m.respond(w, obj, cli, method); err != nil {
		log.WithError(err).Error("could not respond correctly to the request")
	}
}

func (m *Manager) respond(w http.ResponseWriter, obj *object.Object, cli *client, method object.Method) error {
	var result interface{} = nil
	var err error = nil

	switch method {
	case object.MethodGet:
		result, err = m.simulator.Get(cli.id, obj.Key)
		if err != nil {
			http.Error(w, "could not find object by the given key", http.StatusUnprocessableEntity)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case object.MethodFind:
		result, err = m.simulator.Find(cli.id, obj.Key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case object.MethodDelete:
		err = m.simulator.Delete(cli.id, obj.Key)
		if err != nil {
			http.Error(w, "could not find object by the given key", http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case object.MethodSet:
		err = m.simulator.Set(cli.id, obj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	case object.MethodWatch:
		err = m.simulator.Watch(cli.id, obj.Key, cli.ch)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		err = fmt.Errorf("invalid method for simulation")
		m.log.Fatal("invalid method for simulation")
	}
	return err
}

func (m *Manager) handleRegister(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("method", "register")

	cred := struct{ ID string }{}
	if msg := decodeJSONBody(w, r, &cred); msg != "" {
		log.WithFields(logrus.Fields{
			"message": msg,
		}).Error("could not decode the body of request")
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	if err := m.auth.Register(cred.ID); err != nil {
		log.WithField("client-id", cred.ID).WithError(err).Error("could not auth the request")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	cli := m.getClientWithID(cred.ID)
	if cli == nil {
		cli = m.registerNewClient(cred.ID)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: cli.GetToken(),
	})
	w.WriteHeader(http.StatusOK)
}

const (
	readBufSize  = 1024
	writeBufSize = 1024
)

func (m *Manager) handleSocket(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("method", "socket")

	cli, err := m.tokenToClient(r)
	if err != nil {
		log.WithError(err).Error("could not validate token")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Upgrade(w, r, w.Header(), readBufSize, writeBufSize)
	if err != nil {
		log.WithFields(logrus.Fields{
			"client-id": cli.id,
		}).WithError(err).Error("could not upgrade connection to websocket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cli.Reconcile(conn)
}

const tokenKey = "token"

func (m *Manager) tokenToClient(r *http.Request) (*client, error) {
	m.Lock()
	defer m.Unlock()

	cookie, err := r.Cookie(tokenKey)
	if err != nil {
		return nil, fmt.Errorf("invalid token type")
	}
	token := cookie.Value

	cli, exists := m.clients[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}
	return cli, nil
}

func (m *Manager) getClientWithID(id string) *client {
	m.Lock()
	defer m.Unlock()

	for _, c := range m.clients {
		if c.id == id {
			return c
		}
	}
	return nil
}

func (m *Manager) registerNewClient(id string) *client {
	m.log.WithField("client-id", id).Info("starting to register new client")
	m.Lock()
	defer m.Unlock()

	token := uuid.NewV4().String()
	client := NewClient(id, token)

	m.clients[token] = client

	return client
}
