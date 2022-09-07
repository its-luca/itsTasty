// Package userAPI provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.1-0.20220906181851-9c600dddea33 DO NOT EDIT.
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

	// (GET /users/me)
	GetUsersMe(w http.ResponseWriter, r *http.Request)
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

// GetUsersMe operation middleware
func (siw *ServerInterfaceWrapper) GetUsersMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetUsersMe(w, r)
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
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/me", wrapper.GetUsersMe)
	})

	return r
}

type GetDishesDishIDRequestObject struct {
	DishID int64 `json:"dishID"`
}

type GetDishesDishIDResponseObject interface {
	VisitGetDishesDishIDResponse(w http.ResponseWriter) error
}

type GetDishesDishID200JSONResponse GetDishResp

func (response GetDishesDishID200JSONResponse) VisitGetDishesDishIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetDishesDishID400JSONResponse BasicError

func (response GetDishesDishID400JSONResponse) VisitGetDishesDishIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type GetDishesDishID401Response struct {
}

func (response GetDishesDishID401Response) VisitGetDishesDishIDResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type GetDishesDishID404Response struct {
}

func (response GetDishesDishID404Response) VisitGetDishesDishIDResponse(w http.ResponseWriter) error {
	w.WriteHeader(404)
	return nil
}

type GetDishesDishID500JSONResponse BasicError

func (response GetDishesDishID500JSONResponse) VisitGetDishesDishIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type PostDishesDishIDRequestObject struct {
	DishID int64 `json:"dishID"`
	Body   *PostDishesDishIDJSONRequestBody
}

type PostDishesDishIDResponseObject interface {
	VisitPostDishesDishIDResponse(w http.ResponseWriter) error
}

type PostDishesDishID200Response struct {
}

func (response PostDishesDishID200Response) VisitPostDishesDishIDResponse(w http.ResponseWriter) error {
	w.WriteHeader(200)
	return nil
}

type PostDishesDishID400JSONResponse BasicError

func (response PostDishesDishID400JSONResponse) VisitPostDishesDishIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)

	return json.NewEncoder(w).Encode(response)
}

type PostDishesDishID401Response struct {
}

func (response PostDishesDishID401Response) VisitPostDishesDishIDResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type PostDishesDishID404Response struct {
}

func (response PostDishesDishID404Response) VisitPostDishesDishIDResponse(w http.ResponseWriter) error {
	w.WriteHeader(404)
	return nil
}

type PostDishesDishID500JSONResponse BasicError

func (response PostDishesDishID500JSONResponse) VisitPostDishesDishIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetGetAllDishesRequestObject struct {
}

type GetGetAllDishesResponseObject interface {
	VisitGetGetAllDishesResponse(w http.ResponseWriter) error
}

type GetGetAllDishes200JSONResponse GetAllDishesResponse

func (response GetGetAllDishes200JSONResponse) VisitGetGetAllDishesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetGetAllDishes401Response struct {
}

func (response GetGetAllDishes401Response) VisitGetGetAllDishesResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type GetGetAllDishes500JSONResponse BasicError

func (response GetGetAllDishes500JSONResponse) VisitGetGetAllDishesResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type PostSearchDishRequestObject struct {
	Body *PostSearchDishJSONRequestBody
}

type PostSearchDishResponseObject interface {
	VisitPostSearchDishResponse(w http.ResponseWriter) error
}

type PostSearchDish200JSONResponse SearchDishResp

func (response PostSearchDish200JSONResponse) VisitPostSearchDishResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type PostSearchDish401Response struct {
}

func (response PostSearchDish401Response) VisitPostSearchDishResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type PostSearchDish500JSONResponse BasicError

func (response PostSearchDish500JSONResponse) VisitPostSearchDishResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

type GetUsersMeRequestObject struct {
}

type GetUsersMeResponseObject interface {
	VisitGetUsersMeResponse(w http.ResponseWriter) error
}

type GetUsersMe200JSONResponse GetUsersMeResp

func (response GetUsersMe200JSONResponse) VisitGetUsersMeResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type GetUsersMe401Response struct {
}

func (response GetUsersMe401Response) VisitGetUsersMeResponse(w http.ResponseWriter) error {
	w.WriteHeader(401)
	return nil
}

type GetUsersMe500JSONResponse BasicError

func (response GetUsersMe500JSONResponse) VisitGetUsersMeResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)

	return json.NewEncoder(w).Encode(response)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /dishes/{dishID})
	GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) (GetDishesDishIDResponseObject, error)

	// (POST /dishes/{dishID})
	PostDishesDishID(ctx context.Context, request PostDishesDishIDRequestObject) (PostDishesDishIDResponseObject, error)

	// (GET /getAllDishes)
	GetGetAllDishes(ctx context.Context, request GetGetAllDishesRequestObject) (GetGetAllDishesResponseObject, error)

	// (POST /searchDish)
	PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) (PostSearchDishResponseObject, error)

	// (GET /users/me)
	GetUsersMe(ctx context.Context, request GetUsersMeRequestObject) (GetUsersMeResponseObject, error)
}

type StrictHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error)

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

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetDishesDishID(ctx, request.(GetDishesDishIDRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetDishesDishID")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetDishesDishIDResponseObject); ok {
		if err := validResponse.VisitGetDishesDishIDResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
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

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.PostDishesDishID(ctx, request.(PostDishesDishIDRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostDishesDishID")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(PostDishesDishIDResponseObject); ok {
		if err := validResponse.VisitPostDishesDishIDResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// GetGetAllDishes operation middleware
func (sh *strictHandler) GetGetAllDishes(w http.ResponseWriter, r *http.Request) {
	var request GetGetAllDishesRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetGetAllDishes(ctx, request.(GetGetAllDishesRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetGetAllDishes")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetGetAllDishesResponseObject); ok {
		if err := validResponse.VisitGetGetAllDishesResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
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

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.PostSearchDish(ctx, request.(PostSearchDishRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "PostSearchDish")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(PostSearchDishResponseObject); ok {
		if err := validResponse.VisitPostSearchDishResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// GetUsersMe operation middleware
func (sh *strictHandler) GetUsersMe(w http.ResponseWriter, r *http.Request) {
	var request GetUsersMeRequestObject

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetUsersMe(ctx, request.(GetUsersMeRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetUsersMe")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetUsersMeResponseObject); ok {
		if err := validResponse.VisitGetUsersMeResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+RYUW/bRhL+K4O9e7gDCElOnEOgNyc6pEKbxLDTvgR+WJEjcRNyl9kZSlAN/fdidimJ",
	"MleNiyZNgz7JJrncb2a++eZb3qvc1Y2zaJnU9F5RXmKtw58vNJn8/947L/813jXo2WC4tyk1yy9vG1RT",
	"ReyNXandLttfcYsPmLPaZeoV8lVVzQyVSDdIjbOEsrRAyr1p2DirpurKe72FjeESdFXBR+s2FgpDJcxn",
	"pDJlGGsaLpMn5jOVqaXztUBSxvL/LtUBh7GMK/TqiEzLTh0wQSWYhi+eIWtTYQG9y+CWoAMolT3Ih16v",
	"bjRLDoaRrdHrFYIP92HpPHBpKLxnBG9rw4wFmCVwiR5BewTrYO0YCbbIx1BsWy9iJFbXiRS+0TUKRC5x",
	"D/JBeTLl8rz1Hm2OL11rOYG2luvhNaZGOkKFbmmRzG0M7u3yZ0I/fOlNIvQ9Uo+fWqRwvyX0DzMSLkKp",
	"CaxjySEWkpWRyhTatlbT9xfZk+xpdpk9uzsPLNaoKIwA0tX1Se2Gi1Lo6WHlfsQtQY3adpXNYK2rFk+u",
	"UWQ0l5pBHzJLrD2NYL6E2nmUuxZ+Rb8vujCg8UhoGQ68gqXBqoDcWdbGUkiNPmHWSCWaz2OOlt8eqp5o",
	"odeOGOJzcKQH9Zk0gtdmVXIoQYcg3NuUrkIoDbHz2/NNGnsTttvtdlTXo6Lo92uhGVNMfdithH6NxVWC",
	"sj+5XIf23IQGOjLMEMRVw/eH1HxqjfB5+j421LA7Uvk7UqqH6S6te9IN9BrTCjO3MQeCXC9cy6l2GAgN",
	"1tpUaeXtBxQfS8G60YxR9j4lGjXuDhy5GCh/RvH8Qe4e0YcP0HVrU/BuUfu8fAzAyrmPbdPBm89gsT3w",
	"Fbp6niKWO2+SyinbySspbC5Rpxh5noF95a3+PBsPQD/DsH6uUgx72ReLLks9jslwS1W2m6lDws72McY8",
	"detFqA3DRotCtrY4KriTibYxhI+bz+fr08/vYW+BvpMXt7aQHAyXvfMt7sdIQHrAeNx/4VyF2v5eDY47",
	"DIsgy4xdusTmUvir67nYGbeh0M0kJJMZBtoWsDa4CdhYS8MjBc11rYcFVm6NBdRoKfI7CA8brmTz+btb",
	"+M8PrsFlW1Xb/8I7TbwFERvZUGVqjZ4iisnoYjQJg79BqxujpurpaDJ6KkXXXIZ6j+MG4/tY+J1cW2GC",
	"5q+QoQi+iKAyH/Ew5CSa/ug4HfTG5lVbiKLt5zmBuLtOBQI2H3pmXsRdolmc7b1do72ukdGTmr6/V0aw",
	"CHq1d0JHG3gsIfsWs87OSiifJeDuTpZHfxry8mQykR8ZeBitkm6aysT2Hn8gycl9b4d/e1yqqfrX+Oio",
	"x52dHve9ZiDNGbdpkkOh6Lh++QUR9dx9AtALXcDcNi1DoVnHvS+GjAics4gFRUVeGRufvTxn1YOHiD24",
	"y9SzvyyguWX0VleA8gQsWgYT4guaYCzGRY0jTvpX7NvqU8JeO/qWjA1a/sIV2y+Wyb5F2J0Ko2Dcpfvk",
	"NGW3bZ4j0bdi7ewfxNpdpsar3iH7rH7fILe+swPzWZg2p2ftMGUGatw/wKuvq5HDDwWJpMQwUp8J/mDB",
	"/4Z1pIOlC99ckmp0e3CpnfkV67s/wQyU6WgS1ddRi1PH/ni9+OKbp0friRB919wIzmkc/fFZf5Y2EOEj",
	"SuGiBzO0PwEk2707tH7lTu8fjRNJ+UVXpgBCEh87gq7lJbguqnA4/s5ruj9Tdi5h+FWjAu6+BcTnVKZa",
	"X6mpKpmb6XgsZ82qdMTT55Pnk0CQq+v5eH2hdne73wIAAP//MYBEKtoVAAA=",
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
