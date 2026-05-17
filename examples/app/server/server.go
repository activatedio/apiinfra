// Package server contains the example gRPC server implementation stubs.
package server

import (
	"context"

	"github.com/activatedio/apiinfra/examples/app/pb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type server struct{}

func (s *server) GetProduct(_ context.Context, _ *pb.GetProductRequest) (*pb.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) ListProduct(_ context.Context, _ *pb.ListProductRequest) (*pb.ListProductResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) CreateProduct(_ context.Context, _ *pb.CreateProductRequest) (*pb.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) UpdateProduct(_ context.Context, _ *pb.UpdateProductRequest) (*pb.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) PatchProduct(_ context.Context, _ *pb.PatchProductRequest) (*pb.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) DeleteProduct(_ context.Context, _ *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) GetProductReview(_ context.Context, _ *pb.GetProductReviewRequest) (*pb.ProductReview, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) ListProductReview(_ context.Context, _ *pb.ListProductReviewRequest) (*pb.ListProductReviewResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) CreateProductReview(_ context.Context, _ *pb.CreateProductReviewRequest) (*pb.ProductReview, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) UpdateProductReview(_ context.Context, _ *pb.UpdateProductReviewRequest) (*pb.ProductReview, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) PatchProductReview(_ context.Context, _ *pb.PatchProductReviewRequest) (*pb.ProductReview, error) {
	// TODO implement me
	panic("implement me")
}

func (s *server) DeleteProductReview(_ context.Context, _ *pb.DeleteProductReviewRequest) (*emptypb.Empty, error) {
	// TODO implement me
	panic("implement me")
}

// NewAPIServer returns a new instance of the example gRPC server.
func NewAPIServer() pb.ExampleApiServiceServer {
	return &server{}
}
