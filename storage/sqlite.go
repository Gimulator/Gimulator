package storage

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Gimulator/Gimulator/config"
	"github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var memoryPath string = ":memory:"

type Sqlite struct {
	*gorm.DB
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
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		s.log.WithError(err).Error("could not open sqlite db")
		return err
	}
	s.DB = db

	s.log.Info("starting to create tables")
	if err := s.AutoMigrate(&User{}, &Rule{}, &Message{}); err != nil {
		s.log.WithError(err).Error("could not create tables")
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

func (s *Sqlite) fillUserTable(config *config.Config) error {
	for _, cred := range config.Credentials {
		if err := s.insertUser(&User{
			Name:      cred.Name,
			Token:     cred.Token,
			Character: cred.Character,
			Role:      cred.Role,
			Readiness: false,
			Status:    api.Status_unknown,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sqlite) fillRuleTable(config *config.Config) error {
	for _, rule := range config.Character.Director {
		for _, method := range rule.Methods {
			if err := s.insertRule(&Rule{
				Method:    method,
				Type:      rule.Key.Type,
				Name:      rule.Key.Name,
				Namespace: rule.Key.Namespace,
				Role:      "",
				Character: api.Character_name[int32(api.Character_director)],
			}); err != nil {
				return err
			}
		}
	}

	for _, rule := range config.Character.Operator {
		for _, method := range rule.Methods {
			if err := s.insertRule(&Rule{
				Method:    method,
				Type:      rule.Key.Type,
				Name:      rule.Key.Name,
				Namespace: rule.Key.Namespace,
				Role:      "",
				Character: api.Character_operator,
			}); err != nil {
				return err
			}
		}
	}

	for _, rule := range config.Character.Master {
		for _, method := range rule.Methods {
			if err := s.insertRule(&Rule{
				Method:    method,
				Type:      rule.Key.Type,
				Name:      rule.Key.Name,
				Namespace: rule.Key.Namespace,
				Role:      "",
				Character: api.Character_master,
			}); err != nil {
				return err
			}
		}
	}

	for role, rules := range config.Character.Actors {
		for _, rule := range rules {
			for _, method := range rule.Methods {
				if err := s.insertRule(&Rule{
					Method:    method,
					Type:      rule.Key.Type,
					Name:      rule.Key.Name,
					Namespace: rule.Key.Namespace,
					Role:      role,
					Character: api.Character_actor,
				}); err != nil {
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
	// **Attention**: I suppose the owner of message was validated before calling this fucntion,
	// the reason is performance.
	mes := &Message{
		Type:      message.Key.Type,
		Name:      message.Key.Name,
		Namespace: message.Key.Namespace,
		UserName:  message.Meta.Owner.Name,
		Content:   message.Content,
	}

	if err := s.insertOrUpdateMessage(mes); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not put message=%v: %v", message, err.Error()))
	}
	return nil
}

func (s *Sqlite) Get(key *api.Key) (*api.Message, error) {
	messages, err := s.selectMessage(key.Type, key.Name, key.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not get message with key=%v: %v", key, err.Error()))
	}

	if len(messages) < 1 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("could not find any message with key=%v: %v", key, err.Error()))
	}

	return s.sqliteToAPIMessage(messages[0]), nil
}

func (s *Sqlite) GetAll(key *api.Key) ([]*api.Message, error) {
	messages, err := s.selectMessage(key.Type, key.Name, key.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not get all messages with key=%v: %v", key, err.Error()))
	}

	res := make([]*api.Message, 0)
	for _, mes := range messages {
		res = append(res, s.sqliteToAPIMessage(mes))
	}
	return res, nil
}

func (s *Sqlite) Delete(key *api.Key) error {
	if err := s.deleteMessage(key.Type, key.Name, key.Namespace); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not delete message with key=%v: %v", key, err.Error()))
	}
	return nil
}

func (s *Sqlite) DeleteAll(key *api.Key) error {
	if err := s.deleteMessage(key.Type, key.Name, key.Namespace); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not delete all messages with key=%v: %v", key, err.Error()))
	}
	return nil
}

func (s *Sqlite) insertOrUpdateMessage(message *Message) error {
	if err := s.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "name"}, {Name: "namespace"}, {Name: "type"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"user_name", "content"}),
	}).Create(message).Error; err != nil {
		return err
	}
	return nil
}

