// This provides wrappers around http.ServeMux with a variety of convenience functions.
//
// The most significant component is replacing HandleFunc(func(http.ResponseWriter, *http.Request))
// with `HandleFunc(func(*Nu))`, which is both more concise and provides convenience methods for
// sending json replies, plaintext replies, error messages, forwarding headers, and CORS headers.
package numux

import (
	"github.com/gmalmquist/flowershop/regu"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Wraps an HTTP request.
//
// Contains a `http.ResponseWriter` and `*http.Request` which may be extracted with Unwrap().
type Nu struct {
	mux *http.ServeMux
	w   http.ResponseWriter
	r   *http.Request
}

// Wraps an http.ServeMux.
type NuMux struct {
	*http.ServeMux              // the internal http.ServeMux
	StandardHeader  http.Header // headers to be sent on every http response
	AllowAllCors    bool        // allows all CORS requests. only enable this for public apis
	CorsCredentials bool        // whether to pass credentials along with cors requests
}

// Wraps an individual request in a Nu struct.
//
// This can be used to use Nu's convenience methods with a regular http.ServeMux.
func WrapRequest(
	mux *http.ServeMux,
	w http.ResponseWriter,
	r *http.Request,
) *Nu {
	return &Nu{
		mux: mux,
		w:   w,
		r:   r,
	}
}

func (mux *NuMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.ServeMux.ServeHTTP(w, r)
}

// Binds the given route (e.g. "GET /api/flowers") to the a handler which accepts a *Nu request.
func (mux *NuMux) HandleFunc(route string, f func(*Nu)) {
	HandleFunc(mux.ServeMux, route, func(nu *Nu) {
		for key, vals := range mux.StandardHeader {
			for _, val := range vals {
				nu.AddHeader(key, val)
			}
		}
		if mux.AllowAllCors {
			nu.AllowAllCors(mux.CorsCredentials)
		}
		f(nu)
	})
}

// Adds response headers which allow all CORS headers requested by the client.
func (nu *Nu) AllowAllCors(withCredentials bool) {
	w, r := nu.Unwrap()
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Add("Access-Control-Allow-Origin", origin)

	for _, method := range r.Header.Values("Access-Control-Request-Method") {
		w.Header().Add("Access-Control-Allow-Method", method)
	}

	for _, header := range r.Header.Values("Access-Control-Request-Headers") {
		w.Header().Add("Access-Control-Allow-Headers", header)
	}

	if origin != "*" && withCredentials {
		w.Header().Add("Access-Control-Allow-Credentials", "true")
	}

	nu.ExposeCurrentHeaders()
}

// Exposes all headers that have already been set in the response to CORS.
func (nu *Nu) ExposeCurrentHeaders() {
	exposeCorsKey := "Access-Control-Expose-Headers"
	already := map[string]bool{}
	for _, name := range nu.w.Header().Values(exposeCorsKey) {
		already[name] = true
	}
	expose := []string{}
	for key, _ := range nu.w.Header() {
		if already[key] {
			continue
		}
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "allow-") || strings.HasPrefix(lower, "access-") {
			continue
		}
		expose = append(expose, key)
	}
	for _, expose := range expose {
		nu.w.Header().Add(exposeCorsKey, expose)
	}
}

// Always expose these headers to CORS on every http request.
func (mux *NuMux) AlwaysExposeHeaders(names ...string) {
	for _, name := range names {
		mux.StandardHeader.Add("Access-Control-Expose-Headers", name)
	}
}

// Wraps an http.ServeMux.
func Wrap(mux *http.ServeMux) *NuMux {
	return &NuMux{
		ServeMux: mux,
		StandardHeader: http.Header(
			map[string][]string{},
		),
	}
}

// Creates a new NuMux by wrapping a new http.ServeMux.
func New() *NuMux {
	return Wrap(http.NewServeMux())
}

// Binds the given route (e.g. "GET /api/flowers") to the a handler which accepts a *Nu request.
func HandleFunc(
	mux *http.ServeMux,
	route string,
	f func(u *Nu),
) {
	mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		f(WrapRequest(mux, w, r))
	})
}

var reErrCode = regexp.MustCompile(`^(?P<code>\d+)(([:]\s*)|(\s+)).*`)

// Parses things that look like http error codes from the beginning of an error.
//
// This uses a regex that just looks for a number at the beginning of the string.
func ParseErrCode(err any) int {
	m := regu.RegMatch(reErrCode, fmt.Sprintf("%v", err))
	if m == nil {
		return 0
	}
	code, e := strconv.Atoi(m["code"])
	if e != nil {
		log.Printf("Failed to parse err code from %v: %v", err, e)
		return 0
	}
	return code
}

