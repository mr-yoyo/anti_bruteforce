package http

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/infrastructure/db"
)

var (
	limiter       domain.Limiter
	whiteListRepo domain.IPListRepository
	blackListRepo domain.IPListRepository
	json          = jsoniter.ConfigCompatibleWithStandardLibrary
)

type AuthDataRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
	IP       string `json:"ip" validate:"required,ip"`
}

type AuthDataResponse struct {
	Ok bool `json:"ok"`
}

func auth(w http.ResponseWriter, r *http.Request) {
	var data AuthDataRequest
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validate.Struct(data)
	if err != nil {
		validationFailed(w, err)
		return
	}

	authData := domain.AuthData{
		Login:    data.Login,
		Password: data.Password,
		IP:       net.ParseIP(data.IP),
	}

	ok, err := limiter.IsAllowed(authData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_ = json.NewEncoder(w).Encode(AuthDataResponse{Ok: ok})
}

type AddIPRequest struct {
	Subnet string `json:"subnet" validate:"required,ip4_net"`
}

type AddIPResponse struct {
	ID      int       `json:"id"`
	Address string    `json:"address"`
	AddedAt time.Time `json:"addedAt"`
}

func addIPNet(w http.ResponseWriter, r *http.Request) {
	var data AddIPRequest
	w.Header().Set("Content-Type", "application/json")
	kind := mux.Vars(r)["kind"]

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validate.Struct(data)
	if err != nil {
		validationFailed(w, err)
		return
	}

	_, ip4net, err := net.ParseCIDR(data.Subnet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var item *domain.AddressItem
	switch kind {
	case "blacklist":
		item, err = blackListRepo.Add(ip4net)
	case "whitelist":
		item, err = whiteListRepo.Add(ip4net)
	default:
		item, err = &domain.AddressItem{}, errors.New("unknown kind of list")
	}

	if err != nil {
		if errors.Is(err, &domain.IPDuplicateError{}) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if item.ID == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = json.NewEncoder(w).Encode(AddIPResponse{
		ID:      item.ID,
		Address: item.Address,
		AddedAt: item.AddedAt,
	})
}

type DeleteIPRequest struct {
	AddIPRequest
}

func deleteIPNet(w http.ResponseWriter, r *http.Request) {
	var data DeleteIPRequest

	w.Header().Set("Content-Type", "application/json")
	kind := mux.Vars(r)["kind"]

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validate.Struct(data)
	if err != nil {
		validationFailed(w, err)
		return
	}

	_, ip4net, err := net.ParseCIDR(data.Subnet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch kind {
	case "blacklist":
		_, err = blackListRepo.Delete(ip4net)
	case "whitelist":
		_, err = whiteListRepo.Delete(ip4net)
	default:
		err = errors.New("unknown kind of list")
	}

	if err != nil {
		if errors.Is(err, &domain.IPNotExistsError{}) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

type BucketDataRequest struct {
	Login string `json:"login" validate:"required_without=IP"`
	IP    string `json:"ip" validate:"required_without=Login,omitempty,ip"`
}

func resetBucket(w http.ResponseWriter, r *http.Request) {
	var data BucketDataRequest
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = validate.Struct(data)
	if err != nil {
		validationFailed(w, err)
		return
	}

	bucketData := domain.BucketData{
		Login: data.Login,
		IP:    net.ParseIP(data.IP),
	}

	isDeleted, err := limiter.DeleteBucket(bucketData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isDeleted {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) bindHandlers() {
	registerValidations()
	limiter = s.Limiter
	whiteListRepo = db.NewWhitelist()
	blackListRepo = db.NewBlacklist()
	s.router.HandleFunc("/auth", auth).Methods("POST")
	s.router.HandleFunc("/{kind:(?:blacklist|whitelist)}", addIPNet).Methods("POST")
	s.router.HandleFunc("/{kind:(?:blacklist|whitelist)}", deleteIPNet).Methods("DELETE")
	s.router.HandleFunc("/bucket", resetBucket).Methods("DELETE")
}
