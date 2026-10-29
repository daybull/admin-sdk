package module

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// Principal, isteği yapan giriş yapmış kullanıcıdır.
type Principal struct {
	UserID       int64
	Email        string
	Name         string
	IsSuperadmin bool
	Permissions  map[string]bool
	IP           string
}

// Has, kullanıcının yetkiye sahip olup olmadığını söyler. Süper admin her şeye sahiptir.
func (p *Principal) Has(perm string) bool {
	if p == nil {
		return false
	}
	return p.IsSuperadmin || p.Permissions[perm]
}

type principalKey struct{}
type ipKey struct{}

// WithPrincipal, kullanıcıyı ctx'e ekler (çekirdek tarafından kullanılır).
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom, ctx'teki kullanıcıyı döner; giriş yoksa nil.
func PrincipalFrom(ctx context.Context) *Principal {
	p, _ := ctx.Value(principalKey{}).(*Principal)
	return p
}

// WithClientIP, istemci IP'sini ctx'e ekler (çekirdek tarafından kullanılır).
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey{}, ip)
}

// ClientIP, ctx'teki istemci IP'sini döner.
func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(ipKey{}).(string)
	return ip
}

// SecurityCookie, OpenAPI'deki oturum çerezi güvenlik şemasının adıdır.
const SecurityCookie = "cookieAuth"

// Protect, bir operasyonu giriş ve (verilirse) yetki şartına bağlar.
// Kullanım: huma.Register(api, module.Protect(api, op, "notes.read"), handler)
func Protect(api huma.API, op huma.Operation, perm string) huma.Operation {
	op.Security = []map[string][]string{{SecurityCookie: {}}}
	if perm != "" {
		op.Description += "\n\nGerekli yetki: `" + perm + "`"
		if op.Metadata == nil {
			op.Metadata = map[string]any{}
		}
		op.Metadata["permission"] = perm
	}
	op.Middlewares = append(op.Middlewares, requireMiddleware(api, perm))
	return op
}

func requireMiddleware(api huma.API, perm string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		p := PrincipalFrom(ctx.Context())
		if p == nil {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "giriş gerekli")
			return
		}
		if perm != "" && !p.Has(perm) {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden, "bu işlem için yetkiniz yok")
			return
		}
		next(ctx)
	}
}
