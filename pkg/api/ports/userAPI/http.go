package userAPI

import (
	"context"
	"fmt"
)

//go:generate oapi-codegen --config ./server.cfg.yml ../../userAPI.yml
//go:generate oapi-codegen --config ./types.cfg.yml ../../userAPI.yml

type contextKey string

const userEmailContextKey = contextKey("userEmail")

func ContextWithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailContextKey, email)
}

func GetUserEmailFromCTX(ctx context.Context) (string, error) {
	userEmail, ok := ctx.Value(userEmailContextKey).(string)
	if !ok {
		return "", fmt.Errorf("failed to cast to string type")
	}
	return userEmail, nil
}

type HttpServer struct {
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (h HttpServer) GetDishesDishID(ctx context.Context, request GetDishesDishIDRequestObject) interface{} {
	//TODO implement me
	panic("implement me")
}

func (h HttpServer) PostDishesDishID(ctx context.Context, request PostDishesDishIDRequestObject) interface{} {
	//TODO implement me
	panic("implement me")
}

func (h HttpServer) GetGetAllDishes(ctx context.Context, request GetGetAllDishesRequestObject) interface{} {
	//TODO implement me
	panic("implement me")
}

func (h HttpServer) PostSearchDish(ctx context.Context, request PostSearchDishRequestObject) interface{} {
	//TODO implement me
	panic("implement me")
}
