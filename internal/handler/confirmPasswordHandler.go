package handler

import (
	"net/http"

	"auth-service/internal/logic"
	"auth-service/internal/svc"
	"auth-service/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ConfirmPasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConfirmPasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewConfirmPasswordLogic(r.Context(), svcCtx)
		resp, err := l.ConfirmPassword(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
