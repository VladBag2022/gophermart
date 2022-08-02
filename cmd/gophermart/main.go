package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"VladBag2022/gophermart/internal/server"
	"VladBag2022/gophermart/internal/storage"
)

func main() {
	config, err := server.NewConfig()
	if err != nil {
		log.Error(fmt.Sprintf("Unable to read configuration from environment variables: %s", err))
		return
	}

	addressPtr := flag.StringP("address", "a", "", "server address - host:port")
	databasePtr := flag.StringP("database", "d", "", "database URI")
	accrualPtr := flag.StringP("accrual", "r", "", "accrual system address")
	flag.Parse()

	if len(*addressPtr) != 0 {
		config.Address = *addressPtr
	}
	if len(*accrualPtr) != 0 {
		config.Accrual = *accrualPtr
	}

	repository, err := storage.NewPostgresRepository(
		context.Background(),
		*databasePtr,
	)
	if err != nil {
		log.Error(err)
		return
	}
	defer repository.Close()

	app := server.NewServer(repository, config)

	go func() {
		app.ListenAndServer()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigChan
}
