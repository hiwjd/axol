package axol

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type RequestType int

const (
	SIGN_UP RequestType = iota
	SIGN_IN
	PROJ_CREATE
	PROJ_ALL
	PROJ_MY
	PROJ_UPLOAD
	PROJ_VIEW_VERSION
	PROJ_VIEW_LATEST
	INDEX
	STATIC
	NOT_FOUND
	NOT_SUPPORT
)

func NewHttpHandler(ss StoreService, projDir string) *HttpHandler {
	hh := &HttpHandler{
		ss:       ss,
		projDir:  projDir,
		sessions: make(map[string]string, 128),
	}

	hh.handlers = make(map[RequestType]func(http.ResponseWriter, *http.Request), 10)
	hh.handlers[SIGN_UP] = hh.signup
	hh.handlers[SIGN_IN] = hh.signin
	hh.handlers[PROJ_CREATE] = hh.projectCreate
	hh.handlers[PROJ_ALL] = hh.projectAll
	hh.handlers[PROJ_MY] = hh.projectMy
	hh.handlers[PROJ_UPLOAD] = hh.projectUpload
	hh.handlers[PROJ_VIEW_VERSION] = hh.projectViewVersion
	hh.handlers[PROJ_VIEW_LATEST] = hh.projectViewLatest
	hh.handlers[NOT_FOUND] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}
	hh.handlers[NOT_SUPPORT] = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	}
	hh.handlers[INDEX] = func(w http.ResponseWriter, r *http.Request) {
		fpath := "./static/index.html"
		http.ServeFile(w, r, fpath)
	}
	hh.handlers[STATIC] = func(w http.ResponseWriter, r *http.Request) {
		// /static/*
		sArr := strings.Split(r.URL.Path, "/")
		params := []string{"./static"}
		params = append(params, sArr[2:]...)
		fpath := path.Join(params...)

		http.ServeFile(w, r, fpath)
	}

	return hh
}

type HttpHandler struct {
	ss       StoreService
	projDir  string
	handlers map[RequestType]func(http.ResponseWriter, *http.Request)
	sessions map[string]string
}

func (hh *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// /sign_up
	// /sign_in
	// /project/create
	// /project/all
	// /project/my
	// /project/upload
	// /p/latest/8941966ae7817d063d1f2be0c1d558b2/index.html
	// /p/v3/8941966ae7817d063d1f2be0c1d558b2/subdir/index.html
	fmt.Println(r.URL.Path)

	rt := hh.matchRequestType(r)
	if handler, ok := hh.handlers[rt]; ok {
		handler(w, r)
	} else {
		w.WriteHeader(400)
	}
}

func (hh *HttpHandler) signup(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("pass")
	if name == "" || pass == "" {
		w.WriteHeader(400)
		w.Write([]byte("请输入用户名密码"))
		return
	}

	user := &User{Name: name, Pass: pass}
	if err := hh.ss.CreateUser(user); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("用户注册失败"))
	}

	sid := uuid.NewV4().String()
	hh.sessions[sid] = name

	w.WriteHeader(200)
	w.Write([]byte(sid))
}
func (hh *HttpHandler) signin(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("pass")
	if name == "" || pass == "" {
		w.WriteHeader(400)
		w.Write([]byte("请输入用户名密码"))
		return
	}

	if !hh.ss.UserLogin(name, pass) {
		w.WriteHeader(400)
		w.Write([]byte("登录失败"))
		return
	}

	sid := uuid.NewV4().String()
	hh.sessions[sid] = name

	w.WriteHeader(200)
	w.Write([]byte(sid))
}
func (hh *HttpHandler) projectCreate(w http.ResponseWriter, r *http.Request) {
	un := hh.checkLogin(w, r)
	if un == "" {
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(400)
		w.Write([]byte("请输入项目名"))
		return
	}

	s := fmt.Sprintf("%s-%s-%d", un, name, time.Nanosecond)
	ID := fmt.Sprintf("%x", md5.Sum([]byte(s)))

	proj := &Project{
		ID:   ID,
		Name: name,
	}
	if err := hh.ss.CreateProject(un, proj); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("项目创建失败"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("项目创建成功"))
}
func (hh *HttpHandler) projectUpload(w http.ResponseWriter, r *http.Request) {
	un := hh.checkLogin(w, r)
	if un == "" {
		return
	}

	projID := r.URL.Query().Get("projID")
	if projID == "" {
		w.WriteHeader(400)
		w.Write([]byte("请输入项目ID"))
		return
	}

	projs, err := hh.ss.ListProjectsByUser(un)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("您无项目"))
		return
	}

	var projFind *Project
	for _, proj := range projs {
		if proj.ID == projID {
			projFind = proj
			break
		}
	}

	if projFind == nil {
		w.WriteHeader(500)
		w.Write([]byte("没有这个项目"))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("上传失败"))
		return
	}

	// 版本+1
	nv := projFind.GetNextVersion()
	dstDir := path.Join(hh.projDir, projFind.ID, nv)

	rd := bytes.NewReader(body)

	if err = UnzipFile(rd, dstDir); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`上传失败`))
		return
	}

	projFind.AppendVersion(nv)
	if err := hh.ss.UpdateProject(projFind); err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("上传失败"))
		return
	}

	w.Write([]byte("上传成功"))
}
func (hh *HttpHandler) projectAll(w http.ResponseWriter, r *http.Request) {
	projs, err := hh.ss.ListProjects()
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("获取项目列表失败"))
		return
	}

	bs, err := json.Marshal(projs)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("获取项目列表失败"))
		return
	}

	w.Write(bs)
}
func (hh *HttpHandler) projectMy(w http.ResponseWriter, r *http.Request) {
	un := hh.checkLogin(w, r)
	if un == "" {
		return
	}

	projs, err := hh.ss.ListProjectsByUser(un)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("获取项目列表失败"))
		return
	}

	bs, err := json.Marshal(projs)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("获取项目列表失败"))
		return
	}

	w.Write(bs)
}

