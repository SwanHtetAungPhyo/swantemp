package closure

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

type JsonResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSONMe(ctx *Context, statusCode int, message string, data any) error {
	response := &JsonResponse{
		Status:  statusCode,
		Message: message,
		Data:    data,
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(statusCode)

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		errorResponse := &JsonResponse{
			Status:  fasthttp.StatusInternalServerError,
			Message: "Internal Server Error",
			Data:    nil,
		}
		errorJSON, _ := json.Marshal(errorResponse)
		ctx.Response.SetBody(errorJSON)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return err
	}
	ctx.Response.SetBody(jsonResponse)
	return nil
}

func Binder(ctx *Context, target any) error {
	body := ctx.PostBody()
	err := json.Unmarshal(body, target)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return fmt.Errorf("error in json binding %s", err.Error())
	}
	return nil
}
