package internal

type pluginConfiguration struct {
	Rules []*convertRules `json:"rules"`
}

type convertRules struct {
	CookieName        string `json:"cookie_name"`
	HeaderName        string `json:"header_name"`
	HeaderValuePrefix string `json:"header_value_prefix"`
}
