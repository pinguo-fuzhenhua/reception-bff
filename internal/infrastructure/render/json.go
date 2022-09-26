package render

import (
	"encoding/json"
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var MarshalOptions = protojson.MarshalOptions{
	UseEnumNumbers:  true,
	EmitUnpopulated: true,
}

type ErrorJSON struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func RenderJSON(ctx khttp.Context, data interface{}) (err error) {
	var buf []byte
	if pbmsg, ok := data.(proto.Message); ok {
		buf, err = MarshalOptions.Marshal(pbmsg)
	} else {
		buf, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}

	w := ctx.Response()
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(200)
	_, err = w.Write(buf)
	return
}

func LoginRequired(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(403)
	err, _ := json.Marshal(&ErrorJSON{
		Code:    403,
		Message: "Login Required",
	})
	w.Write(err)
}
