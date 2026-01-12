// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SSOProvidersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewSSOProvidersLogic(r.Context(), svcCtx)
		resp, err := l.SSOProviders()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
