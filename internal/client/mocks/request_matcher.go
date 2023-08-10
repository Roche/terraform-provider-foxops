package client_mocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"go.uber.org/mock/gomock"
)

type RequestMatcher struct {
	method  *string
	path    *string
	headers map[string][]string
	body    []byte
}

type RequestMatcherOption interface {
	apply(*RequestMatcher)
}

type requestMethod string

func RequestMethod(method string) RequestMatcherOption {
	return requestMethod(method)
}

func (m requestMethod) apply(matcher *RequestMatcher) {
	matcher.method = (*string)(&m)
}

type requestPath string

func RequestPath(url string) RequestMatcherOption {
	return requestPath(url)
}

func RequestPathf(url string, args ...interface{}) RequestMatcherOption {
	return requestPath(fmt.Sprintf(url, args...))
}

func (p requestPath) apply(matcher *RequestMatcher) {
	matcher.path = (*string)(&p)
}

type requestHeader []string

func RequestHeader(key string, value string, values ...string) RequestMatcherOption {
	header := []string{key, value}
	header = append(header, values...)
	return requestHeader(header)
}

func (h requestHeader) apply(matcher *RequestMatcher) {
	matcher.headers[http.CanonicalHeaderKey(h[0])] = h[1:]
}

type requestBody []byte

func RequestBody(body interface{}) RequestMatcherOption {
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return requestBody(data)
}

func (b requestBody) apply(matcher *RequestMatcher) {
	matcher.body = b
}

func NewRequestMatcher(opts ...RequestMatcherOption) gomock.Matcher {
	matcher := &RequestMatcher{headers: make(map[string][]string)}
	for _, opt := range opts {
		opt.apply(matcher)
	}
	return matcher
}

func (m *RequestMatcher) String() string {
	return ""
}

func (m *RequestMatcher) Matches(x interface{}) bool {
	req, ok := x.(*http.Request)

	if !ok {
		return false
	}

	if m.method != nil && *m.method != req.Method {
		return false
	}

	if m.path != nil && *m.path != req.URL.Path {
		return false
	}

	for key, values := range m.headers {
		if !reflect.DeepEqual(req.Header[key], values) {
			return false
		}
	}

	if m.body != nil {
		data, err := io.ReadAll(req.Body)
		if err != nil {
			return false
		}
		if !reflect.DeepEqual(data, m.body) {
			return false
		}
	}

	return true
}
