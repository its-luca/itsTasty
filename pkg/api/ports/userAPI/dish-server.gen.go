// Package userAPI provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.1-0.20220812203637-fec990c8f823 DO NOT EDIT.
package userAPI

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /dishes/{dishID})
	GetDishesDishID(w http.ResponseWriter, r *http.Request, dishID int64)

	// (POST /dishes/{dishID})
	PostDishesDishID(w http.ResponseWriter, r *http.Request, dishID int64)

	// (GET /getAllDishes)
	GetGetAllDishes(w http.ResponseWriter, r *http.Request)

	// (POST /searchDish)
	PostSearchDish(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetDishesDishID operation middleware
func (siw *ServerInterfaceWrapper) GetDishesDishID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "dishID" -------------
	var dishID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "dishID", runtime.ParamLocationPath, chi.URLParam(r, "dishID"), &dishID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "dishID", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetDishesDishID(w, r, dishID)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// PostDishesDishID operation middleware
func (siw *ServerInterfaceWrapper) PostDishesDishID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "dishID" -------------
	var dishID int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "dishID", runtime.ParamLocationPath, chi.URLParam(r, "dishID"), &dishID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "dishID", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostDishesDishID(w, r, dishID)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetGetAllDishes operation middleware
func (siw *ServerInterfaceWrapper) GetGetAllDishes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetGetAllDishes(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// PostSearchDish operation middleware
func (siw *ServerInterfaceWrapper) PostSearchDish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PostSearchDish(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/dishes/{dishID}", wrapper.GetDishesDishID)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/dishes/{dishID}", wrapper.PostDishesDishID)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/getAllDishes", wrapper.GetGetAllDishes)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/searchDish", wrapper.PostSearchDish)
	})

	return r
}

type GetDishesDishIDRequestObject struct {
	DishID int64 `json:"dishID"`
}

type GetDishesDishID200JSONResponse GetDishResp

func (t GetDishesDishID200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((GetDishResp)(t))
}

type GetDishesDishID400JSONResponse BasicError

