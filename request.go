package easyroute

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Request struct {
	urlVars     map[string]string
	urlParams   map[string][]string
	writer      http.ResponseWriter
	request     *http.Request
	userUuid    string
	SessionInfo map[string]interface{}
}

// Create a new router.Request based on an http Request and ResponseWriter
func NewRequest(w http.ResponseWriter, r *http.Request) Request {
	r.ParseForm()
	request := Request{
		urlVars:   mux.Vars(r),
		urlParams: r.Form,
		writer:    w,
		request:   r,
	}
	return request
}

// Writer returns an io.Writer for the request. Can be
// used to output templated html and the like.
func (r *Request) Writer() io.Writer {
	return r.writer
}

// Vars returns route variables for the request.
func (r *Request) Vars() map[string]string {
	return r.urlVars
}

// Params returns Url Params for the request.
func (r *Request) Params() map[string][]string {
	return r.urlParams
}

// ParamValue returns the first value of the given url param.
func (r *Request) ParamValue(key string) string {
	if vs := r.urlParams[key]; len(vs) > 0 {
		return vs[0]
	}
	return ""
}

func (r *Request) Body(v interface{}) error {
	var err error
	if r.request.Method == "POST" || r.request.Method == "PUT" {
		if strings.ToLower(r.request.Header.Get("Content-Type")) == "application/json;charset=utf-8" {
			if err = json.NewDecoder(r.request.Body).Decode(v); err != nil {
				return err
			}
		} else {
			v, err = ioutil.ReadAll(r.request.Body)
		}
	}
	return err
}

// Path returns Url Path for the request.
func (r *Request) Path() string {
	return r.request.URL.Path
}

// Origin returns the host where the request originated.
func (r *Request) Origin() string {
	return r.request.RemoteAddr
}

// Redirect forwards the request to the path provided by forwardTo
func (r *Request) Redirect(forwardTo string) {
	http.Redirect(r.writer, r.request, forwardTo, http.StatusFound)
}

// Json sends json representation of the interface v with the provided
// status code to the client.
func (r *Request) Json(status int, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		r.sendJson(r.writer, status, b)
	}
}

// SendFile forwards the file at the provided path to the client.
func (r *Request) SendFile(file string) {
	http.ServeFile(r.writer, r.request, file)
}

func (r *Request) sendJson(w http.ResponseWriter, status int, b []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(b)
	w.Write([]byte("\n"))
}
