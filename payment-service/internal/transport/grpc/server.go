package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentpb "github.com/Garryflop/microservices-gen-go/payment"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

// Server implements the gRPC PaymentServiceServer interface.
// It delegates all business logic to the UseCase layer (Clean Architecture).
type Server struct {
	paymentpb.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewServer(uc *usecase.PaymentUseCase) *Server {
	return &Server{uc: uc}
}

// RegisterServer registers the PaymentService gRPC server.
func RegisterServer(s *grpc.Server, srv *Server) {
	paymentpb.RegisterPaymentServiceServer(s, srv)
}

// ProcessPayment handles gRPC payment requests by delegating to the use case layer.
func (s *Server) ProcessPayment(ctx context.Context, req *paymentpb.PaymentRequest) (*paymentpb.PaymentResponse, error) {
	payment, err := s.uc.ProcessPayment(ctx, req.GetOrderId(), req.GetAmount())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to process payment")
	}

	return &paymentpb.PaymentResponse{
		Id:            payment.ID,
		OrderId:       payment.OrderID,
		TransactionId: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
		CreatedAt:     timestamppb.Now(),
	}, nil
}
