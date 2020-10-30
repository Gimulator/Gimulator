package storage

import (
	"database/sql"
	"os"

	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/Gimulator/types"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/golang/protobuf/ptypes/timestamp"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var memoryPath string = ":memory:"

type Sqlite struct {
	*sql.DB
	log *logrus.Entry
}

func NewSqlite(path string, config *config.Config) (*Sqlite, error) {
	sqlite := &Sqlite{
		log: logrus.WithField("component", "sqlite"),
	}

	if path == "" {
		path = memoryPath
	}

	if err := sqlite.prepare(path, config); err != nil {
		return nil, err
	}

	return sqlite, nil
}

/////////////////////////////////////////////
/////////////////////////////////// Setup ///
/////////////////////////////////////////////

func (s *Sqlite) prepare(path string, config *config.Config) error {
	s.log.Info("starting to setup sqlite")

	s.log.WithField("path", path).Info("starting to create sqlite file")
	if path != memoryPath {
		file, err := os.Create(path)
		if err != nil {
			s.log.WithField("path", path).WithError(err).Error("could not create sqlite file")
			return err
		}
		defer file.Close()
	}

	s.log.Info("starting to open sqlite db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		s.log.WithError(err).Error("could not open sqlite db")
		return err
	}
	s.DB = db

	s.log.Info("starting to create credential table")
	if err := s.createUserTable(); err != nil {
		s.log.WithError(err).Error("could not create credential table")
		return err
	}

	s.log.Info("starting to create role table")
	if err := s.createRoleTable(); err != nil {
		s.log.WithError(err).Error("could not create role table")
		return err
	}

	s.log.Info("starting to create message table")
	if err := s.createMessageTable(); err != nil {
		s.log.WithError(err).Error("could not create message table")
		return err
	}

	s.log.Info("starting to fill credential table")
	if err := s.fillUserTable(config); err != nil {
		s.log.WithError(err).Error("could not fill credential table")
		return err
	}

	s.log.Info("starting to fill role table")
	if err := s.fillRoleTable(config); err != nil {
		s.log.WithError(err).Error("could not fill role table")
		return err
	}

	return nil
}

func (s *Sqlite) createUserTable() error {
	query := `
	CREATE TABLE user (
		"id" TEXT,
		"role" TEXT,
		"token" TEXT,
		"status" TEXT,
		PRIMARY KEY (id, token)
	);`

	stmt, err := s.Prepare(query)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) createRoleTable() error {
	query := `
	CREATE TABLE role (
		"role" TEXT,
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"method" TEXT
	);`

	stmt, err := s.Prepare(query)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) createMessageTable() error {
	query := `
	CREATE TABLE message (
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"content" TEXT,
		"owner" TEXT,
		"creationtimeseconds" INTEGER,
		"creationtimenanos" INTEGER,
		PRIMARY KEY (type, name, namespace)
	);`

	stmt, err := s.Prepare(query)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) fillUserTable(config *config.Config) error {
	for _, cred := range config.Credentials {
		query := `INSERT INTO user VALUES (?, ?, ?, ?)`

		stmt, err := s.Prepare(query)
		if err != nil {
			return err
		}

		if _, err = stmt.Exec(cred.ID, cred.Role, cred.Token, types.StatusUnknown); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sqlite) fillRoleTable(config *config.Config) error {
	for _, rule := range config.Roles.Director {
		for _, method := range rule.Methods {
			insertSQL := `INSERT INTO role VALUES (?, ?, ?, ?, ?)`

			stmt, err := s.Prepare(insertSQL)
			if err != nil {
				return err
			}

			if _, err = stmt.Exec(types.DirectorRole, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
				return err
			}
		}
	}

	for role, rules := range config.Roles.Actors {
		for _, rule := range rules {
			for _, method := range rule.Methods {
				insertSQL := `INSERT INTO role VALUES (?, ?, ?, ?, ?)`

				stmt, err := s.Prepare(insertSQL)
				if err != nil {
					return err
				}

				if _, err = stmt.Exec(role, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

/////////////////////////////////////////////
////////////////////////// MessageStorage ///
/////////////////////////////////////////////

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

func (s *Sqlite) put(message *api.Message) error {
	query := `INSERT OR REPLACE INTO message VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, err := s.Prepare(query)
	if err != nil {
		return status.Errorf(codes.Internal, "could not prepare statement for database: %v", err)
	}

	seconds := message.Meta.CreationTime.Seconds
	nanos := message.Meta.CreationTime.Nanos

	_, err = stmt.Exec(message.Key.Type, message.Key.Name, message.Key.Namespace, message.Content, message.Meta.Owner, seconds, nanos)
	if err != nil {
		return status.Errorf(codes.Internal, "could not execute statement on database: %v", err)
	}

	return nil
}

func (s *Sqlite) get(key *api.Key) (*api.Message, error) {
	query := `SELECT * FROM message WHERE type = ? AND name = ? AND namespace = ?`

	rows, err := s.Query(query, key.Type, key.Name, key.Namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	flag := false
	var seconds, nanos int
	messageR := api.Message{
		Content: "",
		Key:     &api.Key{},
		Meta:    &api.Meta{},
	}

	for rows.Next() {
		flag = true
		if err := rows.Scan(&messageR.Key.Type, &messageR.Key.Name, &messageR.Key.Namespace, &messageR.Content, &messageR.Meta.Owner, &seconds, &nanos); err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
		messageR.Meta.CreationTime = s.marshalTimestamp(seconds, nanos)
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "message with key=%v does not exist", key)
	}

	return &messageR, nil
}

func (s *Sqlite) getAll(key *api.Key) ([]*api.Message, error) {
	query := `SELECT * FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`

	t, n, ns := s.makeKey(key)

	rows, err := s.Query(query, t, n, ns)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

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
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}

		messageR.Meta.CreationTime = s.marshalTimestamp(seconds, nanos)
		messages = append(messages, &messageR)
	}

	return messages, nil
}

func (s *Sqlite) delete(key *api.Key) error {
	query := `DELETE FROM message WHERE type = ? AND name = ? AND namespace = ? `

	stmt, err := s.Prepare(query)
	if err != nil {
		return status.Errorf(codes.Internal, "could not prepare statement for database: %v", err)
	}

	if _, err = stmt.Exec(key.Type, key.Name, key.Namespace); err != nil {
		return status.Errorf(codes.Internal, "could not execute statement on database: %v", err)
	}

	return nil
}

func (s *Sqlite) deleteAll(key *api.Key) error {
	deleteStatement := `DELETE FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`

	stmt, err := s.Prepare(deleteStatement)
	if err != nil {
		return status.Errorf(codes.Internal, "could not prepare statement for database: %v", err)
	}

	t, n, ns := s.makeKey(key)

	if _, err = stmt.Exec(t, n, ns); err != nil {
		return status.Errorf(codes.Internal, "could not execute statement on database: %v", err)
	}

	return nil
}

/////////////////////////////////////////////
///////////////////////////// UserStorage ///
/////////////////////////////////////////////

func (s *Sqlite) GetUserWithToken(token string) (*User, error) {
	return s.getUserWithToken(token)
}

func (s *Sqlite) GetUserWithID(id string) (*User, error) {
	return s.getUserWithID(id)
}

func (s *Sqlite) UpdateUserStatus(id string, status types.Status) error {
	return s.updateUserStatus(id, status)
}

func (s *Sqlite) getUserWithToken(token string) (*User, error) {
	selectStatement := `SELECT * FROM user WHERE token = ?`

	rows, err := s.Query(selectStatement, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	user := &User{}
	flag := false

	for rows.Next() {
		flag = true

		err = rows.Scan(user.ID, user.Role, user.Token, user.Status)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "")
	}

	return user, nil
}

func (s *Sqlite) getUserWithID(id string) (*User, error) {
	selectStatement := `SELECT * FROM user WHERE id = ?`

	rows, err := s.Query(selectStatement, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	user := &User{}
	flag := false

	for rows.Next() {
		flag = true

		err = rows.Scan(user.ID, user.Role, user.Token, user.Status)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "")
	}

	return user, nil
}

func (s *Sqlite) updateUserStatus(id string, st types.Status) error {
	query := `UPDATE user SET status = ? WHERE id = ?`

	stmt, err := s.Prepare(query)
	if err != nil {
		return status.Errorf(codes.Internal, "could not prepare statement for database: %v", err)
	}

	if _, err = stmt.Exec(st, id); err != nil {
		return status.Errorf(codes.Internal, "could not execute statement on database: %v", err)
	}

	return nil
}

/////////////////////////////////////////////
///////////////////////////// RoleStorage ///
/////////////////////////////////////////////

func (s *Sqlite) GetRules(role string, method types.Method) ([]*api.Key, error) {
	return s.getRules(role, method)
}

func (s *Sqlite) getRules(role string, method types.Method) ([]*api.Key, error) {
	selectStatement := `SELECT type, name, namespace FROM role WHERE role = ? AND method = ?`

	rows, err := s.Query(selectStatement, role, method)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	keys := make([]*api.Key, 0)

	for rows.Next() {
		key := &api.Key{}

		err = rows.Scan(&key.Type, &key.Name, &key.Namespace)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}

		keys = append(keys, key)
	}
	return keys, nil
}

/////////////////////////////////////////////
////////////////////////////////// Helper ///
/////////////////////////////////////////////

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
