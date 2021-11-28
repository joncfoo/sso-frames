.PHONY: init
init:
	mkcert -install
	mkcert -cert-file cert.pem -key-file cert-key.pem '*.lvh.me' '*.localtest.me'
