package caddy_saml_sso

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/crewjam/saml/samlsp"
)

const version = "0.0.1"

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("saml_sso", parseCaddyfile)
}

// Holds all the module's data
type Middleware struct {
	SamlIdpUrl   string `json:"saml_idp_url,omitempty"`
	SamlCertFile string `json:"saml_cert_file,omitempty"`
	SamlKeyFile  string `json:"saml_cert_key,omitempty"`
	SamlRootUrl  string `json:"saml_root_url,omitempty"`

	SamlSP      *samlsp.Middleware
	SamlHandler http.Handler
}

// CaddyModule returns the Caddy module information.
func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.saml_sso",
		New: func() caddy.Module { return new(Middleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *Middleware) Provision(ctx caddy.Context) error {
	keyPair, err := tls.LoadX509KeyPair(m.SamlCertFile, m.SamlKeyFile)
	if err != nil {
		return err
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return err
	}

	idpMetadataURL, err := url.Parse(m.SamlIdpUrl)
	if err != nil {
		return err
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient,
		*idpMetadataURL)
	if err != nil {
		return err
	}

	rootURL, err := url.Parse(m.SamlRootUrl)
	if err != nil {
		return err
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	if err != nil {
		return err
	}

	m.SamlSP = samlSP
	nullHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	m.SamlHandler = samlSP.RequireAccount(nullHandler)

	log("loaded saml_sso v%s", version)
	return nil
}

// Validate implements caddy.Validator.
func (m *Middleware) Validate() error {
	// TODO
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	log("middleware path=%s", r.URL.Path)

	// If the request is part of the SAML flow,
	// handle the request with the SAML library
	if strings.HasPrefix(r.URL.Path, "/saml") {
		m.SamlSP.ServeHTTP(w, r)
		return nil
	} else {
		// before going down the middleware stack, make sure
		// we are in a SAML session
		m.SamlHandler.ServeHTTP(w, r)

		// Let's grab the SAML session attributes and add them to the header
		// so other services can use it
		attributes, err := m.extractAttributes(r)
		if attributes != nil && err == nil {
			log("number of attributes=%d", len(attributes))
			for k, v := range attributes {
				if len(v) == 1 {
					if w.Header().Get(k) == "" {
						w.Header().Add(k, v[0])
					}
				}
			}
		} else {
			log("attributes=%v err=%s", attributes, err)
		}
		log("saml_sso v%s middlware done", version)
		return next.ServeHTTP(w, r)
	}
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

func (m *Middleware) extractAttributes(r *http.Request) (samlsp.Attributes, error) {
	session, _ := m.SamlSP.Session.GetSession(r)
	if session == nil {
		return nil, nil
	}

	r = r.WithContext(samlsp.ContextWithSession(r.Context(), session))
	jwtSessionClaims, ok := session.(samlsp.JWTSessionClaims)
	if !ok {
		return nil, fmt.Errorf("Unable to decode session into JWTSessionClaims")
	}

	return jwtSessionClaims.Attributes, nil
}

func log(msg string, args ...interface{}) {
	caddy.Log().Sugar().Infof(msg, args)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
