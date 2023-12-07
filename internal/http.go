package internal

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

var _ types.HttpContext = (*httpContext)(nil)

type httpContext struct {
	types.DefaultHttpContext

	configuration *pluginConfiguration
}

func (c *httpContext) OnHttpRequestHeaders(_ int, _ bool) types.Action {
	cookieHeader, err := proxywasm.GetHttpRequestHeader("cookie")
	if err != nil {
		if err == types.ErrorStatusNotFound {
			return types.ActionContinue
		}

		setErrorHTTPResponseWithLog("failed to get the cookie header from the http request: %s", err)
		return types.ActionPause
	}

	cookies := make([]*cookie, 0, strings.Count(cookieHeader, ";"))

	cookieHeader = textproto.TrimString(cookieHeader)

	var part string
	for len(cookieHeader) > 0 {
		part, cookieHeader, _ = strings.Cut(cookieHeader, ";")
		part = textproto.TrimString(part)
		if part == "" {
			continue
		}
		name, val, _ := strings.Cut(part, "=")
		name = textproto.TrimString(name)
		if !isCookieNameValid(name) {
			continue
		}
		val, ok := parseCookieValue(val, true)
		if !ok {
			continue
		}
		cookies = append(cookies, &cookie{name: name, value: val})
	}

	if len(cookies) == 0 {
		return types.ActionContinue
	}

	for _, co := range cookies {
		for _, cr := range c.configuration.Rules {
			if co.name == cr.CookieName {
				val := co.value
				if cr.HeaderValuePrefix != "" {
					val = fmt.Sprintf("%s%s", cr.HeaderValuePrefix, co.value)
				}

				if err := proxywasm.ReplaceHttpRequestHeader(cr.HeaderName, val); err != nil {
					setErrorHTTPResponseWithLog("failed to set the new header: %s", err)
					return types.ActionPause
				}
			}
		}
	}

	return types.ActionContinue
}

func setErrorHTTPResponseWithLog(format string, args ...interface{}) {
	proxywasm.LogErrorf(format, args...)
	if err := proxywasm.SendHttpResponse(500, nil, []byte(`{"error": "internal server error"}`), -1); err != nil {
		proxywasm.LogErrorf("failed to set the http error response: %s", err)
	}
}

type cookie struct {
	name  string
	value string
}

func isCookieNameValid(raw string) bool {
	if raw == "" {
		return false
	}
	return strings.IndexFunc(raw, isNotToken) < 0
}

func isNotToken(r rune) bool {
	return !isTokenRune(r)
}

func isTokenRune(r rune) bool {
	i := int(r)
	return i < len(isTokenTable) && isTokenTable[i]
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

var isTokenTable = [127]bool{
	'!':  true,
	'#':  true,
	'$':  true,
	'%':  true,
	'&':  true,
	'\'': true,
	'*':  true,
	'+':  true,
	'-':  true,
	'.':  true,
	'0':  true,
	'1':  true,
	'2':  true,
	'3':  true,
	'4':  true,
	'5':  true,
	'6':  true,
	'7':  true,
	'8':  true,
	'9':  true,
	'A':  true,
	'B':  true,
	'C':  true,
	'D':  true,
	'E':  true,
	'F':  true,
	'G':  true,
	'H':  true,
	'I':  true,
	'J':  true,
	'K':  true,
	'L':  true,
	'M':  true,
	'N':  true,
	'O':  true,
	'P':  true,
	'Q':  true,
	'R':  true,
	'S':  true,
	'T':  true,
	'U':  true,
	'W':  true,
	'V':  true,
	'X':  true,
	'Y':  true,
	'Z':  true,
	'^':  true,
	'_':  true,
	'`':  true,
	'a':  true,
	'b':  true,
	'c':  true,
	'd':  true,
	'e':  true,
	'f':  true,
	'g':  true,
	'h':  true,
	'i':  true,
	'j':  true,
	'k':  true,
	'l':  true,
	'm':  true,
	'n':  true,
	'o':  true,
	'p':  true,
	'q':  true,
	'r':  true,
	's':  true,
	't':  true,
	'u':  true,
	'v':  true,
	'w':  true,
	'x':  true,
	'y':  true,
	'z':  true,
	'|':  true,
	'~':  true,
}
