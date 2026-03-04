package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strconv"

	"github.com/Shimanta12/go-rest-api-postgres/internal/model"
	"github.com/Shimanta12/go-rest-api-postgres/internal/store"
)


type UserHandler struct{
	store *store.UserStore
}

func NewUserHandler(store *store.UserStore) *UserHandler{
	return &UserHandler{store: store}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux){
	mux.HandleFunc("POST /users", h.CreateUserHandler)
	mux.HandleFunc("GET /users", h.ListUsers)
	mux.HandleFunc("GET /users/{id}", h.GetUserById)
	mux.HandleFunc("PATCH /users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", h.DeleteUser)
}

func writeJson(w http.ResponseWriter, status int, data any){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil{
		log.Printf("failed to encode response: %v\n", err)
	}
}

func parseId(r *http.Request) (int, error){
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil{
		return -1, err
	}
	return id, nil
}

func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request){
	userReq := &model.UserRequest{}
	err := json.NewDecoder(r.Body).Decode(userReq)
	if err != nil{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "bad request: invalid json",
		})
		return
	}
	if userReq.Name == ""{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "name is required",
		})
		return
	}
	if userReq.Email == ""{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "email is required",
		})
		return
	}

	if _, err := mail.ParseAddress(userReq.Email); err != nil{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "invalid email",
		})
		return
	}

	user, err := h.store.Create(userReq)
	if err != nil{
		if errors.Is(err, store.ErrConflict){
			writeJson(w, http.StatusConflict, map[string]string{
				"error" : "email already in use",
			})
			return
		}
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error" : "something went wrong",
		})
		return
	}
	writeJson(w, http.StatusCreated, user)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request){
	users, err := h.store.List()
	if err != nil{
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error" : "something went wrong",
		})
		return
	}
	writeJson(w, http.StatusOK, users)
}

func (h *UserHandler) GetUserById(w http.ResponseWriter, r *http.Request){
	id, err := parseId(r)
	if err != nil{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "id must be a valid number",
		})
		return
	}
	user, err := h.store.GetById(id)
	if err != nil{
		if errors.Is(err, store.ErrNotFound){
			writeJson(w, http.StatusNotFound, map[string]string{
				"error" : "user not found",
			})
			return
		}
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error" : "something went wrong",
		})
		return
	}
	writeJson(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request){
	id, err := parseId(r)
	if err != nil{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "id must be a valid number",
		})
		return
	}
	userUpdateReq := &model.UpdateUserRequest{}
	if err := json.NewDecoder(r.Body).Decode(userUpdateReq); err != nil{
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error" : "bad request: invalid json",
		})
		return
	}
	if userUpdateReq.Email != nil{
		if _, err := mail.ParseAddress(*userUpdateReq.Email); err != nil{
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error" : "invalid email",
			})
			return
		}
	}

	user, err := h.store.Update(id, userUpdateReq)
	if err != nil{
		if errors.Is(err, store.ErrConflict){
			writeJson(w, http.StatusConflict, map[string]string{
				"error" : "email already in use",
			})
			return
		}
		if errors.Is(err, store.ErrEmptyField){
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error" : "json field cannot be empty",
			})
			return
		}
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error" : "something went wrong",
		})
		return
	}
	writeJson(w, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request){
	id, err := parseId(r)
	if err != nil{
		writeJson(w, http.StatusUnprocessableEntity, map[string]string{
			"error" : "id must be a valid number",
		})
		return
	}
	if err := h.store.Delete(id); err != nil{
		if errors.Is(err, store.ErrNotFound){
			writeJson(w, http.StatusNotFound, map[string]string{
				"error" : "user not found",
			})
			return
		}
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error" : "something went wrong",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}