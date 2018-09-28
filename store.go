package axol

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	scribble "github.com/nanobox-io/golang-scribble"
)

// User represents user
type User struct {
	// user name must be unique
	Name      string   `json:"name"`
	Pass      string   `json:"pass"`
	Projectes []string `json:"projectes"`
}

// Project represents project
type Project struct {
	ID       string   `json:"ID"`
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

func (p *Project) AppendVersion(ver string) {
	p.Versions = append(p.Versions, ver)
}

func (p Project) GetNextVersion() string {
	max := p.getMaxVersionNumber()
	return fmt.Sprintf("v%d", max+1)
}

func (p Project) GetMaxVersion() string {
	max := p.getMaxVersionNumber()
	return fmt.Sprintf("v%d", max)
}

func (p Project) getMaxVersionNumber() int64 {
	if p.Versions == nil {
		return 1
	}

	var max int64
	for _, ver := range p.Versions {
		verNum, err := strconv.ParseInt(strings.TrimLeft(ver, "v"), 10, 32)
		if err != nil {
			log.Println(err)
			return 9999
		}

		if verNum > max {
			max = verNum
		}
	}

	return max
}

// StoreService represents store service
type StoreService interface {
	CreateUser(user *User) error
	UserLogin(name, pass string) bool
	CreateProject(userName string, proj *Project) error
	UpdateProject(proj *Project) error
	GetProjectByID(ID string) (*Project, error)
	ListProjects() ([]*Project, error)
	ListProjectsByUser(name string) ([]*Project, error)
}

type storeService struct {
	db *scribble.Driver
}

// NewStoreService return instance of StoreService
func NewStoreService(dbDir string) (StoreService, error) {
	db, err := scribble.New(dbDir, nil)
	if err != nil {
		return nil, err
	}

	return &storeService{
		db: db,
	}, nil
}

func (s *storeService) CreateUser(user *User) error {
	return s.db.Write("user", user.Name, user)
}

func (s *storeService) UserLogin(name, pass string) bool {
	user := User{}
	if err := s.db.Read("user", name, &user); err != nil {
		return false
	}
	return user.Pass == pass
}

func (s *storeService) CreateProject(userName string, proj *Project) error {
	user := User{}
	if err := s.db.Read("user", userName, &user); err != nil {
		return err
	}
	user.Projectes = append(user.Projectes, proj.ID)

	if err := s.db.Write("proj", proj.ID, proj); err != nil {
		return err
	}

	return s.db.Write("user", userName, user)
}

func (s *storeService) UpdateProject(proj *Project) error {
	if proj.ID == "" {
		return errors.New("missing project ID")
	}

	return s.db.Write("proj", proj.ID, proj)
}

func (s *storeService) GetProjectByID(ID string) (*Project, error) {
	proj := &Project{}
	if err := s.db.Read("proj", ID, proj); err != nil {
		return nil, err
	}

	return proj, nil
}

func (s *storeService) ListProjects() ([]*Project, error) {
	records, err := s.db.ReadAll("proj")
	if err != nil {
		return nil, err
	}

	var projects []*Project
	for _, f := range records {
		proj := Project{}
		if err := json.Unmarshal([]byte(f), &proj); err != nil {
			return nil, err
		}
		projects = append(projects, &proj)
	}

	return projects, nil
}

func (s *storeService) ListProjectsByUser(name string) ([]*Project, error) {
	user := User{}
	if err := s.db.Read("user", name, &user); err != nil {
		return nil, err
	}

	var projects []*Project
	for _, projID := range user.Projectes {
		proj := Project{}
		if err := s.db.Read("proj", projID, &proj); err != nil {
			return nil, err
		}
		projects = append(projects, &proj)
	}
	return projects, nil
}