// Marshal the given value into JSON and send it to the ResponseWriter.
//
// Marshalling errors will be converted to HTTP errors.
func ReplyJson(w http.ResponseWriter, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal response: %q", err), 500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

// Send an plaintext error message to the ResponseWriter with the given http error code.
func ReplyErr(w http.ResponseWriter, code int, err any, args ...any) {
	if len(args) > 0 {
		err = fmt.Sprintf(err.(string), args...)
	}
	pcode := ParseErrCode(err)
	if pcode > 0 && pcode < code {
		code = pcode
	}
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%v", err)))
}

func ReplyHTML(w http.ResponseWriter, html string) {
  res := []byte(html)
  w.Header().Add("Content-Type", "text/html; charset=UTF-8")
  w.Header().Add("Content-Length", fmt.Sprintf("%v", len(res)))
  w.WriteHeader(200)
  w.Write(res)
}

// Reply with a <script> tag that executes a browser redirect
func ReplyHTMLRedirect(w http.ResponseWriter, uri string) {
  ReplyHTML(w, fmt.Sprintf(strings.TrimSpace(`
    <script>window.location.href = '%v';</script>
  `), uri))
}

// Send a plaintext error message with the given http error code.
func (u *Nu) ReplyErr(code int, err any, args ...any) {
	ReplyErr(u.w, code, err, args...)
}

// Marshal the given value into JSON and send it.
//
// Marshalling errors will be converted to HTTP errors.
func (u *Nu) ReplyJson(blob any) {
	ReplyJson(u.w, blob)
}

// Send a plaintext response.
func (u *Nu) ReplyPlaintext(text string) {
	u.w.Header().Add("Content-Type", "text/html; charset=utf-8")
	data := []byte(text)
	u.w.Header().Add("Content-Length", fmt.Sprintf("%v", len(data)))
	u.w.Write(data)
}

// Send a plaintext response.
func (u *Nu) ReplyHTML(html string) {
  ReplyHTML(u.w, html)
}

// Reply with a <script> tag that executes a browser redirect
func (u *Nu) ReplyHTMLRedirect(uri string) {
  ReplyHTMLRedirect(u.w, uri)
}

// Rather than sending an http error header, sends back formatted human-readable HTML.
func (u *Nu) ReplyHTMLErr(code int, err any) {
	u.ReplyHTML(fmt.Sprintf(`
    <div class="error-message">
      <div class="error-code">%v</div>
      <pre class="error-body">%v</pre>
    </div>
  `, code, err))
}

// Read the cookie from the request or return the default value.
func (u *Nu) CookieOr(key string, defaultValue string) string {
	value := defaultValue
	for _, c := range u.r.CookiesNamed(key) {
		if c.Value != "" {
			value = c.Value
		}
	}
	return value
}

// Formats the given error message and puts it in a `X-Error-Message` header.
func (u *Nu) SetErrHeader(err any, args ...any) string {
	msg := fmt.Sprintf("%v", err)
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	u.SetHeader("X-Error-Message", msg)
	return msg
}

// Adds the given header to both the request and the response.
func (u *Nu) AddHeader(key string, val string) {
	u.r.Header.Add(key, val)
	u.w.Header().Add(key, val)
}

// Sets the given header in both the request and the response.
func (u *Nu) SetHeader(key string, val string) {
	u.r.Header.Set(key, val)
	u.w.Header().Add(key, val)
}

// Sets the given cookie in both the request and the response.
func (u *Nu) SetCookie(cookie *http.Cookie) {
	u.r.AddCookie(cookie)
	http.SetCookie(u.w, cookie)
}

// Returns any error that's set in the requests's `X-Error-Message`.
func (u *Nu) GetErrHeader() string {
	return u.r.Header.Get("X-Error-Message")
}

// Returns a function that send an http redirect to the give route.
func (u *Nu) Forwarder(method string, route string, args ...any) func() {
	return func() {
		u.Forward(method, route, args...)
	}
}

// Sends an http redirect to the give route.
func (u *Nu) Forward(method string, path string, args ...any) {
	if len(args) > 0 {
		path = fmt.Sprintf(path, args...)
	}
	if method != "" {
		u.r.Method = method
	}
	if path != "" {
		u.r.URL.Path = path
	}
	switch method {
	case "HEAD":
		fallthrough
	case "OPTIONS":
		fallthrough
	case "GET":
		fallthrough
	case "DELETE":
		var b bytes.Buffer
		u.r.Body = io.NopCloser(&b)
		u.r.Header.Set("Content-Length", strconv.Itoa(0))
	}
	u.mux.ServeHTTP(u.w, u.r)
}

// Returns the underlying `http.ResponseWriter` and `*http.Request`.
func (u *Nu) Unwrap() (w http.ResponseWriter, r *http.Request) {
	return u.w, u.r
}
