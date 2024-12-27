use proxy_wasm::traits::{Context, HttpContext, RootContext};
use proxy_wasm::types::{Action, ContextType, LogLevel};
use serde::{Deserialize, Serialize};

#[cfg(not(all(target_arch = "wasm32", target_os = "unknown")))]

proxy_wasm::main! {{
    proxy_wasm::set_log_level(LogLevel::Trace);
        proxy_wasm::set_root_context(|_| -> Box<dyn RootContext> { Box::new(CookieHTTPHeaderConverterRoot{
            config: Config { rules: vec![] }
        })
    });
}}

struct CookieHTTPHeaderConverterRoot {
    config: Config,
}

impl Context for CookieHTTPHeaderConverterRoot {}

impl RootContext for CookieHTTPHeaderConverterRoot {
    fn get_type(&self) -> Option<ContextType> {
        Some(ContextType::HttpContext)
    }

    fn on_configure(&mut self, _: usize) -> bool {
        if let Some(config_bytes) = self.get_plugin_configuration() {
            self.config = serde_json::from_reader(config_bytes.as_slice()).unwrap();
        }
        true
    }

    fn create_http_context(&self, _: u32) -> Option<Box<dyn HttpContext>> {
        Some(Box::new(CookieHTTPConverter {
            config: self.config.clone(),
        }))
    }
}

struct CookieHTTPConverter {
    config: Config,
}

impl Context for CookieHTTPConverter {}

impl HttpContext for CookieHTTPConverter {
    fn on_http_request_headers(&mut self, _: usize, _: bool) -> Action {
        let cookie_header = self.get_http_request_header("cookie");
        if let Some(cookie_header) = cookie_header {
            let cookies: Vec<&str> = cookie_header.split(";").collect();
            for cookie in cookies {
                let cookie_key_value: Vec<&str> = cookie.split("=").collect();
                if cookie_key_value.len() == 2 {
                    let cookie_name = cookie_key_value[0].trim();
                    let cookie_value = cookie_key_value[1].trim();
                    for rule in &self.config.rules {
                        if rule.cookie_name == cookie_name {
                            let header_name = rule.header_name.clone();

                            let mut header_value = cookie_value.to_string();
                            if let Some(prefix) = &rule.header_value_prefix {
                                header_value = format!("{}{}", prefix, cookie_value);
                            }

                            self.set_http_request_header(&header_name, Some(&header_value));
                        }
                    }
                }
            }
        }

        Action::Continue
    }
}

#[derive(Serialize, Deserialize, Clone)]
struct Config {
    rules: Vec<CookieConversionRule>,
}

#[derive(Serialize, Deserialize, Clone)]
struct CookieConversionRule {
    cookie_name: String,
    header_name: String,
    header_value_prefix: Option<String>,
}
