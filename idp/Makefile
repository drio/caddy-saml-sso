HOSTNAME=idp.magnetux.net
IDP_HOST=192.168.8.222
IDP_PORT=16000

.PHONY: deps run metadata

run:
	@echo "Remember to send your SP metadata to the IDP."
	@echo "Use:"
	@echo '> curl -X PUT -d @</path/to/metadata.xml> -H "Content-Type: application/xml"  http://<idp:port>/services/<sp_name>'
	@echo ''
	go run ./idp.go -idp http://$(HOSTNAME) -bind $(IDP_HOST):$(IDP_PORT)

deps:
	go get github.com/zenazn/goji
	go get github.com/crewjam/saml/samlidp	

metadata:
	curl $(IDP_HOST):$(IDP_PORT)/metadata
