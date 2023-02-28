# caddy-saml-sso

A caddy module that provides SSO via SAML. For the SAML implementation we use [this](https://github.com/crewjam/saml) library.

## Build

`make build` will caddy with the plugin.

`make saml-cert` will create a directory with the cert and key necessary to sign xml documents.

## Caddyfile example

```Caddy
(enable_saml) {
  saml_sso {
    saml_idp_url https://samltest.id/saml/idp
    saml_cert_file saml-cert/service.cert
    saml_key_file saml-cert/service.key
    saml_root_url https://foo.bar.net
  }
}

https://foo.bar.net:12000 {
  handle /ping {
    respond "pong"
  }

  handle /* {
    route /* {
      import enable_saml
      respond "ok"
    }
  }
}
```

In this Caddyfile we have a TLS server on `foo.bar.net:12000`.
The first handler `/ping` is not protected and we use it for testing.
The second handler handles the rest of the traffic. It loads the saml_sso
plugin and runs the middleware it provides.

The middleware passes SAML requests (`/saml`) to the SAML library. For other
paths, it runs the SAML middleware to make sure that each requests has a valid
SAML session. If not, it will redirect the user.

If all goes well, caddy will continue with the next middleware, in this case we
send an "ok" back. Here you probably want to redirect all traffic to your app
server.