func (t GetDishesDishID400JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

type GetDishesDishID401Response struct {
}

type GetDishesDishID404Response struct {
}

type GetDishesDishID500JSONResponse BasicError

func (t GetDishesDishID500JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

type PostDishesDishIDRequestObject struct {
	DishID int64 `json:"dishID"`
	Body   *PostDishesDishIDJSONRequestBody
}

type PostDishesDishID200Response struct {
}

type PostDishesDishID400JSONResponse BasicError

func (t PostDishesDishID400JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

type PostDishesDishID401Response struct {
}

type PostDishesDishID404Response struct {
}

type PostDishesDishID500JSONResponse BasicError

func (t PostDishesDishID500JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

type GetGetAllDishesRequestObject struct {
}

type GetGetAllDishes200JSONResponse GetAllDishesResponse

func (t GetGetAllDishes200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((GetAllDishesResponse)(t))
}

type GetGetAllDishes401Response struct {
}

type GetGetAllDishes500JSONResponse BasicError

func (t GetGetAllDishes500JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

type PostSearchDishRequestObject struct {
	Body *PostSearchDishJSONRequestBody
}

type PostSearchDish200JSONResponse SearchDishResp

func (t PostSearchDish200JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((SearchDishResp)(t))
}

type PostSearchDish401Response struct {
}

type PostSearchDish500JSONResponse BasicError

func (t PostSearchDish500JSONResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((BasicError)(t))
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /dishes/{dishID})
	GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) interface{}

	// (POST /dishes/{dishID})
	PostDishesDishID(ctx context.Context, request PostDishesDishIDRequestObject) interface{}

	// (GET /getAllDishes)
	GetGetAllDishes(ctx context.Context, request GetGetAllDishesRequestObject) interface{}

	// (POST /searchDish)
	PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) interface{}
}

type StrictHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) interface{}

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// GetDishesDishID operation middleware
func (sh *strictHandler) GetDishesDishID(w http.ResponseWriter, r *http.Request, dishID int64) {
	var request GetDishesDishIDRequestObject

	request.DishID = dishID

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.GetDishesDishID(ctx, request.(GetDishesDishIDRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetDishesDishID")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case GetDishesDishID200JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		writeJSON(w, v)
	case GetDishesDishID400JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		writeJSON(w, v)
	case GetDishesDishID401Response:
		w.WriteHeader(401)
	case GetDishesDishID404Response:
		w.WriteHeader(404)
	case GetDishesDishID500JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		writeJSON(w, v)
	case error:
		sh.options.ResponseErrorHandlerFunc(w, r, v)
	case nil:
	default:
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", v))
	}
}

// PostDishesDishID operation middleware
func (sh *strictHandler) PostDishesDishID(w http.ResponseWriter, r *http.Request, dishID int64) {
	var request PostDishesDishIDRequestObject

	request.DishID = dishID

	var body PostDishesDishIDJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.PostDishesDishID(ctx, request.(PostDishesDishIDRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostDishesDishID")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case PostDishesDishID200Response:
		w.WriteHeader(200)
	case PostDishesDishID400JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		writeJSON(w, v)
	case PostDishesDishID401Response:
		w.WriteHeader(401)
	case PostDishesDishID404Response:
		w.WriteHeader(404)
	case PostDishesDishID500JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		writeJSON(w, v)
	case error:
		sh.options.ResponseErrorHandlerFunc(w, r, v)
	case nil:
	default:
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", v))
	}
}

// GetGetAllDishes operation middleware
func (sh *strictHandler) GetGetAllDishes(w http.ResponseWriter, r *http.Request) {
	var request GetGetAllDishesRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.GetGetAllDishes(ctx, request.(GetGetAllDishesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetGetAllDishes")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case GetGetAllDishes200JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		writeJSON(w, v)
	case GetGetAllDishes401Response:
		w.WriteHeader(401)
	case GetGetAllDishes500JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		writeJSON(w, v)
	case error:
		sh.options.ResponseErrorHandlerFunc(w, r, v)
	case nil:
	default:
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", v))
	}
}

// PostSearchDish operation middleware
func (sh *strictHandler) PostSearchDish(w http.ResponseWriter, r *http.Request) {
	var request PostSearchDishRequestObject

	var body PostSearchDishJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) interface{} {
		return sh.ssi.PostSearchDish(ctx, request.(PostSearchDishRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostSearchDish")
	}

	response := handler(r.Context(), w, r, request)

	switch v := response.(type) {
	case PostSearchDish200JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		writeJSON(w, v)
	case PostSearchDish401Response:
		w.WriteHeader(401)
	case PostSearchDish500JSONResponse:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		writeJSON(w, v)
	case error:
		sh.options.ResponseErrorHandlerFunc(w, r, v)
	case nil:
	default:
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", v))
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintln(w, err)
	}
}

func writeRaw(w http.ResponseWriter, b []byte) {
	if _, err := w.Write(b); err != nil {
		fmt.Fprintln(w, err)
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+RYQY/bNhP9KwN+36EFBNubbIrAt01cpEabZLGbnoI90NLYYiKRCjmyoS7034shZUle",
	"0c0WTZoWPW0iiZzHmTdvHn0vUlNWRqMmJ5b3wqU5ltL/84V0Kv3RWmP5f5U1FVpS6N8dckn8l5oKxVI4",
	"skrvRNsmxydm8wFTEm0iXiFdFcVKuRzdDbrKaIe8NEOXWlWRMlosxZW1soGDohxkUcBHbQ4aMuVyWK+c",
	"SIQiLN10GX+xXolEbI0tGZJQmn64FD0OpQl3aMWATHKkDhijYkzTjVdIUhWYwegxmC1ID0okD/Ih97sb",
	"SZyD6cn2aOUOwfr3sDUWKFfO7zODt6UiwgzUFihHiyAtgjawN4QOGqThKLouN+EkWpaRFL6RJTJEyvEI",
	"8kF5EmHStLYWdYovTa0pgrbk534bVaIboEK3NIvmNhzu7fZXh3a66U3k6EekFj/V6Pz72qF9mBH/EHLp",
	"QBviHGLGWZmJRKCuS7F8f5E8SZ4ml8mzu/PAQo2yTDEgWVyf1G66KIbePazcz9g4KFHqrrIJ7GVR48kz",
	"FxhNuSSQfWYdSetmsN5CaSzyWw2/oT0WnRlQWXSoCXpewVZhkUFqNEmlnU+NPGHWTESaz2KKmt72VY+0",
	"0GvjCMJ3MNDDjZk0g9dql5MvQYfAvzvkpkDIlSNjm/NNGnoTmqZpZmU5y7Jxv2aSMMbUh93q0O4xu4pQ",
	"9heTSt+eB99AA8OUg7Bqur9PzadaMZ+X70NDTbsjlr+BUiNMd5HU30jCoC+fIh0RWA8Uiu65dUZabK8r",
	"jyD8g3N1a2PwblHaNH8MwMKYj3XVwVuvYNP0xIAucaeI+c2bqERxON7S+eB86ljpz5d6LHHFXy97D/Qz",
	"pRznKjYsXo67ssvSSNt4isQq2w2vyXbr1fGMIU/delZERXCQLEW1zgapNDw6Dsrh4wbh+fqM89vHZugt",
	"b1zrjHMwXfbO1njUa4+0xzjE3xhToNR/VIMhwrQIvEzprYkE58JfXa/ZN5iD8wPDMcl4WIDUGewVHjw2",
	"kjxo0HlxM7WFDRZmjxmUqF3gt+9wUlRw8PW7W/juJ1Phti6K5nt4Jx01wDOOA4pE7NG6gGIxu5gt/ISt",
	"UMtKiaV4OlvMnnLRJeW+3vMQYH4fCt/ysx1GaP4KCTJvQBwU6iP204RPM9bo04mqdFrUGQ+L4+B0wDaq",
	"UwGPzfqeWWchSnBlq6OJqqSVJRJaJ5bv74ViLIxeHC3H4LeGEpKtMel8Ix/lswRs73h5MII+L08WC/7D",
	"kwWDJ5FVVajQ3vMPjnNyP4rwf4tbsRT/mw/Wdd751vnY1HnSnLF1zCXGyQIiN6amwTa1ibj8gohGNjoC",
	"6IXMYK2rmiCTJEPsiykjPOc0YuaCIu+UDt9envPEfliHHmwT8exvO9BaE1otC0D+AjY1gfLn85qgNIZF",
	"lXEUNYo49q+nhL027lsy1mv5C5M1XyyTY4vQngojY2zjfXKasts6TdG5b8Xa1X+ItW0i5rvRbfasft8g",
	"1bazA+uVnzanl1o/ZSZqPL4pi6+rkdMbeSQp4Rix+/ifLPg/sI6ut3T+x42oGt32LrUzv2x9j1eFiTIN",
	"JlF8HbU4deyP14svHjw+Wk+E6N/LjeP9o5so06tmAdT9XhG+E4mobSGWIieqlvM530uK3DhaPl88X8zZ",
	"hl1dr+f7C9Hetb8HAAD///N+q0tvEwAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
