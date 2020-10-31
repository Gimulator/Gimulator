package storage

import (
	"database/sql"
	"os"

	"github.com/Gimulator/Gimulator/config"
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
	if err := s.createRuleTable(); err != nil {
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
	if err := s.fillRuleTable(config); err != nil {
		s.log.WithError(err).Error("could not fill role table")
		return err
	}

	return nil
}

func (s *Sqlite) createUserTable() error {
	query := `
	CREATE TABLE user (
		"id" TEXT NOT NULL,
		"token" TEXT NOT NULL UNIQUE,
		"character" INTEGER NOT NULL,
		"role" TEXT NOT NULL,
		"status" INTEGER NOT NULL,
		"readiness" BOOLEAN NOT NULL,
		PRIMARY KEY (id)
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

func (s *Sqlite) createRuleTable() error {
	query := `
	CREATE TABLE rule (
		"character" INTEGER NOT NULL,
		"role" TEXT,
		"type" TEXT,
		"name" TEXT,
		"namespace" TEXT,
		"method" INTEGER NOT NULL
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
		"type" TEXT NOT NULL,
		"name" TEXT NOT NULL,
		"namespace" TEXT NOT NULL,
		"content" TEXT NOT NULL,
		"owner" TEXT NOT NULL,
		"creationtimeseconds" INTEGER NOT NULL,
		"creationtimenanos" INTEGER NOT NULL,
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
		query := `INSERT INTO user VALUES (?, ?, ?, ?, ?, ?)`

		stmt, err := s.Prepare(query)
		if err != nil {
			return err
		}

		if _, err = stmt.Exec(cred.ID, cred.Token, cred.Character, cred.Role, api.Status_unknown, false); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sqlite) fillRuleTable(config *config.Config) error {
	for _, rule := range config.Character.Director {
		for _, method := range rule.Methods {
			insertSQL := `INSERT INTO rule VALUES (?, ?, ?, ?, ?, ?)`

			stmt, err := s.Prepare(insertSQL)
			if err != nil {
				return err
			}

			if rule.Key == nil {
				if _, err = stmt.Exec(api.Character_director, nil, nil, nil, nil, method); err != nil {
					return err
				}
			} else {
				if _, err = stmt.Exec(api.Character_director, nil, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
					return err
				}
			}
		}
	}

	for _, rule := range config.Character.Operator {
		for _, method := range rule.Methods {
			insertSQL := `INSERT INTO rule VALUES (?, ?, ?, ?, ?, ?)`

			stmt, err := s.Prepare(insertSQL)
			if err != nil {
				return err
			}
			if rule.Key == nil {
				if _, err = stmt.Exec(api.Character_operator, nil, nil, nil, nil, method); err != nil {
					return err
				}
			} else {
				if _, err = stmt.Exec(api.Character_operator, nil, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
					return err
				}
			}
		}
	}

	for _, rule := range config.Character.Master {
		for _, method := range rule.Methods {
			insertSQL := `INSERT INTO rule VALUES (?, ?, ?, ?, ?, ?)`

			stmt, err := s.Prepare(insertSQL)
			if err != nil {
				return err
			}
			if rule.Key == nil {
				if _, err = stmt.Exec(api.Character_master, nil, nil, nil, nil, method); err != nil {
					return err
				}
			} else {
				if _, err = stmt.Exec(api.Character_master, nil, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
					return err
				}
			}
		}
	}

	for role, rules := range config.Character.Actors {
		for _, rule := range rules {
			for _, method := range rule.Methods {
				insertSQL := `INSERT INTO rule VALUES (?, ?, ?, ?, ?, ?)`

				stmt, err := s.Prepare(insertSQL)
				if err != nil {
					return err
				}

				if _, err = stmt.Exec(api.Character_actor, role, rule.Key.Type, rule.Key.Name, rule.Key.Namespace, method); err != nil {
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
	query := `DELETE FROM message WHERE type LIKE ? AND name LIKE ? AND namespace LIKE ?`

	stmt, err := s.Prepare(query)
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

func (s *Sqlite) GetUserWithToken(token string) (*api.User, error) {
	return s.getUserWithToken(token)
}

func (s *Sqlite) GetUserWithID(id string) (*api.User, error) {
	return s.getUserWithID(id)
}

func (s *Sqlite) GetUsers(character *api.Character, role *string, st *api.Status, readiness *bool) ([]*api.User, error) {
	return s.getUsers(character, role, st, readiness)
}

func (s *Sqlite) UpdateUserStatus(id string, status api.Status) error {
	return s.updateUserStatus(id, status)
}

func (s *Sqlite) UpdateUserReadiness(id string, isReady bool) error {
	return s.updateUserReadiness(id, isReady)
}

func (s *Sqlite) getUserWithToken(token string) (*api.User, error) {
	query := `SELECT * FROM user WHERE token = ?`

	rows, err := s.Query(query, token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	user := &api.User{}
	flag := false

	for rows.Next() {
		flag = true

		err = rows.Scan(&user.Id, &user.Token, &user.Character, &user.Role, &user.Status, &user.Readiness)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "")
	}

	return user, nil
}

func (s *Sqlite) getUserWithID(id string) (*api.User, error) {
	query := `SELECT * FROM user WHERE id = ?`

	rows, err := s.Query(query, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	user := &api.User{}
	flag := false

	for rows.Next() {
		flag = true

		err = rows.Scan(&user.Id, &user.Token, &user.Character, &user.Role, &user.Status, &user.Readiness)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
	}

	if !flag {
		return nil, status.Errorf(codes.NotFound, "")
	}

	return user, nil
}

func (s *Sqlite) getUsers(character *api.Character, role *string, st *api.Status, readiness *bool) ([]*api.User, error) {
	stmt := `SELECT * FROM user WHERE character LIKE ? AND role LIKE ? AND status LIKE ? AND readiness LIKE ?`

	rows, err := s.Query(stmt, character)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not query database: %v", err)
	}
	defer rows.Close()

	users := make([]*api.User, 0)
	userR := &api.User{}

	for rows.Next() {
		err = rows.Scan(&userR.Id, &userR.Token, &userR.Character, &userR.Role, &userR.Status, &userR.Readiness)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not scan results of query: %v", err)
		}
		users = append(users, userR)
	}

	return users, nil
}

func (s *Sqlite) updateUserStatus(id string, st api.Status) error {
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

func (s *Sqlite) updateUserReadiness(id string, isReady bool) error {
	query := `UPDATE user SET readiness = ? WHERE id = ?`

	stmt, err := s.Prepare(query)
	if err != nil {
		return status.Errorf(codes.Internal, "could not prepare statement for database: %v", err)
	}

	if _, err = stmt.Exec(isReady, id); err != nil {
		return status.Errorf(codes.Internal, "could not execute statement on database: %v", err)
	}

	return nil
}

/////////////////////////////////////////////
///////////////////////////// RuleStorage ///
/////////////////////////////////////////////

func (s *Sqlite) GetRules(character api.Character, role string, method api.Method) ([]*api.Key, error) {
	return s.getRules(character, role, method)
}

func (s *Sqlite) getRules(character api.Character, role string, method api.Method) ([]*api.Key, error) {
	selectStatement := `SELECT type, name, namespace FROM role WHERE character = ? AND role = ? AND method = ?`

	rows, err := s.Query(selectStatement, character, role, method)
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