func (s *Sqlite) selectMessage(typ, name, namespace string) ([]*Message, error) {
	messages := []*Message{}

	db := s.DB
	if typ != "" {
		db = db.Where("type = ?", typ)
	}
	if name != "" {
		db = db.Where("name = ?", name)
	}
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}

	if err := db.Preload("User").Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *Sqlite) deleteMessage(typ, name, namespace string) error {
	db := s.DB

	if typ != "" {
		db = db.Where("type = ?", typ)
	}
	if name != "" {
		db = db.Where("name = ?", name)
	}
	if namespace != "" {
		db = db.Where("namespace = ?", namespace)
	}

	return db.Delete(&Message{}).Error
}

///////////////////////////////////////////////
/////////////////////////////// UserStorage ///
///////////////////////////////////////////////

func (s *Sqlite) GetUserWithName(n string) (*api.User, error) {
	user, err := s.selectUserWithName(n)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("could not found any user with name=%v", n))
	} else if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not found any user with name=%v: %v", n, err.Error()))
	}
	return s.sqliteToAPIUser(user), nil
}

func (s *Sqlite) GetUserWithToken(t string) (*api.User, error) {
	user, err := s.selectUserWithToken(t)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("could not find any user with token=%v", t))
	} else if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not find any user with token=%v: %v", t, err.Error()))
	}
	return s.sqliteToAPIUser(user), nil
}

func (s *Sqlite) GetUsers(n *string, t *string, c *api.Character, r *string, rd *bool, st *api.Status) ([]*api.User, error) {
	res := make([]*api.User, 0)

	users, err := s.selectUsers(n, t, c, r, rd, st)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not find any user with token=%v: %v", t, err.Error()))
	}

	for _, u := range users {
		res = append(res, s.sqliteToAPIUser(u))
	}

	return res, nil
}

func (s *Sqlite) UpdateUserReadiness(name string, readiness bool) error {
	if err := s.updateUser(name, &readiness, nil); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not update readiness of user with name=%v: %v", name, err.Error()))
	}
	return nil
}

func (s *Sqlite) UpdateUserStatus(name string, st api.Status) error {
	if err := s.updateUser(name, nil, &st); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("could not update status of user with name=%v: %v", name, err.Error()))
	}
	return nil
}

