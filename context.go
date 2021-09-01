package gate

import (
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sync"

	"github.com/julienschmidt/httprouter"
)

type Serializable interface {
	Marshal() ([]byte, error)
}

type Deserializable interface {
	Unmarshal([]byte) error
}

type Payload interface {
	Serializable
	Deserializable
}

type NoPayload []byte

func NOPE() *NoPayload {
	return new(NoPayload)
}

func (zp *NoPayload) Unmarshal(bs []byte) error {
	return nil
}

func (dp NoPayload) Marshal() ([]byte, error) {
	return nil, nil
}

type Handler func(*RequestCtx, Payload) (Payload, error)

type StreamHandler func(*RequestCtx, io.WriteCloser) error

var rcPool sync.Pool

func init() {
	rcPool.New = func() interface{} {
		return new(RequestCtx)
	}
}

type RequestCtx struct {
	Method           string
	URL              *url.URL
	Proto            string // "HTTP/1.0"
	ProtoMajor       int    // 1
	ProtoMinor       int    // 0
	Header           http.Header
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
	r                *http.Request
	rw               http.ResponseWriter
	formOnce         sync.Once
	multiformOnce    sync.Once
	Params           httprouter.Params
}

// Must happen after payload unmarshal
func (rc *RequestCtx) update(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
	rc.r = r
	rc.rw = rw
	rc.Method = rc.r.Method
	rc.URL = rc.r.URL
	rc.Proto = rc.r.Proto
	rc.ProtoMajor = rc.r.ProtoMajor
	rc.Header = rc.r.Header
	rc.ContentLength = rc.r.ContentLength
	rc.TransferEncoding = rc.r.TransferEncoding
	rc.Host = rc.r.Host
	rc.Trailer = rc.r.Trailer
	rc.RemoteAddr = rc.r.RemoteAddr
	rc.RequestURI = rc.r.RequestURI
	rc.TLS = rc.r.TLS
	rc.Params = params
}

func (rc *RequestCtx) parseForm() {
	rc.formOnce.Do(func() {
		rc.r.ParseForm()
	})
}

func (rc *RequestCtx) parseMultipartForm(mm int64) {
	rc.multiformOnce.Do(func() {
		rc.r.ParseMultipartForm(mm)
	})
}

func (rc *RequestCtx) Form() url.Values {
	rc.parseForm()
	return rc.r.Form
}

func (rc *RequestCtx) PostForm() url.Values {
	rc.parseForm()
	return rc.r.PostForm
}

func (rc *RequestCtx) MultipartForm(maxMemory int64) *multipart.Form {
	rc.parseMultipartForm(maxMemory)
	return rc.r.MultipartForm
}

func (rc *RequestCtx) ResponseHeader() http.Header {
	return rc.rw.Header()
}

func (rc *RequestCtx) Reset() {
	rc.Method = ""
	rc.URL = nil
	rc.Proto = ""
	rc.ProtoMajor = 0
	rc.ProtoMinor = 0
	rc.Header = nil
	rc.ContentLength = 0
	rc.TransferEncoding = nil
	rc.Host = ""
	rc.Trailer = nil
	rc.RemoteAddr = ""
	rc.RequestURI = ""
	rc.TLS = nil
	rc.r = nil
	rc.rw = nil
	rc.formOnce = sync.Once{}
	rc.multiformOnce = sync.Once{}
	rc.Params = nil
}
