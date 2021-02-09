package service

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	AddToIgnoredStocks endpoint.Endpoint
	GetArticlesByStock endpoint.Endpoint
}

type StockSymbolRequest struct {
	Stock string
}

func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		AddToIgnoredStocks: MakeAddToIgnoredStockEndpoint(s),
		GetArticlesByStock: MakeGetArticlesByStock(s),
	}
}

func MakeAddToIgnoredStockEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(StockSymbolRequest)

		return s.AddToIgnoredStocks(ctx, req.Stock), nil
	}
}

func MakeGetArticlesByStock(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(StockSymbolRequest)

		return s.GetArticles(ctx, req.Stock), nil
	}
}
