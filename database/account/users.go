package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"prodata/database"
	"prodata/logs"
	"time"

	"github.com/google/uuid"
)

type DataUserRegistry struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CPF       string `json:"cpf"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	Country   string `json:"country"`
	Birthdate string `json:"birthdate"`
	Company   string `json:"company,omitempty"`
}

type DataUser struct {
	UUID      string `json:"UUID"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	CPF       string `json:"cpf"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	Address2  string `json:"address2,omitempty"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
	Country   string `json:"country"`
	Birthdate string `json:"birthdate"`
	Company   string `json:"company,omitempty"`
}

func GetUser(email string) *DataUser {
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

	var user DataUser
	query := "SELECT uuid, first_name, last_name, email, password, cpf, phone, address, address2, city, state, zipcode, country, birth_date, company FROM userdata WHERE email =?"
	err = db.QueryRow(query, email).Scan(
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CPF,
		&user.Phone,
		&user.Address,
		&user.Address2,
		&user.City,
		&user.State,
		&user.ZipCode,
		&user.Country,
		&user.Birthdate,
		&user.Company)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.LogAndSendSystemMessage("User not found")
			return nil
		}
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	user.FirstName = Decrypt(user.FirstName)
	user.LastName = Decrypt(user.LastName)
	user.Password = Decrypt(user.Password)
	user.CPF = Decrypt(user.CPF)
	user.Phone = Decrypt(user.Phone)
	user.Address = Decrypt(user.Address)
	user.Address2 = Decrypt(user.Address2)
	user.City = Decrypt(user.City)
	user.State = Decrypt(user.State)
	user.ZipCode = Decrypt(user.ZipCode)
	user.Country = Decrypt(user.Country)
	user.Birthdate = Decrypt(user.Birthdate)
	user.Company = Decrypt(user.Company)

	return &user
}

func CreateUser(user *DataUserRegistry, data *Data) {
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

	exist := UserExistFromEmail(user.Email)
	if exist {
		logger.LogAndSendSystemMessage("User already exists")
		return
	}

	user.FirstName = Encrypt(user.FirstName)
	user.LastName = Encrypt(user.LastName)
	user.Password = Encrypt(HashPassword(user.Password))
	user.CPF = Encrypt(user.CPF)
	user.Phone = Encrypt(user.Phone)
	user.Address = Encrypt(user.Address)
	user.Address2 = Encrypt(user.Address2)
	user.City = Encrypt(user.City)
	user.State = Encrypt(user.State)
	user.ZipCode = Encrypt(user.ZipCode)
	user.Country = Encrypt(user.Country)
	user.Birthdate = Encrypt(user.Birthdate)
	user.Company = Encrypt(user.Company)

	query := "INSERT INTO userdata (uuid, first_name, last_name, email, password, cpf, phone, address, address2, city, state, zipcode, country, birth_date, company) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_, err = db.Exec(query,
		uuid.New(),
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.CPF,
		user.Phone,
		user.Address,
		user.Address2,
		user.City,
		user.State,
		user.ZipCode,
		user.Country,
		user.Birthdate,
		user.Company)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	userUUID := GetUserUUID(user.Email)
	var dataArray []Data

	dataArray = append(dataArray, *data)

	dataJson, err := json.Marshal(&dataArray)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	query = "INSERT INTO userinfo (uuid, data) VALUES (?, ?)"
	_, err = db.Exec(query,
		userUUID,
		string(dataJson))

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func GetUserUUID(email string) string {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	var userUuid string
	query := "SELECT uuid FROM userdata WHERE email = ?"
	err = db.QueryRow(query, email).Scan(&userUuid)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return userUuid
}

func UserExist(userID string) bool {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	var user string
	query := "SELECT uuid FROM userdata WHERE uuid =?"
	err = db.QueryRow(query, userID).Scan(&user)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return true
}

func UserExistFromEmail(email string) bool {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}(db)

	var user string
	query := "SELECT email FROM userdata WHERE email =?"
	err = db.QueryRow(query, email).Scan(&user)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return true
}

type Data struct {
	Device              string
	IPAddress           string
	Auth                bool
	MagicLinkId         string
	MagicLinkVerified   bool
	MagicLinkExpiration time.Time
}

func GetDataInfoUser(userUuid string) *[]Data {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	var dataString string

	err = db.QueryRow("SELECT data FROM userinfo WHERE uuid = ?", userUuid).Scan(&dataString)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	var values []Data

	err = json.Unmarshal([]byte(dataString), &values)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	return &values
}

func MagicLinkMarker(email, device, magicId string) string {
	logger := logs.NewSistemLogger()

	if !HasData(email) {
		logger.LogAndSendSystemMessage(email + " not have data")
	}

	db, err := database.InitializeDB()

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	userUUID := GetUserUUID(email)

	datas := GetDataInfoUser(userUUID)

	for i, data := range *datas {
		if device == data.Device {
			(*datas)[i].MagicLinkId = magicId
			(*datas)[i].MagicLinkExpiration = time.Now().Add(time.Minute * 10)
		}
	}

	bytes, err := json.Marshal(datas)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
	}

	query := "UPDATE userinfo SET data =? WHERE uuid =?"
	_, err = db.Exec(query, string(bytes), userUUID)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return magicId
}

func HasData(email string) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	userUuid := GetUserUUID(email)

	var data string

	err = db.QueryRow("SELECT data FROM userinfo WHERE uuid = ?", userUuid).Scan(&data)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return data != ""
}

func GetEmailByUuid(userUuid string) string {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	var email string

	err = db.QueryRow("SELECT email FROM userdata WHERE uuid = ?", userUuid).Scan(&email)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return email
}

func IsValidMagicLink(userUuid, magicLink string) (bool, string) {
	datas := GetDataInfoUser(userUuid)

	for _, data := range *datas {
		if data.MagicLinkId != "" && data.MagicLinkId == magicLink && data.MagicLinkExpiration.After(time.Now()) {
			return true, ""
		}

		if data.MagicLinkId != "" && data.MagicLinkId == magicLink && data.MagicLinkExpiration.Before(time.Now()) {
			return true, "retry"
		}
	}

	return false, ""
}

func ConvertToJson(data *[]Data, value *string) {
	logger := logs.NewSistemLogger()
	bytes, err := json.Marshal(&data)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	*value = string(bytes)
}

func UpdateDataInfo(param, userUuid string, args any) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("UPDATE userinfo SET %s=? WHERE uuid=?", param), args, userUuid)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return true
}

func ValidMagicLink(device, userUuid string) {
	datas := GetDataInfoUser(userUuid)

	for i, data := range *datas {
		if data.Device == device {
			(*datas)[i].MagicLinkId = ""
			(*datas)[i].MagicLinkVerified = true
		}
	}
	var value string

	ConvertToJson(datas, &value)

	UpdateDataInfo("data", userUuid, value)
}