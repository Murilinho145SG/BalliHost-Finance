package transactions

import (
	"database/sql"
	"prodata/database"
	"prodata/logs"
	"time"
)

type Transaction struct {
	ID          int       `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	ItemName    string    `json:"item_name"`
	Price       float64   `json:"price"`
	PaymentType string    `json:"payment_type"`
	Status      string    `json:"status"`
	Date        time.Time `json:"date"`
}

func SetTransaction(tx Transaction) {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	query := "INSERT INTO transaction (id, first_name, last_name, email, item_name, price, payment_type, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	_, err = db.Exec(query,
		tx.ID,
		tx.FirstName,
		tx.LastName,
		tx.Email,
		tx.ItemName,
		tx.Price,
		tx.PaymentType,
		tx.Status,
	)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func GetTransaction(id int) *Transaction {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	var tx Transaction

	err = db.QueryRow("SELECT * FROM transaction WHERE id = ?", id).Scan(&tx.ID, &tx.FirstName, &tx.LastName, &tx.Email, &tx.ItemName, &tx.Price, &tx.PaymentType, &tx.Status, &tx.Date)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.LogAndSendSystemMessage("Transaction not found")
			return nil
		}
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	return &tx
}

func UpdateTransactionStatus(id int, status string) {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	query := "UPDATE transaction SET status = ? WHERE id =?"

	_, err = db.Exec(query, status, id)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func GetAllTransactions() *[]Transaction {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	var tx []Transaction

	rows, err := db.Query("SELECT * FROM transaction")
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	defer rows.Close()

	for rows.Next() {
		var t Transaction
		var date string

		err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.ItemName, &t.Price, &t.PaymentType, &t.Status, &date)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return nil
		}

		parse, err := time.Parse(time.DateTime, date)

		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return nil
		}

		t.Date = parse
		tx = append(tx, t)
	}
	return &tx
}
