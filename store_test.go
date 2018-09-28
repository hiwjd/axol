package axol

import (
	"os"
	"testing"
)

var dbDir = "./test-db"

func clean() {
	os.RemoveAll(dbDir)
}

func TestCreateUser(t *testing.T) {
	s, err := NewStoreService(dbDir)
	if err != nil {
		t.Errorf("NewStoreService fail with err: %s\n", err)
	}

	user := &User{
		Name:      "hiwjd",
		Pass:      "123456",
		Projectes: nil,
	}
	err = s.CreateUser(user)
	if err != nil {
		t.Errorf("CreateUser fail with err: %s\n", err)
	}

	clean()
}

func TestUserLogin(t *testing.T) {
	s, _ := NewStoreService(dbDir)

	user := &User{
		Name:      "hiwjd",
		Pass:      "123456",
		Projectes: nil,
	}
	s.CreateUser(user)

	if !s.UserLogin("hiwjd", "123456") {
		t.Error("UserLogin should success")
	}

	clean()
}

func TestCreateProject(t *testing.T) {
	s, _ := NewStoreService(dbDir)

	user := &User{
		Name:      "hiwjd",
		Pass:      "123456",
		Projectes: nil,
	}
	s.CreateUser(user)

	proj := &Project{
		ID:       "proj1",
		Name:     "proj1",
		Versions: nil,
	}
	err := s.CreateProject("hiwjd", proj)
	if err != nil {
		t.Errorf("CreateProject fail with error: %s\n", err)
	}

	projs, err := s.ListProjects()
	if err != nil {
		t.Errorf("ListProjects fail with error: %s\n", err)
	}
	if len(projs) != 1 {
		t.Error("ListProjects should return 1 project")
	}

	projs, err = s.ListProjectsByUser("hiwjd")
	if err != nil {
		t.Errorf("ListProjectsByUser fail with error: %s\n", err)
	}
	if len(projs) != 1 {
		t.Error("ListProjectsByUser should return 1 project")
	}

	projs, err = s.ListProjectsByUser("userNotExists")
	if err == nil {
		t.Error("ListProjectsByUser should throw error")
	}

	clean()
}
