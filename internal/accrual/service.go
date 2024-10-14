package accrual

import (
	"encoding/json"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/domain"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type accrualStorage interface {
	GetNewOrdersToSend() ([]domain.Order, error)
	UpdateOrder(order *domain.AccrualResponse, userID domain.UserID) error
	AddBalance(o *domain.Balance) error
}

type Service struct {
	storage accrualStorage
	log     *zap.Logger
}

func NewService(orderStore accrualStorage, log *zap.Logger) *Service {
	return &Service{
		storage: orderStore,
		log:     log,
	}
}

func (s *Service) OrderStatusChecker() {

	for {
		newOrders, err := s.storage.GetNewOrdersToSend()
		if err != nil || len(newOrders) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, order := range newOrders {
			accrualResp, err := s.UploadWithdrawalFromAccrual(order.Number)
			if len(accrualResp.Order) != 0 || err == nil {
				if err = s.storage.UpdateOrder(accrualResp, order.UserID); err == nil {
					balance := &domain.Balance{
						OrderNumber: order.Number,
						UserID:      order.UserID,
						Sum:         accrualResp.Accrual,
						Type:        domain.BalanceTypeAdded,
						CreatedAt:   time.Now(),
					}
					if err = s.storage.AddBalance(balance); err != nil {
						continue
					}
				}
			}
		}
	}
}

func (s *Service) UploadWithdrawalFromAccrual(orderNumber string) (*domain.AccrualResponse, error) {
	var (
		accrualResp *domain.AccrualResponse
		response    *http.Response
	)

	for {
		response, err := http.Get(config.Options.FlagAccrualAddress + "/api/orders/" + orderNumber)
		if err != nil || response.StatusCode == http.StatusTooManyRequests ||
			response.StatusCode == http.StatusNoContent {

			if response != nil && response.Body != nil {
				response.Body.Close()
			}

			time.Sleep(5 * time.Second)
			continue
		}

		break
	}

	defer func() {
		if response.Body != nil {
			response.Body.Close()
		}
	}()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &accrualResp); err != nil {
		return nil, err
	}
	return accrualResp, nil
}
