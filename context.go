package gate

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/goccy/go-json"
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

type Pagination struct {
	Page       int32 `json:"page"`
	ItemCount  int8  `json:"item_count"`
	TotalItems int32 `json:"total_items"`
	Pages      int32 `json:"pages"`
	HasNext    bool  `json:"has_next"`
}

func (p Pagination) Marshal() ([]byte, error) {
	bs, err := json.Marshal(p)
	if err != nil {
		return nil, wrapErr(err, "json marshal failed")
	}
	return bs, nil
}

func (p *Pagination) Unmarshal(src []byte) error {
	var t Pagination
	if err := json.Unmarshal(src, &t); err != nil {
		return wrapErr(err, "json unmarshal failed")
	}
	*p = t
	return nil
}

func (Pagination) ContentType() ContentType {
	return ContentTypeJSON
}

type RequestData struct {
	Params      httprouter.Params
	Body        Payload
	QueryParams Payload
	Custom      map[string]interface{}
}

type Handler func(*RequestCtx, *RequestData) (Payload, error)

// type StreamHandler func(*RequestCtx, io.WriteCloser) error

var rcPool sync.Pool

func init() {
	rcPool.New = func() interface{} {
		return new(RequestCtx)
	}
}

type ResponseWriter struct {
	rw         http.ResponseWriter
	written    bool
	statusCode int
	mu         sync.Mutex
}

func (rw *ResponseWriter) Write(bs []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.written = true
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	i, err := rw.rw.Write(bs)
	if err != nil {
		return 0, wrapErr(err)
	}
	return i, nil
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.rw.Header()
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.rw.WriteHeader(statusCode)
	rw.statusCode = statusCode
	rw.written = true
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		rw: w,
	}
}

type RequestCtx struct {
	Request        *http.Request
	ResponseWriter *ResponseWriter
}

// Must happen after payload unmarshal
func (rc *RequestCtx) update(rw http.ResponseWriter, r *http.Request) {
	rc.Request = r
	rc.ResponseWriter = &ResponseWriter{
		rw: rw,
	}
}

func (rc *RequestCtx) reset() {
	rc.Request = nil
	rc.ResponseWriter = nil
}

// Will return 0 until Write or Writeheader is called
func (rc *RequestCtx) StatusCode() int {
	return rc.ResponseWriter.statusCode
}

// Returns the underlying *http.Request.Context
func (rc *RequestCtx) Context() context.Context {
	return rc.Request.Context()
}

// Tries it's best to find the real IP of the client. The header precedence
// from highest to lowest is 'Forwarded' > 'X-Forwarded-For' > 'X-Real-IP'
func (rc *RequestCtx) IP() string {
	r := rc.Request
	var (
		forwarded     []string
		xforwardedfor []string
		xrealip       []string
	)
	splitByCommas := func(a []string) []string {
		var vs []string
		for _, v := range a {
			v = strings.ReplaceAll(v, " ", "")
			vs = append(vs, strings.Split(v, ",")...)
		}
		return vs
	}
	for h, vs := range r.Header {
		switch h {
		case http.CanonicalHeaderKey("forwarded"):
			forwarded = append(forwarded, vs...)
			forwarded = splitByCommas(forwarded)
		case http.CanonicalHeaderKey("x-forwarded-for"):
			xforwardedfor = append(xforwardedfor, vs...)
			xforwardedfor = splitByCommas(xforwardedfor)
		case http.CanonicalHeaderKey("x-real-ip"):
			xrealip = append(xrealip, vs...)
			xrealip = splitByCommas(xrealip)
		}
	}
	var ip string
	if len(forwarded) > 0 {
		log.Printf("Received forwarded:\n%v\nLen: %d\n", forwarded, len(forwarded))
		re := regexp.MustCompile(`for=[\[\]a-fA-F0-9:"\.]*;`)
		d := string(re.Find([]byte(forwarded[0])))
		ip = strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(string(d), "for=", ""),
					";", "",
				),
				"\"", "",
			),
			"]", "",
		)
	} else if len(xforwardedfor) > 0 {
		ip = xforwardedfor[0]
	} else if len(xrealip) > 0 {
		ip = xrealip[0]
	} else {
		ip = r.RemoteAddr
	}
	u, err := url.Parse("http://" + ip)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
