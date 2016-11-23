package etag

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"hash"
	"net/http"
	"strconv"

	"github.com/go-http-utils/fresh"
	"github.com/go-http-utils/headers"
)

// Version is this package's version.
const Version = "0.0.1"

type hashWriter struct {
	rw     http.ResponseWriter
	hash   hash.Hash
	buf    *bytes.Buffer
	len    int
	status int
}

func (hw hashWriter) Header() http.Header {
	return hw.rw.Header()
}

func (hw *hashWriter) WriteHeader(status int) {
	hw.status = status
}

func (hw *hashWriter) Write(b []byte) (int, error) {
	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	// bytes.Buffer.Write(b) always return (len(b), nil), so just
	// ignore the return values.
	hw.buf.Write(b)

	return hw.hash.Write(b)
}

// Handler wraps the http.Handler h with ETag support.
func Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		hw := hashWriter{rw: res, hash: sha1.New()}
		h.ServeHTTP(&hw, req)

		resHeader := res.Header()

		if hw.hash == nil ||
			resHeader.Get(headers.ETag) != "" ||
			strconv.Itoa(hw.status)[0] != '2' {
			return
		}

		resHeader.Set(headers.ETag,
			fmt.Sprintf(`"%v-%v"`, strconv.Itoa(hw.len), string(hw.hash.Sum(nil))))

		if fresh.IsFresh(req.Header, resHeader) {
			res.WriteHeader(http.StatusNotModified)
			res.Write(nil)
		} else {
			res.WriteHeader(hw.status)
			res.Write(hw.buf.Bytes())
		}
	})
}
