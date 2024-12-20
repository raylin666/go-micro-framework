package auth

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/raylin666/go-utils/auth"
	"google.golang.org/grpc/metadata"
	"mt/internal/constant/defined"
)

const (
	// AccessToken Headers 头权限认证参数名称
	AccessToken = "Access-Token"
	// AccessTokenID Context 上下文切换保存的权限认证ID名称
	AccessTokenID = "Access-Token-ID"

	// XMdGlobalJwtName Metadata 元数据传递保存的全局权限认证参数名称
	XMdGlobalJwtName = "x-md-global-jwt"
)

// NewJWTAuthServer JWT Server 中间件
func NewJWTAuthServer(jwt auth.JWT) func(handler middleware.Handler) middleware.Handler {
	return selector.Server(
		// JWT 权限验证
		JWTMiddlewareHandler(jwt),
	).Match(func(ctx context.Context, operation string) bool {
		// 路由白名单过滤 | 返回true表示需要处理权限验证, 返回false表示不需要处理权限验证
		return false
	}).Build()
}

// JWTMiddlewareHandler JWT 中间件处理器
func JWTMiddlewareHandler(jwt auth.JWT) func(handler middleware.Handler) middleware.Handler {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var jwtToken string
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				var jwtTokenSlice = md.Get(XMdGlobalJwtName)
				if len(jwtTokenSlice) <= 0 {
					return nil, defined.ErrorNotLogin
				}

				jwtToken = jwtTokenSlice[0]
			} else if header, ok := transport.FromServerContext(ctx); ok {
				jwtToken = header.RequestHeader().Get(AccessToken)
				if len(jwtToken) <= 0 {
					return nil, defined.ErrorNotLogin
				}
			} else {
				// 缺少可认证的 TOKEN，返回错误
				return nil, defined.ErrorNotLogin
			}
			jwtClaims, err := jwt.ParseToken(jwtToken)
			if err != nil {
				// 缺少合法的 TOKEN，返回错误
				return nil, defined.ErrorNotLogin
			}

			// 权限验证及保存 Token 到上下文切换
			var lenAud = len(jwtClaims.Audience)
			if lenAud <= 0 {
				return nil, defined.ErrorNotVisitAuth
			}
			for i := 0; i < lenAud; i++ {
				if jwtClaims.Audience[i] == jwtClaims.ID {
					ctx = context.WithValue(ctx, AccessTokenID, jwtClaims.ID)
					reply, err = handler(ctx, req)
					return
				}
			}

			return nil, defined.ErrorNotVisitAuth
		}
	}
}
