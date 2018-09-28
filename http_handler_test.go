package axol

import (
	"net/http"
	"testing"
)

type matchRequestTypeCase struct {
	method string
	url    string
	rt     RequestType
}

func TestMatchRequestType(t *testing.T) {
	hh := &HttpHandler{}
	cases := []matchRequestTypeCase{
		matchRequestTypeCase{"POST", "/sign_up", SIGN_UP},
		matchRequestTypeCase{"POST", "/sign_in", SIGN_IN},
		matchRequestTypeCase{"POST", "/project/create", PROJ_CREATE},
		matchRequestTypeCase{"POST", "/project/upload", PROJ_UPLOAD},
		matchRequestTypeCase{"GET", "/project/all", PROJ_ALL},
		matchRequestTypeCase{"GET", "/project/all", PROJ_ALL},
		matchRequestTypeCase{"GET", "/p/latest/8941966ae7817d063d1f2be0c1d558b2/index.html", PROJ_VIEW_LATEST},
		matchRequestTypeCase{"GET", "/p/latest/8941966ae7817d063d1f2be0c1d558b2/sub/a.html", PROJ_VIEW_LATEST},
		matchRequestTypeCase{"GET", "/p/v1/8941966ae7817d063d1f2be0c1d558b2/index.html", PROJ_VIEW_VERSION},
		matchRequestTypeCase{"GET", "/p/v1/8941966ae7817d063d1f2be0c1d558b2/sub/b.html", PROJ_VIEW_VERSION},
	}

	for _, c := range cases {
		req, _ := http.NewRequest(c.method, c.url, nil)
		if hh.matchRequestType(req) != c.rt {
			t.Errorf("mthod:%s url:%s should match request type:%d\n", c.method, c.url, c.rt)
		}
	}
}
