package caddy_saml_sso

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	httpcaddyfile.RegisterHandlerDirective("saml_sso", parseCaddyfile)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		// token value
		parameter := d.Val()
		// rest of params
		args := d.RemainingArgs()
		switch parameter {
		case "saml_idp_url":
			if len(args) != 1 {
				return d.Err("invalid saml_idp_url")
			}
			m.SamlIdpUrl = args[0]
		case "saml_cert_file":
			if len(args) != 1 {
				return d.Err("invalid saml_cert_file")
			}
			m.SamlCertFile = args[0]
		case "saml_key_file":
			if len(args) != 1 {
				return d.Err("invalid saml_key_file")
			}
			m.SamlKeyFile = args[0]
		case "saml_root_url":
			if len(args) != 1 {
				return d.Err("invalid saml_root_url")
			}
			m.SamlRootUrl = args[0]
		default:
			//d.Err("Unknow cam parameter: " + parameter)
			log("skipping: %s %v", parameter, args)
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
