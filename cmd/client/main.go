package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	fmt.Printf("\nНайдено %d элементов:\n", len(items))
	for i, item := range items {
		fmt.Printf("%d. %s (%s)\n", i+1, item.Name, getDataTypeName(item.Type))
	}

	fmt.Println("\nДействия:")
	fmt.Println("0. Вернуться назад")
	fmt.Println("1-9. Показать детали элемента")
	fmt.Print("Ваш выбор [0]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" || choice == "0" {
		return
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(items) {
		fmt.Println("Неверный выбор элемента")
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	showItemDetails(items[index-1], cl.GetUsername(), reader, cl)
}

func showItemDetails(item protocol.DataItem, password string, reader *bufio.Reader, cl *client.Client) {
	fmt.Printf("\n=== Детали элемента: %s ===\n", item.Name)
	fmt.Printf("Тип: %s\n", getDataTypeName(item.Type))
	fmt.Printf("Создан: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Обновлен: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

	switch item.Type {
	case protocol.DataTypeBinary:
		// Для бинарных данных показываем информацию о файле
		fmt.Println("\n--- Информация о файле ---")
		if filename, ok := item.Metadata[protocol.MetaOriginalFileName]; ok {
			fmt.Printf("Имя файла: %s\n", filename)
		}
		if size, ok := item.Metadata[protocol.MetaFileSize]; ok {
			fmt.Printf("Размер: %s байт\n", size)
		}
		if ext, ok := item.Metadata[protocol.MetaFileExtension]; ok {
			fmt.Printf("Расширение: %s\n", ext)
		}

	default:
		// Для других типов данных дешифруем и показываем содержимое
		decryptedData, err := decryptItemData(item, password)
		if err != nil {
			fmt.Printf("Ошибка декодирования: %v\n", err)
			fmt.Print("Нажмите Enter для возврата...")
			reader.ReadString('\n')
			return
		}

		switch item.Type {
		case protocol.DataTypeLoginPassword:
			var loginData map[string]string
			if err := json.Unmarshal(decryptedData, &loginData); err == nil {
				fmt.Println("\n--- Учетные данные ---")
				fmt.Printf("Логин: %s\n", loginData["login"])
				fmt.Printf("Пароль: %s\n", loginData["password"])
			} else {
				fmt.Printf("Данные: %s\n", string(decryptedData))
			}

		case protocol.DataTypeText:
			fmt.Println("\n--- Текстовые данные ---")
			fmt.Println(string(decryptedData))

		case protocol.DataTypeBankCard:
			var cardData map[string]string
			if err := json.Unmarshal(decryptedData, &cardData); err == nil {
				fmt.Println("\n--- Данные банковской карты ---")
				fmt.Printf("Номер карты: %s\n", cardData["number"])
				fmt.Printf("Срок действия: %s\n", cardData["expiry"])
				fmt.Printf("CVV: %s\n", cardData["cvv"])
				fmt.Printf("Имя владельца: %s\n", cardData["holder"])
			} else {
				fmt.Printf("Данные: %s\n", string(decryptedData))
			}
		}
	}

	fmt.Println("\nДействия:")
	fmt.Println("0. Вернуться назад")
	fmt.Println("1. Удалить элемент")
	if item.Type == protocol.DataTypeBinary {
		fmt.Println("2. Скачать файл")
	} else {
		fmt.Println("2. Редактировать элемент")
	}
	fmt.Print("Ваш выбор [0]: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" || choice == "0" {
		return
	}

	switch choice {
	case "1":
		deleteItem(item.ID, cl, reader)
	case "2":
		if item.Type == protocol.DataTypeBinary {
			downloadFile(item, cl, reader)
		} else {
			editItem(item, password, cl, reader)
		}
	default:
		fmt.Println("Неверный выбор")
	}
}

func downloadFile(item protocol.DataItem, cl *client.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Скачивание файла ===")

	fmt.Println("Загружаем файл...")
	fileData, err := cl.DownloadData(item.ID)
	if err != nil {
		fmt.Printf("Ошибка загрузки: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	decryptedData, err := decryptData(fileData, cl.GetUsername())
	if err != nil {
		fmt.Printf("Ошибка расшифровки: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	originalName, ok := item.Metadata[protocol.MetaOriginalFileName]
	if !ok {
		originalName = item.Name
	}

	fmt.Printf("Введите путь для сохранения файла [./%s]: ", originalName)
	savePath, _ := reader.ReadString('\n')
	savePath = strings.TrimSpace(savePath)
	if savePath == "" {
		savePath = "./" + originalName
	}

	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Ошибка создания директории: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	if err := ioutil.WriteFile(savePath, decryptedData, 0644); err != nil {
		fmt.Printf("Ошибка сохранения файла: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	fmt.Printf("Файл успешно сохранен: %s (%d байт)\n", savePath, len(decryptedData))
	fmt.Print("Нажмите Enter для продолжения...")
	reader.ReadString('\n')
}

func editItem(item protocol.DataItem, password string, cl *client.Client, reader *bufio.Reader) {
	fmt.Printf("\n=== Редактирование элемента: %s ===\n", item.Name)

	decryptedData, err := decryptItemData(item, password)
	if err != nil {
		fmt.Printf("Ошибка декодирования: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		reader.ReadString('\n')
		return
	}

	var newData string

	switch item.Type {
	case protocol.DataTypeLoginPassword:
		var loginData map[string]string
		if err := json.Unmarshal(decryptedData, &loginData); err == nil {
			fmt.Printf("Текущий логин [%s]: ", loginData["login"])
			login, _ := reader.ReadString('\n')
			login = strings.TrimSpace(login)
			if login != "" {
				loginData["login"] = login
			}

			fmt.Printf("Текущий пароль [%s]: ", loginData["password"])
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)
			if password != "" {
				loginData["password"] = password
			}

			jsonData, _ := json.Marshal(loginData)
			newData = string(jsonData)
		}

	case protocol.DataTypeText:
		fmt.Printf("Текущий текст:\n%s\n", string(decryptedData))
		fmt.Println("Введите новый текст (оставьте пустым для отмены):")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != "" {
			newData = text
		} else {
			fmt.Println("Редактирование отменено")
			return
		}

	case protocol.DataTypeBankCard:
		var cardData map[string]string
		if err := json.Unmarshal(decryptedData, &cardData); err == nil {
			fmt.Printf("Текущий номер карты [%s]: ", cardData["number"])
			number, _ := reader.ReadString('\n')
			number = strings.TrimSpace(number)
			if number != "" {
				cardData["number"] = number
			}

			fmt.Printf("Текущий срок действия [%s]: ", cardData["expiry"])
			expiry, _ := reader.ReadString('\n')
			expiry = strings.TrimSpace(expiry)
			if expiry != "" {
				cardData["expiry"] = expiry
			}

			fmt.Printf("Текущий CVV [%s]: ", cardData["cvv"])
			cvv, _ := reader.ReadString('\n')
			cvv = strings.TrimSpace(cvv)
			if cvv != "" {
				cardData["cvv"] = cvv
			}

			fmt.Printf("Текущий владелец [%s]: ", cardData["holder"])
			holder, _ := reader.ReadString('\n')
			holder = strings.TrimSpace(holder)
			if holder != "" {
				cardData["holder"] = holder
			}

			jsonData, _ := json.Marshal(cardData)
			newData = string(jsonData)
		}

	default:
		fmt.Println("Редактирование данного типа данных не поддерживается")
		return
	}

	if newData == "" {
		fmt.Println("Данные не изменены")
		return
	}

	encryptedData, err := encryptData([]byte(newData), cl.GetUsername())
	if err != nil {
		fmt.Printf("Ошибка шифрования данных: %v\n", err)
		return
	}

	updatedItem := protocol.NewDataItem{
		Type:     item.Type,
		Name:     item.Name,
		Data:     encryptedData,
		Metadata: item.Metadata,
	}

	fmt.Println("Обновляем данные на сервере...")
	if err := cl.UpdateData(item.ID, updatedItem); err != nil {
		fmt.Printf("Ошибка обновления: %v\n", err)
	} else {
		fmt.Println("Данные успешно обновлены!")
	}

	fmt.Print("Нажмите Enter для продолжения...")
	reader.ReadString('\n')
}

func deleteItem(itemID string, cl *client.Client, reader *bufio.Reader) {
	fmt.Print("\nВы уверены, что хотите удалить этот элемент? (y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Удаление отменено")
		return
	}

	fmt.Println("Удаляем элемент...")
	err := cl.DeleteData(itemID)
	if err != nil {
		fmt.Printf("Ошибка удаления: %v\n", err)
	} else {
		fmt.Println("Элемент успешно удален!")
	}

	fmt.Print("Нажмите Enter для продолжения...")
	reader.ReadString('\n')
}

func createNewItem(cl *client.Client, reader *bufio.Reader) {
	fmt.Println("\n=== Создание нового элемента ===")

	fmt.Println("Выберите тип данных:")
	fmt.Println("1. Логин/Пароль")
	fmt.Println("2. Текстовые данные")
	fmt.Println("3. Бинарные данные (файл)")
	fmt.Println("4. Банковская карта")
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
		dataType = protocol.DataTypeBinary
	case "4":
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

	var data []byte
	var metadata map[string]string = make(map[string]string)

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
		data = jsonData

	case protocol.DataTypeText:
		fmt.Print("Введите текст: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		data = []byte(text)

	case protocol.DataTypeBinary:
		fmt.Print("Введите путь к файлу: ")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			fmt.Println("Путь к файлу не может быть пустым")
			return
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Ошибка получения информации о файле: %v\n", err)
			return
		}

		if fileInfo.Size() > 500*1024 {
			fmt.Printf("Файл слишком большой (%d bytes). Максимальный размер: 500КB\n", fileInfo.Size())
			return
		}

		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Ошибка чтения файла: %v\n", err)
			return
		}

		data = fileData

		metadata[protocol.MetaOriginalFileName] = filepath.Base(filePath)
		metadata[protocol.MetaFileSize] = fmt.Sprintf("%d", len(fileData))
		metadata[protocol.MetaFileExtension] = filepath.Ext(filePath)

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
		data = jsonData
	}

	encryptedData, err := encryptData(data, cl.GetUsername())
	if err != nil {
		fmt.Printf("Ошибка шифрования данных: %v\n", err)
		return
	}

	item := protocol.NewDataItem{
		Type:     dataType,
		Name:     name,
		Data:     encryptedData,
		Metadata: metadata,
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

func decryptData(data []byte, password string) ([]byte, error) {
	key := deriveSimpleKey(password)
	return crypto.Decrypt(data, key)
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
