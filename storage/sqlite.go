package storage

import (
	"database/sql"
	"os"

	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var memoryPath string = ":memory:"

type Sqlite struct {
	*sql.DB
}

func NewSqlite(path string, config *config.Config) (*Sqlite, error) {
	sqlite := &Sqlite{}

	if path == "" {
		path = memoryPath
	}

	if err := sqlite.prepare(path); err != nil {
		return nil, err
	}

	if err := sqlite.setupTables(config); err != nil {
		return nil, err
	}

	return sqlite, nil
}

func (s *Sqlite) Put(message *api.Message) error {
	return s.put(message)
}

func (s *Sqlite) Get(key *api.Key) (*api.Message, error) {
	return s.get(key)
}

func (s *Sqlite) GetAll(key *api.Key) ([]*api.Message, error) {
	return s.getAll(key)
}

func (s *Sqlite) Delete(key *api.Key) error {
	return s.delete(key)
}

func (s *Sqlite) DeleteAll(key *api.Key) error {
	return s.deleteAll(key)
}

func (s *Sqlite) GetCredWithToken(token string) (string, string, error) {
	return s.getCredWithToken(token)
}

func (s *Sqlite) GetRules(role string, method types.Method) ([]*api.Key, error) {
	return s.getRules(role, method)
}

func (s *Sqlite) getCredWithToken(token string) (string, string, error) {
	selectStatement := `SELECT id, role FROM credentials WHERE token = ?`

	rows, err := s.Query(selectStatement, token)
	defer rows.Close()
	if err != nil {
		return "", "", err
	}

	var id, role string
	flag := false

	for rows.Next() {
		flag = true

		err = rows.Scan(&id, &role)
		if err != nil {
			return "", "", err
		}
	}

	if !flag {
		return "", "", status.Errorf(codes.NotFound, "")
	}

	return id, role, nil
}

func (s *Sqlite) getRules(role string, method types.Method) ([]*api.Key, error) {
	selectStatement := `SELECT type, name, namespace FROM roles WHERE role = ? AND method = ?`

	rows, err := s.Query(selectStatement, role, method)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	keys := make([]*api.Key, 0)

	for rows.Next() {
		key := &api.Key{}

		err = rows.Scan(&key.Type, &key.Name, &key.Namespace)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}
	return keys, nil
}

func (s *Sqlite) put(message *api.Message) error {
	insertSQL := `INSERT INTO message VALUES (?, ?, ?, ?, ?, ?, ?)`

	statement, err := s.Prepare(insertSQL)
	if err != nil {
		return err
	}

	seconds := message.Meta.CreationTime.Seconds
	nanos := message.Meta.CreationTime.Nanos

	if _, err = statement.Exec(message.Key.Type, message.Key.Name, message.Key.Namespace, message.Content, message.Meta.Owner, seconds, nanos); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) get(key *api.Key) (*api.Message, error) {
	selectStatement := `SELECT * FROM message WHERE type = ? AND name = ? AND namespace = ?`

	rows, err := s.Query(selectStatement, key.Type, key.Name, key.Namespace)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	flag := false
	var seconds, nanos int
	messageR := api.Message{
		Content: "",
		Key:     &api.Key{},
		Meta:    &api.Meta{},
	}

	for rows.Next() {
		flag = true
		err = rows.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &seconds, &nanos)
		if err != nil {
			return nil, err
		}
		messageR.Meta.CreationTime = s.marshalTimestamp(seconds, nanos)
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "message with key=%v does not exist", key)
	}

	return &messageR, nil
}

func (s *Sqlite) getAll(key *api.Key) ([]*api.Message, error) {
	selectStatement := `SELECT * FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`

	t, n, ns := s.makeKey(key)

	rows, err := s.Query(selectStatement, t, n, ns)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var messages []*api.Message
	for rows.Next() {

		messageR := api.Message{
			Content: "",
			Key:     &api.Key{},
			Meta:    &api.Meta{},
		}
		var seconds, nanos int

		err = rows.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &seconds, &nanos)
		if err != nil {
			return nil, err
		}

		messageR.Meta.CreationTime = s.marshalTimestamp(seconds, nanos)
		messages = append(messages, &messageR)
	}

	return messages, nil
}

func (s *Sqlite) delete(key *api.Key) error {
	deleteStatement := `DELETE FROM message WHERE type = ? AND name = ? AND namespace = ? `

	statement, err := s.Prepare(deleteStatement)
	if err != nil {
		return err
	}

	if _, err = statement.Exec(key.Type, key.Name, key.Namespace); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) deleteAll(key *api.Key) error {
	deleteStatement := `DELETE FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`

	statement, err := s.Prepare(deleteStatement)
	if err != nil {
		return err
	}

	t, n, ns := s.makeKey(key)

	if _, err = statement.Exec(t, n, ns); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) prepare(path string) error {
	err := s.createDBFile(path)
	if err != nil {
		return err
	}

	err = s.open(path)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) createDBFile(path string) error {
	if path == memoryPath {
		return nil
	}

	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) open(path string) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	s.DB = db
	return nil
}

func (s *Sqlite) setupTables(config *config.Config) error {
	if err := s.createCredentialsTable(); err != nil {
		return err
	}

	if err := s.createRolesTable(); err != nil {
		return err
	}

	if err := s.createMessageTable(); err != nil {
		return err
	}

	if err := s.fillRolesTable(config); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) createCredentialsTable() error {
	createMessageTable := `CREATE TABLE roles (
		"id" TEXT,
		"role" TEXT,
		"token" TEXT,
		PRIMARY KEY (id, token)
	);`

	statement, err := s.Prepare(createMessageTable)
	if err != nil {
		return err
	}

	if _, err = statement.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) fillCredentialsTable(config *config.Config) error {
	for _, cred := range config.Credentials {
		insertSQL := `INSERT INTO credentials VALUES (?, ?, ?)`

		statement, err := s.Prepare(insertSQL)
		if err != nil {
			return err
		}

		if _, err = statement.Exec(cred.ID, cred.Role, cred.Token); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sqlite) createRolesTable() error {
	createMessageTable := `CREATE TABLE credentials (
		"role" TEXT,
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"method" TEXT,
	);`

	statement, err := s.Prepare(createMessageTable)
	if err != nil {
		return err
	}

	if _, err = statement.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) fillRolesTable(config *config.Config) error {
	for _, rule := range config.Roles.Director {
		for _, method := range rule.Methods {
			insertSQL := `INSERT INTO roles VALUES (?, ?, ?, ?, ?)`

			statement, err := s.Prepare(insertSQL)
			if err != nil {
				return err
			}

			if _, err = statement.Exec(types.DirectorRole, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Sqlite) createMessageTable() error {
	createMessageTable := `CREATE TABLE message (
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"content" TEXT,
		"owner" TEXT,
		"creationtimeseconds" INTEGER,
		"creationtimenanos" INTEGER,
		PRIMARY KEY (type, name, namespace)
	);`

	statement, err := s.Prepare(createMessageTable)
	if err != nil {
		return err
	}

	if _, err = statement.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) marshalTimestamp(seconds, nanos int) *timestamp.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: int64(seconds),
		Nanos:   int32(nanos),
	}
}

func (s *Sqlite) makeKey(key *api.Key) (string, string, string) {
	t := key.Type
	if t == "" {
		t = "%"
	}

	n := key.Name
	if n == "" {
		n = "%"
	}

	ns := key.Namespace
	if ns == "" {
		ns = "%"
	}

	return t, n, ns
}
