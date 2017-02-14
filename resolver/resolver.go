package resolver

import (
	"fmt"
	"net/http"
)

//RequestURL simply returns
//the request url from REQUEST_URI header
//this should not be done in production applications
type RequestURL struct {
	r    http.Request
	Port int
}

//SetRequest to implement `RequestAwareResolverInterface`
func (m *RequestURL) SetRequest(r http.Request) {
	m.r = r
}

//GetBaseURL implements `URLResolver` interface
func (m RequestURL) GetBaseURL() string {
	var baseURL = ""
	if uri := m.r.Header.Get("REQUEST_URI"); uri != "" {
		baseURL = uri
	} else {
		baseURL = fmt.Sprintf("http://localhost:%d", m.Port)
	}
	return baseURL
}
