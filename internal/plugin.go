package internal

import (
	"errors"
	"fmt"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
	"github.com/tidwall/gjson"
)

var _ types.PluginContext = (*pluginContext)(nil)

type pluginContext struct {
	types.DefaultPluginContext

	configuration *pluginConfiguration
}

func (c *pluginContext) NewHttpContext(_ uint32) types.HttpContext {
	return &httpContext{
		configuration: c.configuration,
	}
}

func (c *pluginContext) OnPluginStart(_ int) types.OnPluginStartStatus {
	config, err := getPluginConfiguration()
	if err != nil {
		proxywasm.LogErrorf("failed to get the plugin configuration: %s", err)
		return types.OnPluginStartStatusFailed
	}

	c.configuration = config

	return types.OnPluginStartStatusOK
}

func getPluginConfiguration() (*pluginConfiguration, error) {
	config, err := proxywasm.GetPluginConfiguration()
	if err != nil {
		if err == types.ErrorStatusNotFound {
			return nil, errors.New("the plugin configuration is not found")
		}

		return nil, fmt.Errorf("failed to get the plugin configuration: %w", err)
	}

	if len(config) == 0 {
		return nil, errors.New("the plugin configuration is empty")
	}

	if !gjson.ValidBytes(config) {
		return nil, errors.New("the plugin configuration is not valid JSON")
	}

	jsonConfig := gjson.ParseBytes(config)

	rulesToConvertCookie := jsonConfig.Get("rules").Array()
	if len(rulesToConvertCookie) == 0 {
		return nil, errors.New("the request headers to rename are not found")
	}

	rules := make([]*convertRules, len(rulesToConvertCookie))

	for i, r := range rulesToConvertCookie {
		c := r.Get("cookie_name").String()
		if c == "" {
			return nil, errors.New("the cookie name for converting is empty")
		}

		h := r.Get("header_name").String()
		if h == "" {
			return nil, errors.New("the header name for converting is empty")
		}

		p := r.Get("header_value_prefix").String()

		rules[i] = &convertRules{
			CookieName:        c,
			HeaderName:        h,
			HeaderValuePrefix: p,
		}
	}

	return &pluginConfiguration{
		Rules: rules,
	}, nil
}
