package caddy_saml_sso

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/crewjam/saml/samlsp"
)

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
