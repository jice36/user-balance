package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/avitoTask/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Service struct {
	DB *sql.DB
}

func NewService(name, pass, user string) (*Service, error) {
	s := &Service{}

	connStr := "user=" + user + " password=" + pass + " dbname=" + name + " sslmode=disable"
	fmt.Println(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	s.DB = db
	return s, nil
}

func (s *Service) GetBalance(userId string) (*models.ResponseBalance, error) {
	q := `select balance from user_balance where id = $1`

	var balance sql.NullFloat64

	err := s.DB.QueryRow(q, userId).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователя не существует")
		}
		return nil, err
	}

	res := &models.ResponseBalance{UserId: userId, Balance: balance.Float64}
	return res, nil
}

func (s *Service) Transfer(sender, receiver string, value float64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	q := `UPDATE user_balance
			SET balance = CASE 
				 WHEN (balance - $1 ) > 0 THEN balance - $1 
	  			WHEN (balance - $1 ) = 0 THEN 0
	 			else balance
	 		END
	 	WHERE id = $2;`

	res, err := tx.Exec(q, value, sender)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = checkUser(tx, res); err != nil {
		return err
	}

	q = `UPDATE user_balance
			SET balance = balance + $1
		WHERE id = $2;`

	res, err = tx.Exec(q, value, receiver)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = checkUser(tx, res); err != nil {
		return err
	}

	err = s.addLog(tx, sender, receiver, "debit", value)
	if err != nil {
		return err
	}

	err = s.addLog(tx, receiver, sender, "accrue", value)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (s *Service) ChangeBalance(operation string, userId string, sum float64) error {
	var err error

	switch operation {
	case "accrue":
		err = s.accrueValue(userId, sum)
	case "debit":
		err = s.debitingValue(userId, sum)
	default:
		err = errors.New("неверная операция")
	}

	return err
}

func (s *Service) accrueValue(userId string, sum float64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	q := `UPDATE user_balance
			SET balance = balance + $1
			WHERE id = $2;`

	res, err := tx.Exec(q, sum, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = checkUser(tx, res); err != nil {
		return err
	}

	err = s.addLog(tx, userId, "", "accrue", sum)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (s *Service) debitingValue(userId string, sum float64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	q := `UPDATE user_balance
			SET balance = CASE 
				 WHEN (balance - $1 ) < 0 THEN balance - $1 
	  			WHEN (balance - $1 ) = 0 THEN 0
	 			else balance
	 		END
	 	WHERE id = $2`

	res, err := tx.Exec(q, sum, userId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err = checkUser(tx, res); err != nil {
		return err
	}

	err = s.addLog(tx, userId, "", "debit", sum)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func (s *Service) GetLog(userId string, count int) (*models.ResponsetLogs, error) {
	q := `select id_log, id_user, date, description from logs where id_user = $1 LIMIT $2`

	var logs []models.Log

	rows, err := s.DB.Query(q, userId, count)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("операции не найдены")
		}
		return nil, err
	}
	for rows.Next() {
		log := models.Log{}
		var logId, userId uuid.NullUUID

		err := rows.Scan(&logId, &userId, &log.Date, &log.Description)
		if err != nil {
			return nil, err
		}

		if !userId.Valid || !logId.Valid {
			return nil, errors.New("не удалось получить данный")
		}

		log.Id = logId.UUID.String()
		log.UserId = userId.UUID.String()

		logs = append(logs, log)
	}

	res := &models.ResponsetLogs{Logs: logs}
	return res, nil
}

func (s *Service) addLog(tx *sql.Tx, userId, anotherUser, operation string, sum float64) error {
	q := `INSERT INTO logs (id_log, id_user, date, description) 
				VALUES ($1, $2, $3, $4) returning id_log `

	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	var newId uuid.NullUUID
	var desc string

	switch operation {
	case "accrue":
		desc = "получено " + strconv.FormatFloat(sum, 'f', 2, 64)
		if anotherUser != "" {
			desc += " от " + anotherUser
		}
	case "debit":
		desc = "списано " + strconv.FormatFloat(sum, 'f', 2, 64)
		if anotherUser != "" {
			desc += " пользователю " + anotherUser
		}
	default:
		return errors.New("неверная операция")
	}

	err = tx.QueryRow(q, id, userId, time.Now(), desc).Scan(&newId)
	if err != nil && !newId.Valid {
		tx.Rollback()
		return err
	}

	return nil
}

func checkUser(tx *sql.Tx, res sql.Result) error {
	var err error
	var count int64

	count, err = res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if count == 0 {
		return errors.New("пользователь не найден")
	}
	return nil
}
