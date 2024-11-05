package middleware

import (
	"errors"

	"github.com/OnlyPiglet/droplet/core"
	"github.com/OnlyPiglet/droplet/data"
)

type HttpRespReshapeOpt struct {
	DefaultErrCode int
}

type HttpRespReshapeMiddleware struct {
	BaseMiddleware

	opt         HttpRespReshapeOpt
	respNewFunc func() data.HttpResponse
}

func NewRespReshapeMiddleware(respNewFunc func() data.HttpResponse, opt HttpRespReshapeOpt) *HttpRespReshapeMiddleware {
	return &HttpRespReshapeMiddleware{
		opt:         opt,
		respNewFunc: respNewFunc,
	}
}

func (mw *HttpRespReshapeMiddleware) Handle(ctx core.Context) error {
	handlerErr := mw.BaseMiddleware.Handle(ctx)

	var resp data.HttpResponse
	switch t := ctx.Output().(type) {
	case data.RawHttpResponse, data.HttpFileResponse:
		return nil
	case data.HttpResponse:
		resp = t
	default:
		// wrap result
		resp = mw.respNewFunc()
		resp.Set(0, "", ctx.Output())
		ctx.SetOutput(resp)
	}

	resp.SetReqID(ctx.GetString(KeyRequestID))
	if handlerErr != nil {
		be := &data.BaseError{}
		if errors.As(handlerErr, &be) {
			resp.Set(be.Code, handlerErr.Error(), be.Data)
			return nil
		}

		errCode := data.ErrCodeInternal
		if mw.opt.DefaultErrCode != 0 {
			errCode = mw.opt.DefaultErrCode
		}

		resp.Set(errCode, handlerErr.Error(), nil)
	}

	// response reshape is the last step, so we don't need to return error
	return nil
}
