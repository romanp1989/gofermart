package main

import (
	accrualSer "github.com/romanp1989/gophermart/internal/accrual"
	"github.com/romanp1989/gophermart/internal/api/balance"
	"github.com/romanp1989/gophermart/internal/api/order"
	"github.com/romanp1989/gophermart/internal/api/user"
	balanceSer "github.com/romanp1989/gophermart/internal/balance"
	"github.com/romanp1989/gophermart/internal/config"
	"github.com/romanp1989/gophermart/internal/database"
	"github.com/romanp1989/gophermart/internal/logger"
	orderSer "github.com/romanp1989/gophermart/internal/order"
	"github.com/romanp1989/gophermart/internal/server"
	userSer "github.com/romanp1989/gophermart/internal/user"
	"go.uber.org/zap"
	"log"
	"os"
	"time"
)

func main() {
	os.Exit(Init())
}

func Init() int {
	zapLogger, err := logger.Initialize("info")
	if err != nil {
		log.Printf("can't initalize logger %s", err)
		return 1
	}

	defer func(zLog *zap.Logger) {
		_ = zLog.Sync()
	}(zapLogger)

	db, err := database.NewDB(&database.Config{
		Dsn:             config.Options.FlagDBDsn,
		MaxIdleConn:     1,
		MaxOpenConn:     10,
		MaxLifetimeConn: time.Minute * 1,
	})
	if err != nil {
		zapLogger.Fatal("Database init error: ", zap.String("error", err.Error()))
	}

	userRepository := userSer.NewDBStorage(db)
	userService := userSer.NewService(userRepository, zapLogger)
	userHandler := user.NewUserHandler(userService, zapLogger)

	orderRepository := orderSer.NewDBStorage(db)
	orderValidator := orderSer.NewValidator(orderRepository)
	orderService := orderSer.NewService(orderRepository, orderValidator, zapLogger)
	orderHandler := order.NewOrderHandler(orderService, zapLogger)

	balanceRepository := balanceSer.NewDBStorage(db)
	balanceService := balanceSer.NewService(balanceRepository, zapLogger)
	balanceHandler := balance.NewBalanceHandler(balanceService, zapLogger)

	route := server.NewRoutes(userHandler, orderHandler, balanceHandler)
	httpServer := server.NewApp(zapLogger, route)

	errChannel := make(chan error, 1)
	oss, stop := make(chan os.Signal, 1), make(chan struct{}, 1)

	accrualRepository := accrualSer.NewDBStorage(db)
	accrualService := accrualSer.NewService(accrualRepository, zapLogger)

	//Запускаем горутину для прослушивания Accrual сервиса
	go accrualService.OrderStatusChecker()

	go func() {
		<-oss

		stop <- struct{}{}
	}()

	go func() {
		if err := httpServer.RunServer(); err != nil {
			errChannel <- err
		}
	}()

	for {
		select {
		case err := <-errChannel:
			zapLogger.Warn("Can't run application", zap.Error(err))
			return 0
		case <-stop:
			httpServer.Stop()
			return 0
		}
	}

}
