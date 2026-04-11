package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderpb "github.com/Garryflop/microservices-gen-go/order"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type Server struct {
	orderpb.UnimplementedOrderTrackingServiceServer
	uc *usecase.OrderUseCase
}

func NewServer(uc *usecase.OrderUseCase) *Server {
	return &Server{uc: uc}
}

func RegisterServer(s *grpc.Server, srv *Server) {
	orderpb.RegisterOrderTrackingServiceServer(s, srv)
}

func (s *Server) SubscribeToOrderUpdates(req *orderpb.OrderRequest, stream orderpb.OrderTrackingService_SubscribeToOrderUpdatesServer) error {
	orderID := req.GetOrderId()
	if orderID == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	order, err := s.uc.GetOrder(stream.Context(), orderID)
	if err != nil {
		if err == domain.ErrOrderNotFound {
			return status.Error(codes.NotFound, "order not found")
		}
		return status.Error(codes.Internal, "failed to get order")
	}

	lastStatus := order.Status
	if err := stream.Send(&orderpb.OrderStatusUpdate{
		OrderId:   orderID,
		Status:    lastStatus,
		UpdatedAt: timestamppb.Now(),
	}); err != nil {
		return status.Error(codes.Internal, "failed to send initial status")
	}
	log.Printf("stream: sent initial status %s for order %s", lastStatus, orderID)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("stream: client disconnected for order %s", orderID)
			return nil
		case <-ticker.C:
			current, err := s.uc.GetOrder(context.Background(), orderID)
			if err != nil {
				log.Printf("stream: error polling order %s: %v", orderID, err)
				continue
			}

			if current.Status != lastStatus {
				lastStatus = current.Status
				if err := stream.Send(&orderpb.OrderStatusUpdate{
					OrderId:   orderID,
					Status:    lastStatus,
					UpdatedAt: timestamppb.Now(),
				}); err != nil {
					return status.Error(codes.Internal, "failed to send status update")
				}
				log.Printf("stream: status changed to %s for order %s", lastStatus, orderID)
			}

			if lastStatus == domain.StatusPaid || lastStatus == domain.StatusFailed || lastStatus == domain.StatusCancelled {
				log.Printf("stream: order %s reached terminal state %s", orderID, lastStatus)
				return nil
			}
		}
	}
}