func (s *Sqlite) selectUserWithName(n string) (*User, error) {
	user := &User{}
	if err := s.Where("name = ?", n).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Sqlite) selectUserWithToken(t string) (*User, error) {
	user := &User{}
	if err := s.Where("token = ?", t).First(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Sqlite) insertUser(user *User) error {
	if err := s.Create(user).Error; err != nil {
		return err
	}
	return nil
}

func (s *Sqlite) updateUser(name string, r *bool, st *api.Status) error {
	db := s.Model(&User{}).Where("name = ?", name)
	updates := make(map[string]interface{})

	if r != nil {
		updates["readiness"] = r
	}

	if st != nil {
		updates["status"] = st
	}

	if err := db.Updates(updates).Error; err != nil {
		return err
	}

	tx := db.Updates(updates)

	if err := tx.Error; err != nil {
		return err
	}

	totalAffectedRows := tx.RowsAffected
	if totalAffectedRows == 1 { //chon vase 1 usere
		updates["LastUserStatusUpdateTime"] = time.Now().Format(time.StampMicro)
	}

	return nil
}

func (s *Sqlite) selectUsers(n *string, t *string, c *api.Character, r *string, rd *bool, st *api.Status) ([]*User, error) {
	users := []*User{}

	db := s.DB
	if n != nil {
		db = db.Where("name = ?", n)
	}
	if t != nil {
		db = db.Where("token = ?", t)
	}
	if c != nil {
		db = db.Where("character = ?", c)
	}
	if r != nil {
		db = db.Where("role = ?", r)
	}
	if rd != nil {
		db = db.Where("readiness = ?", rd)
	}
	if st != nil {
		db = db.Where("status = ?", st)
	}

	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

///////////////////////////////////////////////
/////////////////////////////// RuleStorage ///
///////////////////////////////////////////////

func (s *Sqlite) GetRules(character api.Character, role string, method api.Method) ([]*api.Key, error) {
	rules, err := s.selectRules(&character, &role, &method)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not find rules for character=%v, role=%v, and method=%v", character, role, method))
	}

	keys := make([]*api.Key, 0)
	for _, rule := range rules {
		keys = append(keys, s.sqliteRuleToAPIKey(rule))
	}

	return keys, nil
}

func (s *Sqlite) insertRule(rule *Rule) error {
	if err := s.Create(rule).Error; err != nil {
		return err
	}
	return nil
}

func (s *Sqlite) selectRules(c *api.Character, r *string, m *api.Method) ([]*Rule, error) {
	rules := []*Rule{}

	db := s.DB
	if c != nil {
		db = db.Where("character = ?", c)
	}
	if r != nil {
		db = db.Where("role = ?", r)
	}
	if m != nil {
		db = db.Where("method = ?", m)
	}

	if err := db.Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

///////////////////////////////////////////////
//////////////////////////////////// Helper ///
///////////////////////////////////////////////
func (s *Sqlite) sqliteToAPIMessage(src *Message) *api.Message {
	return &api.Message{
		Key: &api.Key{
			Type:      src.Type,
			Name:      src.Name,
			Namespace: src.Namespace,
		},
		Meta: &api.Meta{
			Owner:        s.sqliteToAPIUser(src.User),
			CreationTime: timestamppb.New(src.UpdatedAt),
		},
		Content: src.Content,
	}
}

func (s *Sqlite) sqliteToAPIUser(src *User) *api.User {
	return &api.User{
		Name:      src.Name,
		Character: src.Character,
		Role:      src.Role,
		Readiness: src.Readiness,
		Status:    src.Status,
	}
}

func (s *Sqlite) sqliteRuleToAPIKey(src *Rule) *api.Key {
	return &api.Key{
		Type:      src.Type,
		Name:      src.Name,
		Namespace: src.Namespace,
	}
}

/////////////////////////////////////////////
/////////////////////////////////// Types ///
/////////////////////////////////////////////
type User struct {
	CreatedAt                time.Time      `gorm:""`
	UpdatedAt                time.Time      `gorm:""`
	DeletedAt                gorm.DeletedAt `gorm:"index"`
	Name                     string         `gorm:"primaryKey;autoIncrement:false;notNull"`
	Token                    string         `gorm:"unique;index;notNull"`
	Role                     string         `gorm:"notNull;default:''"`
	Readiness                bool           `gorm:"notNull;default:false"`
	Character                api.Character  `gorm:"notNull;default:0"`
	Status                   api.Status     `gorm:"notNull;default:0"`
	LastUserStatusUpdateTime time.Time      `gorm:""`
}

type Rule struct {
	gorm.Model
	Method    api.Method    `gorm:"notNull;default:0"`
	Type      string        `gorm:"notNull;default:''"`
	Name      string        `gorm:"notNull;default:''"`
	Namespace string        `gorm:"notNull;default:''"`
	Role      string        `gorm:"notNull;default:''"`
	Character api.Character `gorm:"notNull;default:0"`
}

type Message struct {
	CreatedAt time.Time      `gorm:""`
	UpdatedAt time.Time      `gorm:""`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Type      string         `gorm:"primaryKey;autoIncrement:false;notNull"`
	Name      string         `gorm:"primaryKey;autoIncrement:false;notNull"`
	Namespace string         `gorm:"primaryKey;autoIncrement:false;notNull"`
	Content   string         `gorm:"notNull;default:''"`
	UserName  string         `gorm:"notNull"`
	User      *User          `gorm:"foreignKey:UserName;references:Name;notNull"`
}
