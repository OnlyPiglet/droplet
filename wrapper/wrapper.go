package wrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shiningrush/droplet"
	"github.com/shiningrush/droplet/core"
	"github.com/shiningrush/droplet/log"
	"github.com/shiningrush/droplet/middleware"
)

type HandleHttpInPipelineInput struct {
	Req            *http.Request
	RespWriter     core.ResponseWriter
	PathParamsFunc func(key string) string
	Handler        core.Handler
	Opts           []SetWrapOpt
}

func HandleHttpInPipeline(input HandleHttpInPipelineInput) {
	opt := &WrapOptBase{}
	for _, op := range input.Opts {
		op(opt)
	}

	dCtx := core.NewContext()
	dCtx.SetContext(input.Req.Context())

	trafficOpt := droplet.Option.TrafficLogOpt
	if trafficOpt == nil {
		trafficOpt = &opt.trafficLogOpt
	}

	ret, _ := core.NewPipe(droplet.Option.Orchestrator).
		Add(middleware.NewHttpInfoInjectorMiddleware(middleware.HttpInfoInjectorOption{
			ReqFunc: func() *http.Request {
				return input.Req
			},
			HeaderKeyRequestID: droplet.Option.HeaderKeyRequestID,
		})).
		Add(middleware.NewRespReshapeMiddleware(droplet.Option.ResponseNewFunc)).
		Add(middleware.NewHttpInputMiddleWare(middleware.HttpInputOption{
			PathParamsFunc:       input.PathParamsFunc,
			InputType:            opt.inputType,
			IsReadFromBody:       opt.isReadFromBody,
			DisableUnmarshalBody: opt.disableUnmarshalBody,
			Codecs:               droplet.Option.Codec,
		})).
		Add(middleware.NewTrafficLogMiddleware(trafficOpt)).
		SetOrchestrator(opt.orchestrator).
		Run(input.Handler, core.InitContext(dCtx))

	for k, _ := range dCtx.ResponseHeader() {
		input.RespWriter.SetHeader(k, dCtx.ResponseHeader().Get(k))
	}

	switch ret.(type) {
	case core.RawHttpResponse:
		rr := ret.(core.RawHttpResponse)
		if err := rr.WriteRawResponse(input.RespWriter); err != nil {
			logWriteErrors(input.Req, err)
		}
	case core.HttpFileResponse:
		fr := ret.(core.HttpFileResponse)
		fileResp := fr.Get()
		if fileResp.ContentType == "" {
			fileResp.ContentType = "application/octet-stream"
		}
		if fileResp.StatusCode > 0 {
			input.RespWriter.WriteHeader(fileResp.StatusCode)
		}
		input.RespWriter.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileResp.Name))
		input.RespWriter.SetHeader("Content-type", fileResp.ContentType)
		if fileResp.ContentReader != nil {
			defer fileResp.ContentReader.Close()
			_, err := io.Copy(input.RespWriter, fileResp.ContentReader)
			if err != nil {
				logWriteErrors(input.Req, err)
			}
		} else {
			_, err := input.RespWriter.Write(fileResp.Content)
			if err != nil {
				logWriteErrors(input.Req, err)
			}
		}
	case core.SpecCodeHttpResponse:
		resp := ret.(core.SpecCodeHttpResponse)
		if err := writeJsonToResp(input.RespWriter, resp.GetStatusCode(), resp); err != nil {
			logWriteErrors(input.Req, err)
		}
	default:
		if err := writeJsonToResp(input.RespWriter, http.StatusOK, ret); err != nil {
			logWriteErrors(input.Req, err)
		}
	}
}

func logWriteErrors(req *http.Request, err error) {
	log.Error("write resp failed",
		"err", err,
		"url", req.URL.String())
}

func writeJsonToResp(rw core.ResponseWriter, code int, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	rw.SetHeader("Content-Type", "application/json")

	rw.WriteHeader(code)
	if _, err := rw.Write(bs); err != nil {
		return fmt.Errorf("write to response failed: %w", err)
	}
	return nil
}
