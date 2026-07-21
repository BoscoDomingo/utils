# Backend Security Headers

A portable baseline for browser-facing backends and web servers. Set headers centrally, then adapt them to the service's resource, framing, HTTPS, and privacy requirements.

## Recommended baseline

| Concern | Starting point |
| --- | --- |
| Content loading | `Content-Security-Policy` with same-origin defaults |
| HTTPS enforcement | `Strict-Transport-Security` after a gradual rollout |
| Framing | CSP `frame-ancestors` plus `X-Frame-Options` for older browsers |
| MIME sniffing | `X-Content-Type-Options: nosniff` |
| Referrer leakage | `strict-origin-when-cross-origin`, or `no-referrer` for sensitive services |

## Content Security Policy

Start with same-origin resources and add only required sources:

```http
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; frame-ancestors 'none'
```

Avoid `default-src *`, `'unsafe-inline'`, and `'unsafe-eval'`. Prefer self-hosted assets; use Subresource Integrity for unavoidable third-party scripts.

For existing applications, observe violations before enforcing:

```http
Content-Security-Policy-Report-Only: default-src 'self'; report-uri /csp-report
```

## Strict Transport Security

Begin with a short `max-age`, verify HTTPS operation, then increase it:

```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

Use `includeSubDomains` only when every subdomain supports HTTPS. For eligible public services, review and submit through `https://hstspreload.org` before using:

```http
Strict-Transport-Security: max-age=63072000; includeSubDomains; preload
```

Preloading is difficult to reverse and unsuitable for internal-only services. `max-age=0` disables HSTS.

## Framing

Use CSP `frame-ancestors 'none'`, or `'self'` when same-origin framing is required. For older browsers, add the matching header:

```http
X-Frame-Options: DENY
```

Use `SAMEORIGIN` instead when same-origin framing is required.

## MIME sniffing

```http
X-Content-Type-Options: nosniff
```

Also send an accurate `Content-Type` for every response.

## Referrer privacy

For typical public services:

```http
Referrer-Policy: strict-origin-when-cross-origin
```

For internal or sensitive services, use `no-referrer`. Use `same-origin` when same-origin referrers are needed. Avoid `unsafe-url`, which sends the full URL.

## Verify

- Check success, redirect, error, and static-asset responses.
- Test CSP in report-only mode before enforcement.
- Confirm framing restrictions and HTTP-to-HTTPS redirects.
- Re-check policies when frontend dependencies change.
