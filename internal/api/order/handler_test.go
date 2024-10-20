package order

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/romanp1989/gophermart/internal/cookies"
	"github.com/romanp1989/gophermart/internal/domain"
	"github.com/romanp1989/gophermart/internal/logger"
	"github.com/romanp1989/gophermart/internal/order"
	"github.com/romanp1989/gophermart/internal/order/mocks"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandler_CreateOrderHandler(t *testing.T) {
	type fields struct {
		service OrderServ
		logger  *zap.Logger
	}
	type want struct {
		err        error
		statusCode int
	}
	type args struct {
		numberOrder string
		want        want
	}

	userID := domain.UserID(1)
	orderNumber := `4285199826646588`

	zapLogger, _ := logger.Initialize("info")
	mockCtrl := gomock.NewController(t)
	mockOrderRep := mocks.NewMockOrderStorage(mockCtrl)
	defer mockCtrl.Finish()
	orderValidator := order.NewValidator(mockOrderRep)
	orderService := order.NewService(mockOrderRep, orderValidator, zapLogger)

	orderSuccess := domain.Order{
		ID:        0,
		CreatedAt: time.Now(),
		Number:    orderNumber,
		Status:    domain.OrderStatusNew,
		UserID:    userID,
	}

	orderSuccessCreated := orderSuccess
	orderSuccessCreated.ID = 1

	mockOrderRep.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).Return(orderSuccess.ID, nil)
	callFirst := mockOrderRep.EXPECT().LoadOrder(gomock.Any(), orderNumber).Return(nil, order.ErrNotFoundOrder).Times(1)
	mockOrderRep.EXPECT().LoadOrder(gomock.Any(), orderNumber).After(callFirst).Return(&orderSuccessCreated, nil).Times(1)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Order_create_success",
			fields: fields{
				service: orderService,
				logger:  zapLogger,
			},
			args: args{
				numberOrder: orderNumber,
				want: want{
					err:        nil,
					statusCode: http.StatusAccepted,
				},
			},
		},
		{
			name: "Order_create_duplicate_unsuccess",
			fields: fields{
				service: orderService,
				logger:  zapLogger,
			},
			args: args{
				numberOrder: orderNumber,
				want: want{
					err:        nil,
					statusCode: http.StatusOK,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				service: tt.fields.service,
				logger:  tt.fields.logger,
			}

			//reqBody, _ := json.Marshal(tt.args.numberOrder)
			body := strings.NewReader(tt.args.numberOrder)
			r := httptest.NewRequest(http.MethodPost, "/", body)

			rctx := context.WithValue(r.Context(), cookies.AuthKey, userID)
			r = r.WithContext(rctx)
			r.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			h.CreateOrderHandler(w, r)

			result := w.Result()

			_, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			if err != nil {
				message := fmt.Sprintf("Expected error = nil, actual error = %s\n", err.Error())
				t.Error(message)
			}

			if result.StatusCode != tt.args.want.statusCode {
				message := fmt.Sprintf("Expected error = %s, actual error = %s\n", tt.args.want.err, err.Error())
				t.Error(message)
			}
		})
	}
}
