# Gate

---

### This is alpha grade software

---

Opinionated lib based on [httprouter](https://github.com/julienschmidt/httprouter) to build REST APIs.

## Usage

---

Using `Gate` is a little different than most other frameworks. Using it correctly
requires that one decide the request, response and query payload while designing
an API to the maximum extent feasible. When explicitly defined, as in the example
below, and later fetched using the `Gate` provided attribute - one is saving
on re-allocations.

`Gate` maintains a pool of all the structures explicitly provided when writing
the endpoints. These pools are used to reflect back a pre-allocated value,
when present, in the argument of type `RequestData` later received in the handler.

## Open API

---

Using `Gate` as intended lets us generate an OpenAPI based doc right from your code.
Work is underway to achieve this. This feature is inspired by [Dropshot](https://github.com/oxidecomputer/dropshot).

## Example

---

Here's how a POST request looks in Gate

```go
package main

import (
	"encoding/json"
	"log"

	"github.com/enalk-com/gate"
	"github.com/getkin/kin-openapi/openapi3"
)

/* gate provides a Payload interface that must be used
in API definitions and their handlers to access said data.
Since generics are still a little bit away the definitions
need to use a intialized instance of said payload.
Here's how all that looks*/

// Define a payload type
type StringJSON string

func (sj StringJSON) Marshal() ([]byte, error) {
	return json.Marshal(sj)
}
func (sj *StringJSON) Unmarshal(src []byte) error {
	var v string
	if err := json.Unmarshal(src, &v); err != nil {
		return err
	}
	*sj = StringJSON(v)
	return nil
}
func (StringJSON) ContentType() gate.ContentType {
	return gate.ContentTypeJSON
}
// The above three functions Marshal, Unmarshal and
// ContentType implement the gate.Payload interface

// This is just a helper function to get a
// StringJSON pointer - or a gate.Payload
func NewStringJSON(s string) *StringJSON {
	t := StringJSON(s)
	return &t
}

// below is a request handler that accepts a
// JSON string and responds back with a JSON
// string by prepending the pattern "YOLO"
func yoloHandler(rc *gate.RequestCtx, rd *gate.RequestData) (gate.Payload, error) {
	sj, ok := rd.Body.(*StringJSON)
	if !ok {
		// returning an error automatically responds back with
		// the corresponding code of the error
		// Here for example, the client with receive the error code
		// 400 and a text message "Bad Request"
		return nil, gate.ErrBadRequest
	}
	sj = NewStringJSON("YOLO" + string(*sj))
	return sj, nil
}

func main() {
	// Now lets define the api server
	app, err := gate.New(gate.AppOptions{
		Info: openapi3.Info{
			Title:   "sampleAPI",
			Version: "0.0.1",
		},
		Addr: ":8080",
	})
	if err != nil {
		panic(err)
	}

	sj := NewStringJSON("")
	// The usage of `sj` below is only so that gate knows what
	// type to marshal and unmarshal the payloads into
	// the initialized value sj above serves no other purpose
	// at the moment. Maybe this'll be better with generics.
	// But this is where we're at now.
	app.Post(gate.EndpointConfig{
		Path:    "/yolofy",
		Handler: yoloHandler,
		Payload: gate.EndpointPayload{
			RequestPayload:  sj,
			ResponsePayload: sj,
		},
	})
	log.Println("Listening at :8080")
	if err := app.Listen(); err != nil {
		log.Fatal(err)
	}
}

```

Now simply use CURL to verify:
`curl -d '""' -H 'Content-Type: application/json' -X POST http://localhost:8080/yolofy`

You should see `"YOLO"` as output.

---

### This documentation like the whole project is also a WIP. It should be updated as soon as I have more time. Thank you for your patience 🙏

---
