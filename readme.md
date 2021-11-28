# domains resolving to 127.0.0.1

- vcap.me
- lvh.me
- localtest.me

## creating certificates for these domains

- [mkcert](https://github.com/FiloSottile/mkcert)

```shell
mkcert -install
mkcert '*.vcap.me' '*.lvh.me' '*.localtest.me'
```
