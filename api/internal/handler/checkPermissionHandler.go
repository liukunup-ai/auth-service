package handler

import (
	"net/http"

	"auth-service/api/internal/logic"
	"auth-service/api/internal/svc"
	"auth-service/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func CheckPermissionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckPermissionReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewCheckPermissionLogic(r.Context(), svcCtx)
		resp, err := l.CheckPermission(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
