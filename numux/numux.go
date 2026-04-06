package numux

import (
  "flowershop/regu"

  "bytes"
  "fmt"
  "io"
  "net/http"
  "regexp"
  "strings"
  "strconv"
  "log"
  "encoding/json"
)

type Nu struct {
  mux *http.ServeMux
  w http.ResponseWriter
  r *http.Request
}

type NuMux struct {
  *http.ServeMux
  StandardHeader http.Header
  AllowAllCors bool
  CorsCredentials bool
}

func WrapRequest(
  mux *http.ServeMux,
  w http.ResponseWriter,
  r *http.Request,
) *Nu {
  return &Nu{
    mux: mux,
    w: w,
    r: r,
  }
}

func (mux *NuMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  mux.ServeMux.ServeHTTP(w, r)
}

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

func (mux *NuMux) AlwaysExposeHeaders(names ...string) {
  for _, name := range names {
    mux.StandardHeader.Add("Access-Control-Expose-Headers", name)
  }
}

func Wrap(mux *http.ServeMux) *NuMux {
  return &NuMux{
    ServeMux: mux,
    StandardHeader: http.Header(
        map[string][]string{},
    ),
  }
}

func New() *NuMux {
  return Wrap(http.NewServeMux())
}

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

func ReplyErr(w http.ResponseWriter, code int, err any) {
  pcode := ParseErrCode(err)
  if pcode > 0 && pcode < code {
    code = pcode
  }
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%v", err)))
}

func (u *Nu) ReplyErr(code int, err any) {
  ReplyErr(u.w, code, err)
}

func (u *Nu) ReplyJson(blob any) {
  ReplyJson(u.w, blob)
}

func (u *Nu) ReplyHTMLErr(code int, err any) {
  u.w.Header().Add("Content-Type", "text/html; charset=utf-8")
  u.w.Write([]byte(fmt.Sprintf(`
    <div class="error-message">
      <div class="error-code">%v</div>
      <pre class="error-body">%v</pre>
    </div>
  `, code, err)))
}

func (u *Nu) CookieOr(key string, defaultValue string) string {
  value := defaultValue
  for _, c := range u.r.CookiesNamed(key) {
    if c.Value != "" {
      value = c.Value
    }
  }
  return value
}

func (u *Nu) SetErrHeader(err any, args ...any) string {
  msg := fmt.Sprintf("%v", err)
  if len(args) > 0 {
    msg = fmt.Sprintf(msg, args...)
  }
  u.SetHeader("X-Error-Message", msg)
  return msg
}

func (u *Nu) AddHeader(key string, val string) {
  u.r.Header.Add(key, val)
  u.w.Header().Add(key, val)
}

func (u *Nu) SetHeader(key string, val string) {
  u.r.Header.Set(key, val)
  u.w.Header().Add(key, val)
}

func (u *Nu) SetCookie(cookie *http.Cookie) {
  u.r.AddCookie(cookie)
  http.SetCookie(u.w, cookie)
}

func (u *Nu) GetErrHeader() string {
  return u.r.Header.Get("X-Error-Message")
}

func (u *Nu) Forwarder(method string, route string, args ...any) func() {
  return func() {
    u.Forward(method, route, args...)
  }
}

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
  case "HEAD": fallthrough
  case "OPTIONS": fallthrough
  case "GET": fallthrough
  case "DELETE":
    var b bytes.Buffer
    u.r.Body = io.NopCloser(&b)
    u.r.Header.Set("Content-Length", string(0))
  }
  u.mux.ServeHTTP(u.w, u.r)
}

func (u *Nu) Unwrap() (w http.ResponseWriter, r *http.Request) {
  return u.w, u.r
}

