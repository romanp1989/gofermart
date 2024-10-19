package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/romanp1989/gophermart/internal/cookies"
	"github.com/romanp1989/gophermart/internal/domain"
	"github.com/romanp1989/gophermart/internal/logger"
	"github.com/romanp1989/gophermart/internal/user"
	"github.com/romanp1989/gophermart/internal/user/mocks"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_RegisterHandler(t *testing.T) {
	type fields struct {
		service UserServ
		logger  *zap.Logger
	}
	type want struct {
		err        error
		statusCode int
	}
	type args struct {
		Login    string
		Password string
		want     want
	}

	zapLogger, _ := logger.Initialize("info")
	mockCtrl := gomock.NewController(t)
	mockUserRep := mocks.NewMockUserStorage(mockCtrl)
	defer mockCtrl.Finish()
	userService := user.NewService(mockUserRep, zapLogger)

	successUser := domain.User{
		Login:    "userNameTest",
		Password: "TestPass1",
	}
	successUserCreated := domain.User{
		ID:       1,
		Login:    successUser.Login,
		Password: "$2a$10$RzHySEpDUxl4frKhi5lSS.iQQ7.4UaRLo7dul8tbMUy.uLqRh4V2a",
	}

	unsuccessUserLogin := domain.User{
		Login:    "user",
		Password: "TestPass1",
	}

	unsuccessUserPassword := domain.User{
		Login:    successUser.Login,
		Password: "Test",
	}

	mockUserRep.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(&successUserCreated, nil)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "User_Create_Success",
			fields: fields{
				service: userService,
				logger:  zapLogger,
			},
			args: args{
				Login:    successUser.Login,
				Password: successUser.Password,
				want: want{
					err:        nil,
					statusCode: http.StatusOK,
				},
			},
		},
		{
			name: "User_Create_Unsuccess_Login",
			fields: fields{
				service: userService,
				logger:  zapLogger,
			},
			args: args{
				Login:    unsuccessUserLogin.Login,
				Password: unsuccessUserLogin.Password,
				want: want{
					err:        errors.New("неверный формат логина. логин может содержать только буквы латинского алфавита и цифры. длина логина от  до 16 символов"),
					statusCode: http.StatusBadRequest,
				},
			},
		},
		{
			name: "User_Create_Unsuccess_Password",
			fields: fields{
				service: userService,
				logger:  zapLogger,
			},
			args: args{
				Login:    unsuccessUserPassword.Login,
				Password: unsuccessUserPassword.Password,
				want: want{
					err: errors.New("неверный формат пароля. " +
						"пароль должен содержать хотя бы 1 букву латинского алфавита в верхнем регистре, 1 букву в нижнем регистре, " +
						"1 цифру, 1 специальный символ. длина пароля от 8 до 32 символов"),
					statusCode: http.StatusBadRequest,
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

			reqBody, _ := json.Marshal(tt.args)
			body := strings.NewReader(string(reqBody))
			r := httptest.NewRequest(http.MethodPost, "/", body)

			userID := 1
			rctx := context.WithValue(r.Context(), cookies.AuthKey, userID)
			r = r.WithContext(rctx)
			r.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			h.RegisterHandler(w, r)

			result := w.Result()

			_, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			if err != nil {
				message := fmt.Sprintf("expected error = nil, actual error = %s\n", err.Error())
				t.Error(message)
			}

			if result.StatusCode != tt.args.want.statusCode {
				message := fmt.Sprintf("expected error = %s, actual error = %s\n", tt.args.want.err, err.Error())
				t.Error(message)
			}

		})
	}
}
