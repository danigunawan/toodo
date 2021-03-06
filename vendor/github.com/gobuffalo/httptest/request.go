package httptest

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"

	"github.com/gobuffalo/httptest/internal/takeon/github.com/ajg/form"
	"github.com/gobuffalo/httptest/internal/takeon/github.com/markbates/hmax"
)

type Request struct {
	URL      string
	handler  *Handler
	Headers  map[string]string
	Username string
	Password string
}

func (r *Request) SetBasicAuth(username, password string) {
	r.Username = username
	r.Password = password
}

func (r *Request) Get() *Response {
	req, _ := http.NewRequest("GET", r.URL, nil)
	return r.perform(req)
}

func (r *Request) Delete() *Response {
	req, _ := http.NewRequest("DELETE", r.URL, nil)
	return r.perform(req)
}

func (r *Request) Post(body interface{}) *Response {
	req, _ := http.NewRequest("POST", r.URL, toReader(body))
	r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return r.perform(req)
}

func (r *Request) Put(body interface{}) *Response {
	req, _ := http.NewRequest("PUT", r.URL, toReader(body))
	r.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return r.perform(req)
}

func (r *Request) perform(req *http.Request) *Response {
	if r.handler.HmaxSecret != "" {
		hmax.SignRequest(req, []byte(r.handler.HmaxSecret))
	}
	if r.Username != "" || r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	res := &Response{httptest.NewRecorder()}
	for key, value := range r.Headers {
		req.Header.Set(key, value)
	}
	req.RequestURI = r.URL
	req.Header.Set("Cookie", r.handler.Cookies)
	r.handler.ServeHTTP(res, req)
	c := res.Result().Header.Get("Set-Cookie")
	r.handler.Cookies = c
	return res
}

func toReader(body interface{}) io.Reader {
	if _, ok := body.(encodable); !ok {
		body, _ = form.EncodeToValues(body)
	}
	return strings.NewReader(body.(encodable).Encode())
}

func toURLValues(body interface{}) url.Values {
	m := map[string]interface{}{}
	rv := reflect.Indirect(reflect.ValueOf(body))
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		tf := rt.Field(i)
		rf := rv.Field(i)
		if _, ok := rf.Interface().(multipart.File); ok {
			continue
		}
		if n, ok := tf.Tag.Lookup("form"); ok {
			m[n] = rf.Interface()
			continue
		}
		m[tf.Name] = rf.Interface()
	}

	b, _ := form.EncodeToValues(m)
	return b
}
