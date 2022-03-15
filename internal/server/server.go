package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/avitoTask/internal/conversion"
	"github.com/avitoTask/internal/service"
	"github.com/avitoTask/models"

	"github.com/gorilla/mux"
)

type Server struct {
	config  *Config
	Log     *log.Logger
	router  *mux.Router
	service *service.Service
	convS   *conversion.ConversionService
}

func NerServer(conf *Config, s *service.Service, c *conversion.ConversionService) *Server {
	return &Server{config: conf,
		Log:     log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		router:  mux.NewRouter(),
		service: s,
		convS:   c,
	}
}

func (s *Server) StartServer() error {
	s.configRouter()

	return http.ListenAndServe(s.config.Server.Host+":"+s.config.Server.Port, s.router)
}

func (s *Server) configRouter() {
	s.router.HandleFunc("/getBalance", s.getBalance).Methods("GET")
	s.router.HandleFunc("/changeBalance", s.changeBalance).Methods("POST")
	s.router.HandleFunc("/transfer", s.transfer).Methods("POST")
	s.router.HandleFunc("/logs", s.getLogs).Methods("GET")
}

func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
	}

	data := &models.RequestBalance{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	res, err := s.service.GetBalance(data.Id)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

//  Изменить баланс(списание, начисление)
func (s *Server) changeBalance(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
	}

	data := &models.RequestChangeBalance{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	if data.Sum < 0 {
		http.Error(w, errors.New("отрицательная сумма").Error(),
			http.StatusInternalServerError)
		return
	}

	err = s.service.ChangeBalance(data.Operation, data.UserId, data.Sum)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Перевод средств
func (s *Server) transfer(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
		return
	}

	data := &models.RequestTransfer{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	if data.Sum < 0 {
		http.Error(w, errors.New("отрицательная сумма").Error(),
			http.StatusInternalServerError)
		return
	}

	err = s.service.Transfer(data.SenderId, data.ReceiverId, data.Sum)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusBadRequest)
	}

	data := &models.RequestLogs{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	res, err := s.service.GetLog(data.Id, data.CountOperation)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
