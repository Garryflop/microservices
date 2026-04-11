package infrastructure

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	paymentpb "github.com/Garryflop/microservices-gen-go/payment"

	"order-service/internal/usecase"
)

type GRPCPaymentClient struct {
	client paymentpb.PaymentServiceClient
	conn   *grpc.ClientConn
}

func NewGRPCPaymentClient(addr string) (*GRPCPaymentClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	return &GRPCPaymentClient{
		client: paymentpb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *GRPCPaymentClient) Close() error {
	return c.conn.Close()
}

// AuthorizePayment implements usecase.PaymentClient interface via gRPC.
func (c *GRPCPaymentClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResponse, error) {
	resp, err := c.client.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}

	return &usecase.PaymentResponse{
		Status:        resp.GetStatus(),
		TransactionID: resp.GetTransactionId(),
	}, nil
}
