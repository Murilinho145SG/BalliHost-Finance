package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prodata/database/account"
	"prodata/emailHandler"
	"prodata/logs"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.WriteHeader(http.StatusOK)
		return
	}

	var user account.DataUserRegistry
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return
	}

	data := account.Data{
		Device:              r.Header.Get("User-Agent"),
		IPAddress:           r.RemoteAddr,
		Auth:                false,
		MagicLinkId:         "",
		MagicLinkVerified:   false,
		MagicLinkExpiration: time.Now(),
	}

	ok := CheckErrorsRegister(&user, w, &data)

	if ok {
		emailHandler.SendMagicLinkVerification(user.Email, r.Header.Get("User-Agent"))
	}

}

func ValidCpf(user *account.DataUserRegistry) bool {
	re := regexp.MustCompile(`[^\d]`)
	cpf := re.ReplaceAllString(user.CPF, "")

	if len(cpf) != 11 || strings.Count(cpf, string(cpf[0])) == 11 {
		return false
	}

	var plus int
	for i := 0; i < 9; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		plus += num * (10 - i)
	}
	first := (plus * 10 % 11) % 10

	plus = 0
	for i := 0; i < 10; i++ {
		num, _ := strconv.Atoi(string(cpf[i]))
		plus += num * (11 - i)
	}
	second := (plus * 10 % 11) % 10

	return cpf[9] == byte(first+'0') && cpf[10] == byte(second+'0')
}

func ValidEmail(user *account.DataUserRegistry) bool {
	return regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).MatchString(user.Email)
}

func ValidPhone(user *account.DataUserRegistry) bool {
	re := regexp.MustCompile(`^\(\d{2}\)\s\d{4,5}-\d{4}$`)
	return re.MatchString(user.Phone)
}

func ValidPassword(user *account.DataUserRegistry) bool {
	if user.Password == "" {
		return false
	}

	if len(user.Password) < 8 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(user.Password)

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(user.Password)

	hasDigit := regexp.MustCompile(`\d`).MatchString(user.Password)

	hasSpecial := regexp.MustCompile(`[@$!%*?&]`).MatchString(user.Password)

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func ValidBirthdate(user *account.DataUserRegistry) bool {
	parse, err := time.Parse(time.DateOnly, user.Birthdate)
	now := time.Now()

	if err != nil || parse.After(now) {
		return false
	}

	timeAge := now.Year() - parse.Year()
	if now.Month() < parse.Month() || (now.Month() == parse.Month() && now.Day() < parse.Day()) {
		timeAge--
	}

	if timeAge < 18 {
		return false
	}

	return true
}

func ValidNames(user *account.DataUserRegistry) bool {
	if user.FirstName == "" || user.LastName == "" {
		return false
	}

	reg := regexp.MustCompile(`\d`)

	if reg.MatchString(user.FirstName) {
		return false
	}

	if reg.MatchString(user.LastName) {
		return false
	}

	return true
}

type ErrorRegisters struct {
	Errors map[string][]interface{} `json:"errors"`
}

func (s *ErrorRegisters) ErrorsRegister(msg string) *ErrorRegisters {
	if s.Errors == nil {
		s.Errors = make(map[string][]interface{})
	}

	s.Errors["error"] = append(s.Errors["error"], map[string]string{
		"message": msg,
	})

	return s
}

func CheckErrorsRegister(user *account.DataUserRegistry, w http.ResponseWriter, data *account.Data) bool {
	regErr := ErrorRegisters{}
	fields := []string{
		user.FirstName, user.LastName, user.Email, user.Password, user.Birthdate, user.CPF, user.Phone, user.Address, user.City, user.State, user.ZipCode, user.Country,
	}

	for i, field := range fields {
		if field == "" {
			fmt.Println(i)
			regErr.ErrorsRegister("Preencha todos os campos obrigatórios")
			break
		}
	}

	valid := ValidNames(user)

	if !valid {
		regErr.ErrorsRegister("Nomes inválidos. Não podem conter números ou caracteres especiais")
	}

	valid = ValidEmail(user)

	if !valid {
		regErr.ErrorsRegister("Email inválido")
	}

	valid = ValidPassword(user)

	if !valid {
		regErr.ErrorsRegister("A senha deve conter uma letra maiúscula, uma minúscula, um número ou um caractere especial e possuir no mínimo 8 caracteres")
	}

	valid = ValidBirthdate(user)

	if !valid {
		regErr.ErrorsRegister("Data de nascimento inválida ou é menor de 18 anos")
	}

	valid = ValidPhone(user)

	if !valid {
		regErr.ErrorsRegister("Número de telefone inválido")
	}

	valid = ValidCpf(user)

	if !valid {
		regErr.ErrorsRegister("CPF inválido")
	}

	if len(regErr.Errors) > 0 {
		err := json.NewEncoder(w).Encode(regErr.Errors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
			return false
		}
	} else {
		exist := account.UserExistFromEmail(user.Email)

		if !exist {
			account.CreateUser(user, data)
			w.WriteHeader(http.StatusCreated)

			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"success": "Added successfully",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
				return false
			}

			return true

		} else {
			w.WriteHeader(http.StatusBadRequest)

			err := json.NewEncoder(w).Encode(map[string]interface{}{
				"registry": map[string]interface{}{
					"message": "Email already registered",
				},
			})

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
				return false
			}
		}
	}

	return false
}

//I need a make a login auth
//This Handler is so good

func HandlerMagicLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	logger := logs.NewSistemLogger()

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		w.WriteHeader(http.StatusOK)
		return
	}

	id := r.URL.Path[len("/account/verify/"):]

	var email map[string]string

	err := json.NewDecoder(r.Body).Decode(&email)
	if err != nil {
		logger.LogAndSendSystemMessage(err.Error())
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	userUuid := account.GetUserUUID(email["email"])
	ok, token := account.IsValidMagicLink(userUuid, id)

	if ok {
		if token == "retry" {
			w.WriteHeader(http.StatusOK)
			emailHandler.SendMagicLinkVerification(email["email"], r.Header.Get("User-Agent"))
		} else {
			//.WriteHeader(http.StatusCreated)
			account.ValidMagicLink(r.Header.Get("User-Agent"), userUuid)
			tokenJwt := account.GenerateJWT(userUuid)

			tokenJson := map[string]string{
				"token": tokenJwt,
			}

			err = json.NewEncoder(w).Encode(tokenJson)
			if err != nil {
				logger.LogAndSendSystemMessage(err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}