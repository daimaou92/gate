package gate

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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

type RequestCtx struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}

// Must happen after payload unmarshal
func (rc *RequestCtx) update(rw http.ResponseWriter, r *http.Request) {
	rc.Request = r
	rc.ResponseWriter = rw
}

func (rc *RequestCtx) Reset() {
	rc.Request = nil
	rc.ResponseWriter = nil
}

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
