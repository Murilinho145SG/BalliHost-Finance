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
	query := "SELECT uuid, first_name, last_name, email, password FROM userdata WHERE email =?"
	err = db.QueryRow(query, email).Scan(
		&user.UUID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password)

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

	query := "INSERT INTO userdata (uuid, first_name, last_name, email, password) VALUES (?,?,?,?,?)"
	_, err = db.Exec(query,
		uuid.New(),
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	userUUID := GetUserUUID(user.Email)

	query = "INSERT INTO userinfo (uuid, auth, admin, devices, data) VALUES (?, ?, ?, ?, ?)"
	_, err = db.Exec(query,
		userUUID, 0, 0, "[]", "[]")

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
	Devices             []Devices
}

type Devices struct {
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

	err = db.QueryRow("SELECT auth, magic_auth_id, magic_auth_verified, magic_auth_expiration, devices FROM userinfo WHERE uuid = ?", userUuid).Scan(
		&auth,
		&magicAuthId,
		&magicAuthVerified,
		&magicAuthExpirationStr,
		&dataString)

	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	var values []Devices

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

	userData.Devices = values
	userData.Auth = auth
	userData.MagicLinkId = magicAuthId
	userData.MagicLinkVerified = magicAuthVerified
	userData.MagicLinkExpiration = magicAuthExpiration

	return &userData
}

func CreateNewData(device, ip string) Devices {
	DeviceID, err := uuid.NewUUID()
	if err != nil {
		return Devices{}
	}

	return Devices{
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

	var devices string

	err = db.QueryRow("SELECT devices FROM userinfo WHERE uuid = ?", userUuid).Scan(&devices)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return devices != ""
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

func IsValidMagicLink(userUuid, magicLink string) bool {
	data := GetDataInfoUser(userUuid)

	if data.MagicLinkId != "" && data.MagicLinkId == magicLink && data.MagicLinkExpiration.After(time.Now()) {
		return true
	}

	return false
}

func ConvertToJson(devices *[]Devices, value *string) {
	logger := logs.NewSistemLogger()
	bytes, err := json.Marshal(&devices)
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

	devices := CreateNewData(device, ip)
	byte, err := json.Marshal([]Devices{devices})
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}
	_, err = db.Exec("UPDATE userinfo SET magic_auth_id =?, magic_auth_verified =?, devices =? WHERE uuid =?", "", 1, string(byte), userUuid)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return ""
	}

	return devices.DeviceId
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
	var email2 string

	err = db.QueryRow("SELECT attempts, email, can_back FROM registration_attempts WHERE ip = ?", ipAddress).Scan(&attempts, &email2, &canBackStr)
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

	if email != email2 {
		_, err := db.Exec("UPDATE registration_attempts SET email = ? WHERE ip = ?", email, ipAddress)
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
		_, err := db.Exec("UPDATE registration_attempts SET attempts = ?, date = ? WHERE ip = ?", attempts+1, time.Now().Format(time.DateTime), ipAddress)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
		}

		return
	}
}

func GetAttempts(ip string) *Attempts {
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

	err = db.QueryRow("SELECT * FROM registration_attempts WHERE ip = ?", ip).Scan(
		&attempts.Email,
		&attempts.IpAddress,
		&attempts.Attempts,
		&date,
		&canBack)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
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

func ResetAttempts(ip string) {
	logger := logs.NewSistemLogger()

	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	defer db.Close()

	_, err = db.Exec("UPDATE registration_attempts SET attempts = ? WHERE ip = ?", 0, ip)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func RegistryPasswordToken(email, token string) {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
	defer db.Close()

	userUuid := GetUserUUID(email)

	_, err = db.Exec("UPDATE userinfo SET magic_password_id = ?, magic_password_expiration = ? WHERE uuid = ?",
		token,
		time.Now().Add(10*time.Minute).Format(time.DateTime),
		userUuid)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func IsValidPasswordToken(email, token string) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}
	defer db.Close()

	var dToken string
	var expirationStr string
	userUuid := GetUserUUID(email)

	err = db.QueryRow("SELECT magic_password_id, magic_password_expiration FROM userinfo WHERE uuid = ? ", userUuid).Scan(&dToken, &expirationStr)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	if dToken != token {
		return false
	}

	expiration, err := time.ParseInLocation(time.DateTime, expirationStr, time.Local)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	if expiration.Before(time.Now()) {
		logger.LogAndSendSystemMessage("Before")
		_, err = db.Exec("UPDATE userinfo SET magic_password_id = ? WHERE uuid = ?", "", userUuid)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
		}
		return false
	}

	return true
}

func ChangePassword(email, newPassword string) {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
	defer db.Close()

	newPassword = Encrypt(HashPassword(newPassword))

	_, err = db.Exec("UPDATE userdata SET password = ? WHERE email = ?", newPassword, email)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}

	userUuid := GetUserUUID(email)

	_, err = db.Exec("UPDATE userinfo SET magic_password_id = ? WHERE uuid = ?", "", userUuid)
}

func ComparePasswords(userId, jwtPassword string) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}
	defer db.Close()

	var dbPassword string

	err = db.QueryRow("SELECT password FROM userdata WHERE uuid = ?", userId).Scan(&dbPassword)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return Decrypt(dbPassword) == jwtPassword
}

type Services struct {
	Id     string
	Name   string
	Price  float64
	Status string
	Type   string
	Date   string
}

func HasServices(userId string) bool {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return false
	}
	defer db.Close()

	ok := UserExist(userId)

	if !ok {
		logger.LogAndSendSystemMessage("Usuário com id: " + userId + " não existe")
	}

	var dataStr string
	err = db.QueryRow("SELECT data FROM userinfo WHERE uuid = ?", userId).Scan(&dataStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}

		logger.LogAndSendSystemMessage(err.Error())
		return false
	}

	return true
}

func GetServices(userId string) []Services {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}
	defer db.Close()

	if !HasServices(userId) {
		return nil
	}

	var dataStr string
	err = db.QueryRow("SELECT data FROM userinfo WHERE uuid = ?", userId).Scan(&dataStr)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	var services []Services
	err = json.Unmarshal([]byte(dataStr), &services)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return nil
	}

	return services
}

func AddServices(userId string, service *Services) {
	logger := logs.NewSistemLogger()
	db, err := database.InitializeDB()
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
	defer db.Close()

	if !HasServices(userId) {
		_, err := db.Exec("UPDATE userinfo SET data = ? WHERE uuid = ?", "[]", userId)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}

	var bytes []byte

	servicesPointer := GetServices(userId)
	if servicesPointer == nil {
		bytes, err = json.Marshal([]Services{*service})
		if err != nil {
			logger.LogAndSendSystemMessage(err)
			return
		}
	} else {
		services := servicesPointer

		services = append(services, *service)
		bytes, err = json.Marshal(services)
		if err != nil {
			logger.LogAndSendSystemMessage(err.Error())
			return
		}
	}

	_, err = db.Exec("UPDATE userinfo SET data = ? WHERE uuid = ?", string(bytes), userId)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		return
	}
}

func RemoveServices() {}

func EditServices() {}

func GenerateInvoice() {
	
}