// /p/latest/8941966ae7817d063d1f2be0c1d558b2/index.html
func (hh *HttpHandler) projectViewLatest(w http.ResponseWriter, r *http.Request) {
	sArr := strings.Split(r.URL.Path, "/")
	if len(sArr) < 5 {
		w.WriteHeader(404)
		w.Write([]byte("没有这个页面"))
		return
	}

	projID := sArr[3]
	proj, err := hh.ss.GetProjectByID(projID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte("没有这个页面"))
		return
	}

	params := []string{hh.projDir, proj.ID, proj.GetMaxVersion()}
	params = append(params, sArr[4:]...)
	fpath := path.Join(params...)

	http.ServeFile(w, r, fpath)
}

// /p/v2/8941966ae7817d063d1f2be0c1d558b2/index.html
func (hh *HttpHandler) projectViewVersion(w http.ResponseWriter, r *http.Request) {
	sArr := strings.Split(r.URL.Path, "/")
	if len(sArr) < 5 {
		w.WriteHeader(404)
		w.Write([]byte("没有这个页面"))
		return
	}

	projID := sArr[3]
	proj, err := hh.ss.GetProjectByID(projID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(404)
		w.Write([]byte("没有这个页面"))
		return
	}

	params := []string{hh.projDir, proj.ID, sArr[2]}
	params = append(params, sArr[4:]...)
	fpath := path.Join(params...)

	http.ServeFile(w, r, fpath)
}

func (hh *HttpHandler) matchRequestType(r *http.Request) RequestType {
	path := r.URL.Path
	if r.Method == "POST" {
		if path == "/sign_up" {
			return SIGN_UP
		} else if path == "/sign_in" {
			return SIGN_IN
		} else if path == "/project/create" {
			return PROJ_CREATE
		} else if path == "/project/upload" {
			return PROJ_UPLOAD
		} else {
			return NOT_FOUND
		}
	} else if r.Method == "GET" {
		if path == "/project/all" {
			return PROJ_ALL
		} else if path == "/project/my" {
			return PROJ_MY
		} else if strings.HasPrefix(path, "/p/latest/") {
			return PROJ_VIEW_LATEST
		} else if strings.HasPrefix(path, "/p/v") {
			return PROJ_VIEW_VERSION
		} else if strings.HasPrefix(path, "/static/") {
			return STATIC
		} else if path == "/" || path == "" {
			return INDEX
		} else {
			return NOT_FOUND
		}
	} else {
		return NOT_SUPPORT
	}
}

func (hh *HttpHandler) checkLogin(w http.ResponseWriter, r *http.Request) string {
	a := r.Header.Get("Authorization")
	if a == "" {
		w.WriteHeader(400)
		w.Write([]byte(`没有登录`))
		return ""
	}

	if un, ok := hh.sessions[a]; ok {
		return un
	}

	w.WriteHeader(400)
	w.Write([]byte(`没有登录`))
	return ""
}
