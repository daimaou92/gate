package gate

import (
	"net/http"
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
	ContentType() ContentType
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

type RequestData struct {
	Params      httprouter.Params
	Body        Payload
	QueryParams Payload
}

type Handler func(*RequestCtx, *RequestData) (Payload, error)

// type StreamHandler func(*RequestCtx, io.WriteCloser) error

var rcPool sync.Pool

func init() {
	rcPool.New = func() interface{} {
		return new(RequestCtx)
	}
}

type RequestCtx struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

// Must happen after payload unmarshal
func (rc *RequestCtx) update(rw http.ResponseWriter, r *http.Request) {
	rc.Request = r
	rc.ResponseWriter = rw
}

// func (rc *RequestCtx) parseForm() {
// 	rc.formOnce.Do(func() {
// 		rc.Request.ParseForm()
// 	})
// }

// func (rc *RequestCtx) parseMultipartForm(mm int64) {
// 	rc.multiformOnce.Do(func() {
// 		rc.Request.ParseMultipartForm(mm)
// 	})
// }

// func (rc *RequestCtx) Form() url.Values {
// 	rc.parseForm()
// 	return rc.Request.Form
// }

// func (rc *RequestCtx) PostForm() url.Values {
// 	rc.parseForm()
// 	return rc.Request.PostForm
// }

// func (rc *RequestCtx) MultipartForm(maxMemory int64) *multipart.Form {
// 	rc.parseMultipartForm(maxMemory)
// 	return rc.Request.MultipartForm
// }

// func (rc *RequestCtx) ResponseHeader() http.Header {
// 	return rc.ResponseWriter.Header()
// }

func (rc *RequestCtx) Reset() {
	rc.Request = nil
	rc.ResponseWriter = nil
}
