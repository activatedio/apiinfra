package server

import (
	"context"

	"github.com/activatedio/apiinfra/examples/app/pb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type server struct{}

func (s *server) GetProduct(ctx context.Context, request *pb.GetProductRequest) (*pb.Product, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) ListProduct(ctx context.Context, request *pb.ListProductRequest) (*pb.ListProductResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) CreateProduct(ctx context.Context, request *pb.CreateProductRequest) (*pb.Product, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) UpdateProduct(ctx context.Context, request *pb.UpdateProductRequest) (*pb.Product, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) PatchProduct(ctx context.Context, request *pb.PatchProductRequest) (*pb.Product, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) DeleteProduct(ctx context.Context, request *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) GetProductReview(ctx context.Context, request *pb.GetProductReviewRequest) (*pb.ProductReview, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) ListProductReview(ctx context.Context, request *pb.ListProductReviewRequest) (*pb.ListProductReviewResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) CreateProductReview(ctx context.Context, request *pb.CreateProductReviewRequest) (*pb.ProductReview, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) UpdateProductReview(ctx context.Context, request *pb.UpdateProductReviewRequest) (*pb.ProductReview, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) PatchProductReview(ctx context.Context, request *pb.PatchProductReviewRequest) (*pb.ProductReview, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) DeleteProductReview(ctx context.Context, request *pb.DeleteProductReviewRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func NewApiServer() pb.ExampleApiServiceServer {
	return &server{}
}
