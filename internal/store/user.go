package store

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Shimanta12/go-rest-api-postgres/internal/model"
) 

var (
	ErrNotFound = errors.New("user not found")
	ErrConflict = errors.New("email already in use")
	ErrEmptyField = errors.New("json field cannot be empty")
)

type UserStore struct{
	db *sql.DB
}

func NewUserStore(db * sql.DB) *UserStore{
	return &UserStore{db : db}
}

func (s *UserStore) Create(userReq *model.UserRequest) (*model.User, error){
	user := &model.User{}

	err := s.db.QueryRow(
		`INSERT INTO users(name, email)
		VALUES($1, $2)
		RETURNING id, name, email, created_at`,
		userReq.Name, userReq.Email,
	).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil{
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique constraint"){
			return nil, ErrConflict
		}
		return nil, err
	}
	return user, nil
}

func (s *UserStore) GetById(id int) (*model.User, error){
	user := &model.User{}
	err := s.db.QueryRow(
		`SELECT id, name, email, created_at FROM users
		WHERE id = $1`,
		id,
	).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil{
		if err == sql.ErrNoRows{
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserStore) List() ([]*model.User, error){
	rows, err := s.db.Query(`SELECT id, name, email, created_at FROM users`)
	if err != nil{
		return nil, err
	}
	defer rows.Close()
	users := make([]*model.User, 0)
	for rows.Next(){
		user := &model.User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil{
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil{
		return nil, err
	}
	return users, nil
}

func (s *UserStore) Update(id int, updateUserReq *model.UpdateUserRequest) (*model.User, error){
	user, errNotFound := s.GetById(id)
	if errNotFound != nil{
		return nil, ErrNotFound
	}
	if updateUserReq.Name != nil && *updateUserReq.Name == ""{
		return nil, ErrEmptyField
	}
	if updateUserReq.Email != nil && *updateUserReq.Email == ""{
		return nil, ErrEmptyField
	}
	if updateUserReq.Name != nil{
		user.Name = *updateUserReq.Name
	}
	if updateUserReq.Email != nil{
		user.Email = *updateUserReq.Email
	}
	err := s.db.QueryRow(
		`UPDATE users
		SET name = $1,
		email = $2
		WHERE id = $3
		RETURNING id, name, email, created_at`,
		user.Name, user.Email, id,
	).Scan(&user.Id, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil{
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "unique constraint"){
			return nil, ErrConflict
		}
		return nil, err
	}
	return user, nil		
}

func (s *UserStore) Delete(id int) error{
	
	result, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil{
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil{
		return err
	}
	if rowsAffected == 0{
		return ErrNotFound
	}
	return nil
}