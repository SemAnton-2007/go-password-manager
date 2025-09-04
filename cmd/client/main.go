package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"password-manager/internal/client"
	"password-manager/internal/common/crypto"
	"password-manager/internal/common/protocol"
)

var (
	version   = "1.0.0"
	buildDate = "2025-09-04"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("Password Manager Client\nVersion: %s\nBuild Date: %s\n", version, buildDate)
		return
	}

	runClient()
}

func runClient() {
	fmt.Println("=== Password Manager Client ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Введите адрес сервера [localhost]: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}

	fmt.Print("Введите порт сервера [8080]: ")
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	port := 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	cl := client.NewClient(host, port)
	defer cl.Close()

	fmt.Printf("Попытка подключения к %s:%d...\n", host, port)
	if err := cl.Connect(); err != nil {
		fmt.Printf("Ошибка подключения: %v\n", err)
		return
	}
	fmt.Println("Подключение успешно!")

	fmt.Println("\nВыберите тип пользователя:")
	fmt.Println("1. Новый пользователь")
	fmt.Println("2. Зарегистрированный пользователь")
	fmt.Print("Ваш выбор [1]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	var username, password string

	switch choice {
	case "1":
		fmt.Print("Введите логин: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			fmt.Println("Логин не может быть пустым")
			return
		}

		fmt.Print("Введите пароль: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Println("Пароль не может быть пустым")
			return
		}

		fmt.Println("Регистрируем пользователя...")
		if err := cl.Register(username, password); err != nil {
			fmt.Printf("Ошибка регистрации: %v\n", err)
			return
		}
		fmt.Println("Регистрация успешна!")

	case "2":
	default:
		fmt.Println("Неверный выбор")
		return
	}

	if !cl.IsAuthenticated() {
		fmt.Println("\n=== Авторизация ===")
		fmt.Print("Введите логин: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			fmt.Println("Логин не может быть пустым")
			return
		}

		fmt.Print("Введите пароль: ")
		password, _ = reader.ReadString('\n')
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Println("Пароль не может быть пустым")
			return
		}

		fmt.Println("Авторизуем пользователя...")
		if err := cl.Login(username, password); err != nil {
			fmt.Printf("Ошибка авторизации: %v\n", err)
			return
		}
		fmt.Println("Авторизация успешна!")
	}

	mainMenu(cl, reader)
}

func mainMenu(cl *client.Client, reader *bufio.Reader) {
	for {
		fmt.Printf("\n=== Главное меню (пользователь: %s) ===\n", cl.GetUsername())
		fmt.Println("1. Показать мои данные")
		fmt.Println("2. Создать новый элемент")
		fmt.Println("3. Выйти")
		fmt.Print("Ваш выбор [3]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if choice == "" {
			choice = "3"
		}

		switch choice {
		case "1":
			showData(cl, reader)
		case "2":
			createNewItem(cl, reader)
		case "3":
			fmt.Println("Выход...")
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

func showData(cl *client.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Мои данные ===")
	fmt.Println("Синхронизируем данные...")

	items, err := cl.SyncData(time.Time{})
	if err != nil {
		fmt.Printf("Ошибка синхронизации: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("У вас пока нет сохраненных данных")
		return
	}

	fmt.Printf("Найдено %d элементов:\n", len(items))
	for i, item := range items {
		fmt.Printf("%d. %s (%s)\n", i+1, item.Name, getDataTypeName(item.Type))

		decryptedData, err := decryptItemData(item, cl.GetUsername())
		if err != nil {
			fmt.Printf("   Ошибка декодирования: %v\n", err)
		} else {
			fmt.Printf("   Данные: %s\n", string(decryptedData))
		}

		fmt.Printf("   Обновлен: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	fmt.Print("Нажмите Enter для возврата...")
	reader.ReadString('\n')
}

func createNewItem(cl *client.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Создание нового элемента ===")

	fmt.Println("Выберите тип данных:")
	fmt.Println("1. Логин/Пароль")
	fmt.Println("2. Текстовые данные")
	fmt.Println("3. Банковская карта")
	fmt.Print("Ваш выбор [1]: ")

	typeChoice, _ := reader.ReadString('\n')
	typeChoice = strings.TrimSpace(typeChoice)
	if typeChoice == "" {
		typeChoice = "1"
	}

	var dataType uint8
	switch typeChoice {
	case "1":
		dataType = protocol.DataTypeLoginPassword
	case "2":
		dataType = protocol.DataTypeText
	case "3":
		dataType = protocol.DataTypeBankCard
	default:
		fmt.Println("Неверный выбор типа данных")
		return
	}

	fmt.Print("Введите название элемента: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" {
		fmt.Println("Название не может быть пустым")
		return
	}

	var data string
	switch dataType {
	case protocol.DataTypeLoginPassword:
		fmt.Print("Введите логин: ")
		login, _ := reader.ReadString('\n')
		login = strings.TrimSpace(login)

		fmt.Print("Введите пароль: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		loginData := map[string]string{
			"login":    login,
			"password": password,
		}
		jsonData, _ := json.Marshal(loginData)
		data = string(jsonData)

	case protocol.DataTypeText:
		fmt.Print("Введите текст: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		data = text

	case protocol.DataTypeBankCard:
		fmt.Print("Введите номер карта: ")
		cardNumber, _ := reader.ReadString('\n')
		cardNumber = strings.TrimSpace(cardNumber)

		fmt.Print("Введите срок действия (MM/YY): ")
		expiry, _ := reader.ReadString('\n')
		expiry = strings.TrimSpace(expiry)

		fmt.Print("Введите CVV: ")
		cvv, _ := reader.ReadString('\n')
		cvv = strings.TrimSpace(cvv)

		fmt.Print("Введите имя владельца: ")
		holder, _ := reader.ReadString('\n')
		holder = strings.TrimSpace(holder)

		cardData := map[string]string{
			"number": cardNumber,
			"expiry": expiry,
			"cvv":    cvv,
			"holder": holder,
		}
		jsonData, _ := json.Marshal(cardData)
		data = string(jsonData)
	}

	encryptedData, err := encryptData([]byte(data), cl.GetUsername())
	if err != nil {
		fmt.Printf("Ошибка шифрования данных: %v\n", err)
		return
	}

	item := protocol.NewDataItem{
		Type:     dataType,
		Name:     name,
		Data:     encryptedData,
		Metadata: make(map[string]string),
	}

	fmt.Println("Сохраняем данные на сервере...")
	if err := cl.SaveData(item); err != nil {
		fmt.Printf("Ошибка сохранения: %v\n", err)
		return
	}

	fmt.Println("Данные успешно сохранены!")
}

func encryptData(data []byte, password string) ([]byte, error) {
	key := deriveSimpleKey(password)
	return crypto.Encrypt(data, key)
}

func decryptItemData(item protocol.DataItem, password string) ([]byte, error) {
	key := deriveSimpleKey(password)
	return crypto.Decrypt(item.Data, key)
}

func deriveSimpleKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func getDataTypeName(dataType uint8) string {
	switch dataType {
	case protocol.DataTypeLoginPassword:
		return "Логин/Пароль"
	case protocol.DataTypeText:
		return "Текст"
	case protocol.DataTypeBinary:
		return "Бинарные данные"
	case protocol.DataTypeBankCard:
		return "Банковская карта"
	default:
		return "Неизвестный тип"
	}
}
