

# Domains resolving to 127.0.0.1

- vcap.me
- lvh.me
- localtest.me

## Creating local certificates

- [mkcert](https://github.com/FiloSottile/mkcert)

```shell
mkcert -install
mkcert '*.vcap.me' '*.lvh.me' '*.localtest.me'
```
