package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"order-service/internal/usecase"
)

type HTTPPaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPPaymentClient(baseURL string, httpClient *http.Client) *HTTPPaymentClient {
	return &HTTPPaymentClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type paymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func (c *HTTPPaymentClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*usecase.PaymentResponse, error) {
	reqBody := paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal payment request: %w", err)
	}

	url := fmt.Sprintf("%s/payments", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var payResp paymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return nil, fmt.Errorf("decode payment response: %w", err)
	}

	return &usecase.PaymentResponse{
		Status:        payResp.Status,
		TransactionID: payResp.TransactionID,
	}, nil
}
