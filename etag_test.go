package etag

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"crypto/sha1"
	"encoding/hex"

	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/suite"
)

var testStrBytes = []byte("Hello World")
var testStrEtag = "11-0a4d55a8d778e5022fab701977c5d840bbc486d0"

type EmptyEtagSuite struct {
	suite.Suite

	server *httptest.Server
}

func (s *EmptyEtagSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.Handle("/", Handler(http.HandlerFunc(emptyHandlerFunc), false))

	s.server = httptest.NewServer(mux)
}

func (s *EmptyEtagSuite) TestNoEtag() {
	res, err := http.Get(s.server.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusNoContent, res.StatusCode)
	s.Empty(res.Header.Get(headers.ETag))
}

func TestEmptyEtag(t *testing.T) {
	suite.Run(t, new(EmptyEtagSuite))
}

type EtagSuite struct {
	suite.Suite

	server     *httptest.Server
	weakServer *httptest.Server
}

func (s *EtagSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.Handle("/", Handler(http.HandlerFunc(handlerFunc), false))

	s.server = httptest.NewServer(mux)

	wmux := http.NewServeMux()
	wmux.Handle("/", Handler(http.HandlerFunc(handlerFunc), true))

	s.weakServer = httptest.NewServer(wmux)
}

func (s EtagSuite) TestEtagExists() {
	res, err := http.Get(s.server.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	h := sha1.New()
	h.Write(testStrBytes)

	s.Equal(fmt.Sprintf("%v-%v", len(testStrBytes), hex.EncodeToString(h.Sum(nil))), res.Header.Get(headers.ETag))
}

func (s EtagSuite) TestWeakEtagExists() {
	res, err := http.Get(s.weakServer.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	h := sha1.New()
	h.Write(testStrBytes)

	s.Equal(fmt.Sprintf("W/%v-%v", len(testStrBytes), hex.EncodeToString(h.Sum(nil))), res.Header.Get(headers.ETag))
}

func (s EtagSuite) TestMatch() {
	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(headers.IfNoneMatch, testStrEtag)

	cli := &http.Client{}
	res, err := cli.Do(req)

	s.Nil(err)
	s.Equal(http.StatusNotModified, res.StatusCode)
}

func TestEtag(t *testing.T) {
	suite.Run(t, new(EtagSuite))
}

func emptyHandlerFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)

	res.Write(nil)
}

func handlerFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)

	res.Write(testStrBytes)
}
