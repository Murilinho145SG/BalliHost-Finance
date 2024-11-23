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

func CreateUser(user *DataUserRegistry) {
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

	query = "INSERT INTO userinfo (uuid, auth, admin, data) VALUES (?, ?, ?, ?)"
	_, err = db.Exec(query,
		userUUID, 0, 0, "[]")

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
	defer db.Close()

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

type UserData struct {
	Auth                bool
	MagicLinkId         string
	MagicLinkVerified   bool
	MagicLinkExpiration time.Time
	Data                []Data
}

type Data struct {
	Device    string
	DeviceId  string
	IPAddress string
}

func GetDataInfoUser(userUuid string) *UserData {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}
	defer db.Close()

	var userData UserData

	var auth bool
	var magicAuthId string
	var magicAuthVerified bool
	var magicAuthExpirationStr string
	var dataString string

	err = db.QueryRow("SELECT auth, magic_auth_id, magic_auth_verified, magic_auth_expiration, data FROM userinfo WHERE uuid = ?", userUuid).Scan(
		&auth,
		&magicAuthId,
		&magicAuthVerified,
		&magicAuthExpirationStr,
		&dataString)

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

	magicAuthExpiration, err := time.ParseInLocation(time.DateTime, magicAuthExpirationStr, time.Local)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	userData.Data = values
	userData.Auth = auth
	userData.MagicLinkId = magicAuthId
	userData.MagicLinkVerified = magicAuthVerified
	userData.MagicLinkExpiration = magicAuthExpiration

	return &userData
}

func CreateNewData(device, ip string) Data {
	DeviceID, err := uuid.NewUUID()
	if err != nil {
		return Data{}
	}

	return Data{
		IPAddress: ip,
		Device:    device,
		DeviceId:  DeviceID.String(),
	}
}

func MagicLinkMarker(email, magicId string) string {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}
	defer db.Close()

	userUUID := GetUserUUID(email)
	expiration := time.Now().Add(time.Minute * 10)

	query := "UPDATE userinfo SET magic_auth_id = ?, magic_auth_verified = ?, magic_auth_expiration = ? WHERE uuid = ?"
	_, err = db.Exec(query, magicId, 0, expiration.Format(time.DateTime), userUUID)

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
	defer db.Close()

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
	defer db.Close()

	var email string

	err = db.QueryRow("SELECT email FROM userdata WHERE uuid = ?", userUuid).Scan(&email)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return email
}

func IsValidMagicLink(userUuid, magicLink string) (bool, string) {
	data := GetDataInfoUser(userUuid)

	if data.MagicLinkId != "" && data.MagicLinkId == magicLink && data.MagicLinkExpiration.After(time.Now()) {
		return true, ""
	}

	if data.MagicLinkId != "" && data.MagicLinkId == magicLink && data.MagicLinkExpiration.Before(time.Now()) {
		return true, "retry"
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

func ValidMagicLink(userUuid, device, ip string) string {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}
	defer db.Close()

	data := CreateNewData(device, ip)
	byte, err := json.Marshal(data)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}
	_, err = db.Exec("UPDATE userinfo SET magic_auth_id =?, magic_auth_verified =?, data =? WHERE uuid =?", "", 1, string(byte), userUuid)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}
	
	return data.DeviceId
}

func IsAdmin(userId string) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}
	defer db.Close()

	var value bool
	err = db.QueryRow("SELECT admin FROM userinfo WHERE uuid = ?", userId).Scan(&value)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return value
}

type Attempts struct {
	Email       string
	IpAddress   string
	Attempts    int
	Date        time.Time
	CanBackDate time.Time
}

func RegisterAttempts(email, ipAddress string) {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
	defer db.Close()

	var attempts int
	var canBackStr string
	var ip string

	err = db.QueryRow("SELECT attempts, ip, can_back FROM registration_attempts WHERE email = ?", email).Scan(&attempts, &ip, &canBackStr)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = db.Exec("INSERT INTO registration_attempts (email, ip, attempts, date, can_back) VALUES (?, ?, ?, ?, ?)",
				email,
				ipAddress,
				1,
				time.Now().Format(time.DateTime),
				time.Now().Format(time.DateTime))

			if err != nil {
				logger.LogAndSendSystemMessage(err.Error())
			}

			return
		} else {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}

	canBack, err := time.ParseInLocation(time.DateTime, canBackStr, time.Local)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	if ipAddress != ip {
		_, err := db.Exec("UPDATE registration_attempts SET ip = ? WHERE email = ?", ipAddress, email)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}

	if canBack.After(time.Now()) {
		return
	}

	if attempts+1 == 5 {
		_, err := db.Exec("UPDATE registration_attempts SET attempts = ?, can_back = ? WHERE email = ?",
			0,
			time.Now().Add(10*time.Minute).Format(time.DateTime),
			email)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
		}

		return
	} else {
		fmt.Println("Colocando Attempts")
		_, err := db.Exec("UPDATE registration_attempts SET attempts = ? WHERE email = ?", attempts+1, email)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
		}

		return
	}
}

func GetAttempts(email string) *Attempts {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	defer db.Close()

	var attempts Attempts

	var date string
	var canBack string

	err = db.QueryRow("SELECT * FROM registration_attempts WHERE email = ?", email).Scan(
		&attempts.Email,
		&attempts.IpAddress,
		&attempts.Attempts,
		&date,
		&canBack)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	attempts.Date, err = time.ParseInLocation(time.DateTime, date, time.Local)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	attempts.CanBackDate, err = time.ParseInLocation(time.DateTime, canBack, time.Local)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	return &attempts
}

func CanLogin(email string) bool {
	attempts := GetAttempts(email)

	if attempts == nil {
		return true
	}

	if attempts.CanBackDate.After(time.Now()) {
		return false
	}

	return true
}
