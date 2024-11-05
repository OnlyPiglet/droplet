package gin

import (
	"github.com/OnlyPiglet/droplet/core"
	"github.com/OnlyPiglet/droplet/wrapper"
	"github.com/gin-gonic/gin"
)

func Wraps(handler core.Handler, opts ...wrapper.SetWrapOpt) func(*gin.Context) {
	return func(ctx *gin.Context) {
		wrapper.HandleHttpInPipeline(wrapper.HandleHttpInPipelineInput{
			Req:            ctx.Request,
			RespWriter:     wrapper.NewResponseWriter(ctx.Writer),
			PathParamsFunc: ctx.Param,
			Handler:        handler,
			Opts:           opts,
		})
	}
}
