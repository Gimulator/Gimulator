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
	"github.com/sirupsen/logrus"
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
	m.handleRequest(w, r, auth.Get)
}

func (m *Manager) handleFind(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, auth.Find)
}

func (m *Manager) handleSet(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, auth.Set)
}

func (m *Manager) handleDelete(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, auth.Delete)
}

func (m *Manager) handleWatch(w http.ResponseWriter, r *http.Request) {
	m.handleRequest(w, r, auth.Watch)
}

func (m *Manager) handleRequest(w http.ResponseWriter, r *http.Request, method auth.Method) {
	log := m.log.WithField("method", method)

	cli, status, msg := m.validateToken(r)
	if status != http.StatusOK {
		log.WithFields(logrus.Fields{
			"status":  status,
			"message": msg,
		}).Error("could not validate token")
		http.Error(w, msg, status)
		return
	}

	var obj *object.Object
	if status, msg = decodeJSONBody(w, r, &obj); status != http.StatusOK {
		log.WithFields(logrus.Fields{
			"status":    status,
			"message":   msg,
			"client-id": cli.id,
		}).Error("could not decode the body of request")
		http.Error(w, msg, status)
		return
	}

	status, msg = m.auth.Auth(cli.id, method, obj)
	if status != http.StatusOK {
		log.WithFields(logrus.Fields{
			"status":       status,
			"message":      msg,
			"asked-object": obj.String(),
			"client-id":    cli.id,
		}).Error("could not auth the request")
		http.Error(w, msg, status)
		return
	}

	if err := m.respond(w, obj, cli, method); err != nil {
		log.WithFields(logrus.Fields{
			"asked-object": obj.String(),
			"client-id":    cli.id,
		}).WithError(err).Error("could not respond to the request")
	}
}

func (m *Manager) respond(w http.ResponseWriter, obj *object.Object, cli *client, method auth.Method) error {
	var result interface{} = nil
	var err error = nil

	switch method {
	case auth.Get:
		result, err = m.simulator.Get(obj.Key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case auth.Find:
		result, err = m.simulator.Find(obj.Key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if err = json.NewEncoder(w).Encode(result); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case auth.Delete:
		err = m.simulator.Delete(obj.Key)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	case auth.Set:
		err = m.simulator.Set(obj)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	case auth.Watch:
		err = m.simulator.Watch(obj.Key, cli.ch)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	default:
		err = fmt.Errorf("invalid method for simulation")
		m.log.Fatal("invalid method for simulation")
	}
	return err
}

func (m *Manager) handleRegister(w http.ResponseWriter, r *http.Request) {
	log := m.log.WithField("Request", "register")
	log.Info("Start to handle request")

	cred := struct {
		ID string
	}{}
	if status, msg := decodeJSONBody(w, r, &cred); status != http.StatusOK {
		log.WithField("Status", status).WithField("message", msg).Debug("Can not decode json body")
		http.Error(w, msg, status)
		return
	}

	if status, msg := m.auth.Register(cred.ID); status != http.StatusOK {
		log.WithField("status", status).WithField("message", msg).Debug("Fail to authenticate")
		http.Error(w, msg, status)
		return
	}

	cli, status, msg := m.getClientWithID(cred.ID)
	if status != http.StatusOK {
		if cli, status, msg = m.registerNewClient(cred.ID); status != http.StatusOK {
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

func (m *Manager) getClientWithID(id string) (*client, int, string) {
	for _, c := range m.clients {
		if c.id == id {
			return c, http.StatusOK, ""
		}
	}
	return nil, http.StatusNotFound, "User not found"
}

func (m *Manager) registerNewClient(id string) (*client, int, string) {
	m.Lock()
	defer m.Unlock()
	m.log.Info("Start to create a new client")

	token, status := newCookie()
	if status != http.StatusOK {
		return nil, status, "Can not generate new cookie"
	}

	client := NewClient(id, token)
	m.clients[token] = client

	return client, http.StatusOK, ""
}
