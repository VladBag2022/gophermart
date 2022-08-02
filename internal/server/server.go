package server

import (
	"VladBag2022/gophermart/internal/storage"
)

type Server struct {
	repository storage.Repository
	config     *Config
}

func NewServer(repository storage.Repository, config *Config) Server {
	return Server{
		repository: repository,
		config:     config,
	}
}

//func (s Server) ListenAndServer() {
//	if err := http.ListenAndServe(s.config.Address, router(s)); err != nil {
//		fmt.Println(err)
//		return
//	}
//}
