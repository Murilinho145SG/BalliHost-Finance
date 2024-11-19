package logs

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Logger struct {
	Id       string
	Messages []Message
}

type Message struct {
	Date string
	Msg  string
	User string
}

func GenerateMessageId() string {
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%012d", rand.Int63())
	return id
}

func NewLogger() *Logger {
	return &Logger{
		Id: GenerateMessageId(),
	}
}

func NewSistemLogger() *Logger {
	return &Logger{
		Id: time.Now().Format(time.DateOnly),
	}
}

func GetCallerInfos(layer int) (string, int, string, error) {
	if layer == 0 {
		layer = 1
	}

	pc, file, line, ok := runtime.Caller(layer)
	if !ok {
		return "", -1, "", errors.New("error when trying to capture the function that called")
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = funcName[strings.LastIndex(funcName, ".")+1:]

	fileName := file[strings.LastIndex(file, "/")+1:]

	return fileName, line, funcName, nil
}

func (l *Logger) LogSystemMessage(msg string) {
	file, line, funcName, err := GetCallerInfos(1)
	if err != nil {
		fmt.Println("Error when trying to get caller infos:", err)
		return
	}

	l.LogMessage(msg, fmt.Sprintf("[%s:%d %s]", file, line, funcName))
}

func (l *Logger) LogAndSendSystemMessage(msg string) {
	file, line, funcName, err := GetCallerInfos(2)
	if err != nil {
		fmt.Println("Error when trying to get caller infos:", err)
		return
	}

	l.LogMessage("\""+msg+"\"", fmt.Sprintf("[%s:%d %s]", file, line, funcName))
	l.SendSystemMessage()
	l.Messages = nil
}

func (l *Logger) SendSystemMessage() error {
	fmt.Println(l.Id)

	dir := "C:/logs"

	err := os.MkdirAll(dir, 0644)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, l.Id+".log")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, dataMsg := range l.Messages {
		lineMsg := fmt.Sprintf("[%s] %s: %s\n", dataMsg.Date, dataMsg.User, dataMsg.Msg)
		fmt.Println(lineMsg)

		_, err := f.WriteString(lineMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Logger) LogMessage(msg string, user string) {
	date := time.Now().Format(time.DateTime)
	message := Message{
		Date: date,
		Msg:  msg,
		User: user,
	}
	l.Messages = append(l.Messages, message)
}