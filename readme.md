# tl;dr

To get SSO working in your embedded site:

1. Set `SameSite=None` and `Secure` on your cookies (only possible over HTTPS)
2. See the JavaScript code in the following files
    - `static/embed.html`
    - `templates/dashboard-close-login.html`
    - `templates/explicit-store-access.html`

* You *must* use a new login window (no getting around this)

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

# SSO

Quite a few SSO providers do not allow embedding their login pages
in iframes.  While some can be configured to do so, it is best to
perform login via the SSO provider in a new window/tab as embedding
the login window looks phishy!

# SameSite cookie attribute

See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite

## Lax

This is the default if `SameSite` isn't set.

iframe embedding only works in Chrome-based browsers.

> Cookies are not sent on normal cross-site subrequests (for example
> to load images or frames into a third party site), but are sent when
> a user is navigating to the origin site (i.e., when following a link).

## Strict

Use this if you never want to allow embedding a site in an iframe.

> Cookies will only be sent in a first-party context and not be sent
> along with requests initiated by third party websites.

## None

iframe embedding can be made to work in all browsers with the help of
the Storage Access API.

Safari and Firefox take extra measures to ensure a user's privacy by
denying third-party cookies by default.

> Cookies will be sent in all contexts, i.e. in responses to both
> first-party and cross-origin requests. If SameSite=None is set,
> the cookie Secure attribute must also be set (or the cookie will
> be blocked).

# Storage Access API

`document.requestStorageAccess()` *must* be called from an event handler
triggered by a user.

- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite
- https://webkit.org/blog/10218/full-third-party-cookie-blocking-and-more/
- https://webkit.org/blog/8124/introducing-storage-access-api/
- https://webkit.org/blog/11545/updates-to-the-storage-access-api/
