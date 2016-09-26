package mock

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type Server struct {
	Addr    string
	Port    string
	Scheme  string
	server  *httptest.Server
	mux     *http.ServeMux
	routers []*RouterMap
}

type RouterMap struct {
	Path     string
	Method   string
	Response *Response
	Request  *Request
}

type Request struct {
	BodyBuffer []byte
	Error      error
}

type Response struct {
	StatusCode int
	BodyBuffer []byte
	Error      error
}

func NewServer() *Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	parsedUrl, _ := url.Parse(server.URL)
	host, port, _ := net.SplitHostPort(parsedUrl.Host)
	return &Server{
		mux:    mux,
		Addr:   host,
		Port:   port,
		Scheme: parsedUrl.Scheme,
		server: server,
	}
}

func (s *Server) Close() {
	s.server.Close()
}

func (s *Server) Register() {
	for _, router := range s.routers {
		s.mux.HandleFunc(router.Path, router.Handler)
	}
}

func testMethod(r *http.Request, want string) {
	if got := r.Method; got != want {
		log.Errorf("Request method: %v, want %v", got, want)
	}
}

func (s *Server) AddRouter(path string, method string) *RouterMap {
	method = strings.ToUpper(method)
	routerMap := &RouterMap{
		Path:     path,
		Method:   method,
		Request:  &Request{},
		Response: &Response{},
	}
	s.routers = append(s.routers, routerMap)
	return routerMap
}

func (rm *RouterMap) RBody(body io.Reader) *RouterMap {
	rm.Request.BodyBuffer, rm.Request.Error = ioutil.ReadAll(body)
	return rm
}

func (rm *RouterMap) RBodyString(body string) *RouterMap {
	rm.Request.BodyBuffer = []byte(body)
	return rm
}

func (rm *RouterMap) RFile(path string) *RouterMap {
	rm.Request.BodyBuffer, rm.Request.Error = ioutil.ReadFile(path)
	return rm
}

func (rm *RouterMap) RJSON(data interface{}) *RouterMap {
	rm.Request.BodyBuffer, rm.Request.Error = readAndDecode(data, "json")
	return rm
}

func (rm *RouterMap) Reply(status int) *RouterMap {
	rm.Response.StatusCode = status
	return rm
}

func (rm *RouterMap) WBody(body io.Reader) *RouterMap {
	rm.Response.BodyBuffer, rm.Response.Error = ioutil.ReadAll(body)
	return rm
}

func (rm *RouterMap) WBodyString(body string) *RouterMap {
	rm.Response.BodyBuffer = []byte(body)
	return rm
}

func (rm *RouterMap) WFile(path string) *RouterMap {
	rm.Response.BodyBuffer, rm.Response.Error = ioutil.ReadFile(path)
	return rm
}

func (rm *RouterMap) WJSON(data interface{}) *RouterMap {
	rm.Response.BodyBuffer, rm.Response.Error = readAndDecode(data, "json")
	return rm
}

func (rm *RouterMap) Handler(w http.ResponseWriter, r *http.Request) {
	testMethod(r, rm.Method)
	rBody, _ := ioutil.ReadAll(r.Body)
	if bytes.Equal(rBody, rm.Request.BodyBuffer) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(rm.Response.StatusCode)
		w.Write(rm.Response.BodyBuffer)
	} else {
		http.Error(w, `{"message": "body is not equal"}`, 400)
	}
}

func readAndDecode(data interface{}, kind string) ([]byte, error) {
	buf := &bytes.Buffer{}

	switch data.(type) {
	case string:
		buf.WriteString(data.(string))
	case []byte:
		buf.Write(data.([]byte))
	default:
		var err error
		if kind == "xml" {
			err = xml.NewEncoder(buf).Encode(data)
		} else {
			err = json.NewEncoder(buf).Encode(data)
		}
		if err != nil {
			return nil, err
		}
	}

	return ioutil.ReadAll(buf)
}
