// Package client предоставляет клиентскую библиотеку и пользовательский интерфейс для менеджера паролей.
//
// Клиентский UI включает:
// - Интерактивный консольный интерфейс для пользователей
// - Управление аутентификацией и регистрацией
// - Просмотр и редактирование элементов данных
// - Создание новых записей различных типов
// - Работу с файлами и бинарными данными
// - Управление метаинформацией элементов
//
// UI построен на основе базового клиента и предоставляет:
// - Пошаговые wizard'ы для создания данных
// - Валидацию вводимых данных
// - Шифрование/дешифрование данных на лету
// - Визуализацию различных типов данных
//
// Пример использования:
//
//	uiClient := client.NewUIClient("localhost", 8080)
//	err := uiClient.Run()
package client

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"password-manager/internal/common/crypto"
	"password-manager/internal/common/protocol"
)

// UIClient представляет клиент с пользовательским интерфейсом.
// Наследует функциональность базового Client и добавляет интерактивные возможности.
type UIClient struct {
	*Client
	reader *bufio.Reader
}

// NewUIClient создает новый экземпляр UI-клиента.
//
// Parameters:
//
//	host - хост сервера для подключения
//	port - порт сервера для подключения
//
// Returns:
//
//	*UIClient - новый экземпляр UI-клиента
func NewUIClient(host string, port int) *UIClient {
	return &UIClient{
		Client: NewClient(host, port),
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run запускает интерактивный клиентский интерфейс.
//
// Process:
//  1. Запрашивает параметры подключения у пользователя
//  2. Устанавливает соединение с сервером
//  3. Выполняет аутентификацию или регистрацию
//  4. Запускает главное меню
func (c *UIClient) Run() error {
	log.Println("=== Password Manager Client ===")

	// Запрос параметров подключения
	fmt.Print("Введите адрес сервера [localhost]: ")
	host, _ := c.reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}

	fmt.Print("Введите порт сервера [8080]: ")
	portStr, _ := c.reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	port := 8080
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	// Обновляем хост и порт клиента
	c.Client = NewClient(host, port)
	defer c.Client.Close()

	log.Printf("Попытка подключения к %s:%d...\n", host, port)
	if err := c.Connect(); err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}
	log.Println("Подключение успешно!")

	// Аутентификация/регистрация
	if err := c.handleAuth(); err != nil {
		return err
	}

	// Главное меню
	c.mainMenu()
	return nil
}

// handleAuth обрабатывает аутентификацию пользователя
func (c *UIClient) handleAuth() error {
	fmt.Println("\nВыберите тип пользователя:")
	fmt.Println("1. Новый пользователь")
	fmt.Println("2. Зарегистрированный пользователь")
	fmt.Print("Ваш выбор [1]: ")

	choice, _ := c.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1":
		return c.handleRegistration()
	case "2":
		return c.handleLogin()
	default:
		return fmt.Errorf("неверный выбор")
	}
}

// handleRegistration обрабатывает регистрацию нового пользователя
func (c *UIClient) handleRegistration() error {
	fmt.Print("Введите логин: ")
	username, _ := c.reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("логин не может быть пустым")
	}

	fmt.Print("Введите пароль: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("пароль не может быть пустым")
	}

	log.Println("Регистрируем пользователя...")
	if err := c.Register(username, password); err != nil {
		return fmt.Errorf("ошибка регистрации: %v", err)
	}
	log.Println("Регистрация успешна!")

	return c.handleLoginWithCredentials(username, password)
}

// handleLogin обрабатывает вход существующего пользователя
func (c *UIClient) handleLogin() error {
	fmt.Println("\n=== Авторизация ===")
	fmt.Print("Введите логин: ")
	username, _ := c.reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("логин не может быть пустым")
	}

	fmt.Print("Введите пароль: ")
	password, _ := c.reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("пароль не может быть пустым")
	}

	return c.handleLoginWithCredentials(username, password)
}

// handleLoginWithCredentials выполняет авторизацию
func (c *UIClient) handleLoginWithCredentials(username, password string) error {
	log.Println("Авторизуем пользователя...")
	if err := c.Login(username, password); err != nil {
		return fmt.Errorf("ошибка авторизации: %v", err)
	}
	log.Println("Авторизация успешна!")
	return nil
}

// mainMenu отображает главное меню и обрабатывает выбор пользователя.
//
// Parameters:
//
//	cl     - клиент для выполнения операций
//	reader - reader для ввода пользователя
//
// Menu options:
//  1. Показать мои данные
//  2. Создать новый элемент
//  3. Выйти
func (c *UIClient) mainMenu() {
	for {
		fmt.Printf("\n=== Главное меню (пользователь: %s) ===\n", c.GetUsername())
		fmt.Println("1. Показать мои данные")
		fmt.Println("2. Создать новый элемент")
		fmt.Println("3. Выйти")
		fmt.Print("Ваш выбор [3]: ")

		choice, _ := c.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if choice == "" {
			choice = "3"
		}

		switch choice {
		case "1":
			c.showData()
		case "2":
			c.createNewItem()
		case "3":
			log.Println("Выход...")
			return
		default:
			fmt.Println("Неверный выбор")
		}
	}
}

// showData отображает список данных пользователя с возможностью выбора.
//
// Process:
//   - Загружает и отображает список элементов
//   - Предлагает выбрать элемент для детального просмотра
//   - Обрабатывает действия пользователя
func (c *UIClient) showData() {
	fmt.Println("\n=== Мои данные ===")
	log.Println("Синхронизируем данные...")

	items, err := c.SyncData(time.Time{})
	if err != nil {
		log.Printf("Ошибка синхронизации: %v\n", err)
		return
	}

	if len(items) == 0 {
		fmt.Println("У вас пока нет сохраненных данных")
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
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

	choice, _ := c.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" || choice == "0" {
		return
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(items) {
		fmt.Println("Неверный выбор элемента")
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	c.showItemDetails(items[index-1])
}

// showItemDetails отображает детальную информацию об элементе данных.
func (c *UIClient) showItemDetails(item protocol.DataItem) {
	fmt.Printf("\n=== Детали элемента: %s ===\n", item.Name)
	fmt.Printf("Тип: %s\n", getDataTypeName(item.Type))
	fmt.Printf("Создан: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Обновлен: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

	if len(item.Metadata) > 0 {
		fmt.Println("\n--- Метаинформация ---")
		for key, value := range item.Metadata {
			if item.Type == protocol.DataTypeBinary &&
				(key == protocol.MetaOriginalFileName ||
					key == protocol.MetaFileSize ||
					key == protocol.MetaFileExtension) {
				continue
			}
			fmt.Printf("%s: %s\n", key, value)
		}
	}

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
		decryptedData, err := c.decryptItemData(item)
		if err != nil {
			log.Printf("Ошибка декодирования: %v\n", err)
			fmt.Print("Нажмите Enter для возврата...")
			c.reader.ReadString('\n')
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

	choice, _ := c.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	if choice == "" || choice == "0" {
		return
	}

	switch choice {
	case "1":
		c.deleteItem(item.ID)
	case "2":
		if item.Type == protocol.DataTypeBinary {
			c.downloadFile(item)
		} else {
			c.editItem(item)
		}
	default:
		fmt.Println("Неверный выбор")
	}
}

// downloadFile обрабатывает скачивание и сохранение бинарного файла.
//
// Process:
//   - Загружает данные файла с сервера
//   - Дешифрует данные
//   - Сохраняет файл в указанное место
func (c *UIClient) downloadFile(item protocol.DataItem) {
	fmt.Println("\n=== Скачивание файла ===")

	log.Println("Загружаем файл...")
	fileData, err := c.DownloadData(item.ID)
	if err != nil {
		log.Printf("Ошибка загрузки: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	decryptedData, err := c.decryptData(fileData)
	if err != nil {
		log.Printf("Ошибка расшифровки: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	originalName, ok := item.Metadata[protocol.MetaOriginalFileName]
	if !ok {
		originalName = item.Name
	}

	fmt.Printf("Введите путь для сохранения файла [./%s]: ", originalName)
	savePath, _ := c.reader.ReadString('\n')
	savePath = strings.TrimSpace(savePath)
	if savePath == "" {
		savePath = "./" + originalName
	}

	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Ошибка создания директории: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	if err := ioutil.WriteFile(savePath, decryptedData, 0644); err != nil {
		log.Printf("Ошибка сохранения файла: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	log.Printf("Файл успешно сохранен: %s (%d байт)\n", savePath, len(decryptedData))
	fmt.Print("Нажмите Enter для продолжения...")
	c.reader.ReadString('\n')
}

// editItem предоставляет интерфейс редактирования элемента данных.
func (c *UIClient) editItem(item protocol.DataItem) {
	fmt.Printf("\n=== Редактирование элемента: %s ===\n", item.Name)

	decryptedData, err := c.decryptItemData(item)
	if err != nil {
		log.Printf("Ошибка декодирования: %v\n", err)
		fmt.Print("Нажмите Enter для возврата...")
		c.reader.ReadString('\n')
		return
	}

	var newData string
	var updatedItem protocol.NewDataItem

	switch item.Type {
	case protocol.DataTypeLoginPassword:
		var loginData map[string]string
		if err := json.Unmarshal(decryptedData, &loginData); err == nil {
			fmt.Printf("Текущий логин [%s]: ", loginData["login"])
			login, _ := c.reader.ReadString('\n')
			login = strings.TrimSpace(login)
			if login != "" {
				loginData["login"] = login
			}

			fmt.Printf("Текущий пароль [%s]: ", loginData["password"])
			password, _ := c.reader.ReadString('\n')
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
		text, _ := c.reader.ReadString('\n')
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
			number, _ := c.reader.ReadString('\n')
			number = strings.TrimSpace(number)
			if number != "" {
				cardData["number"] = number
			}

			fmt.Printf("Текущий срок действия [%s]: ", cardData["expiry"])
			expiry, _ := c.reader.ReadString('\n')
			expiry = strings.TrimSpace(expiry)
			if expiry != "" {
				cardData["expiry"] = expiry
			}

			fmt.Printf("Текущий CVV [%s]: ", cardData["cvv"])
			cvv, _ := c.reader.ReadString('\n')
			cvv = strings.TrimSpace(cvv)
			if cvv != "" {
				cardData["cvv"] = cvv
			}

			fmt.Printf("Текущий владелец [%s]: ", cardData["holder"])
			holder, _ := c.reader.ReadString('\n')
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

	updatedMetadata := make(map[string]string)
	for k, v := range item.Metadata {
		updatedMetadata[k] = v
	}

	fmt.Println("\n--- Редактирование метаинформации ---")
	if len(updatedMetadata) > 0 {
		fmt.Println("Текущая метаинформация:")
		for key, value := range updatedMetadata {
			if item.Type == protocol.DataTypeBinary &&
				(key == protocol.MetaOriginalFileName ||
					key == protocol.MetaFileSize ||
					key == protocol.MetaFileExtension) {
				continue
			}
			fmt.Printf("  %s: %s\n", key, value)
		}
	} else {
		fmt.Println("Метаинформация отсутствует")
	}

	fmt.Println("\nДействия с метаинформацией:")
	fmt.Println("1. Добавить новое поле")
	if len(updatedMetadata) > 0 {
		fmt.Println("2. Удалить поле")
		fmt.Println("3. Редактировать поле")
	}
	fmt.Println("0. Пропустить редактирование метаинформации")
	fmt.Print("Ваш выбор [0]: ")

	metaChoice, _ := c.reader.ReadString('\n')
	metaChoice = strings.TrimSpace(metaChoice)

	switch metaChoice {
	case "1":
		fmt.Print("Введите имя нового поля: ")
		fieldName, _ := c.reader.ReadString('\n')
		fieldName = strings.TrimSpace(fieldName)

		if fieldName != "" {
			fmt.Print("Введите значение поля: ")
			fieldValue, _ := c.reader.ReadString('\n')
			fieldValue = strings.TrimSpace(fieldValue)

			updatedMetadata[fieldName] = fieldValue
			fmt.Printf("Добавлено поле: %s = %s\n", fieldName, fieldValue)
		}

	case "2":
		if len(updatedMetadata) > 0 {
			fmt.Print("Введите имя поля для удаления: ")
			fieldName, _ := c.reader.ReadString('\n')
			fieldName = strings.TrimSpace(fieldName)

			if _, exists := updatedMetadata[fieldName]; exists {
				delete(updatedMetadata, fieldName)
				fmt.Printf("Поле '%s' удалено\n", fieldName)
			} else {
				fmt.Printf("Поле '%s' не найдено\n", fieldName)
			}
		}

	case "3":
		if len(updatedMetadata) > 0 {
			fmt.Print("Введите имя поля для редактирования: ")
			fieldName, _ := c.reader.ReadString('\n')
			fieldName = strings.TrimSpace(fieldName)

			if currentValue, exists := updatedMetadata[fieldName]; exists {
				fmt.Printf("Текущее значение '%s': %s\n", fieldName, currentValue)
				fmt.Print("Введите новое значение: ")
				fieldValue, _ := c.reader.ReadString('\n')
				fieldValue = strings.TrimSpace(fieldValue)

				updatedMetadata[fieldName] = fieldValue
				fmt.Printf("Поле '%s' обновлено: %s\n", fieldName, fieldValue)
			} else {
				fmt.Printf("Поле '%s' не найдено\n", fieldName)
			}
		}
	}

	if newData == "" {
		newData = string(decryptedData)
	}

	encryptedData, err := c.encryptData([]byte(newData))
	if err != nil {
		log.Printf("Ошибка шифрования данных: %v\n", err)
		return
	}

	updatedItem = protocol.NewDataItem{
		Type:     item.Type,
		Name:     item.Name,
		Data:     encryptedData,
		Metadata: updatedMetadata,
	}

	log.Println("Обновляем данные на сервере...")
	if err := c.UpdateData(item.ID, updatedItem); err != nil {
		log.Printf("Ошибка обновления: %v\n", err)
	} else {
		log.Println("Данные успешно обновлены!")
	}

	fmt.Print("Нажмите Enter для продолжения...")
	c.reader.ReadString('\n')
}

// deleteItem удаляет элемент данных
func (c *UIClient) deleteItem(itemID string) {
	fmt.Print("\nВы уверены, что хотите удалить этот элемент? (y/N): ")
	confirm, _ := c.reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("Удаление отменено")
		return
	}

	log.Println("Удаляем элемент...")
	err := c.DeleteData(itemID)
	if err != nil {
		log.Printf("Ошибка удаления: %v\n", err)
	} else {
		log.Println("Элемент успешно удален!")
	}

	fmt.Print("Нажмите Enter для продолжения...")
	c.reader.ReadString('\n')
}

// createNewItem интерактивно создает новый элемент данных.
//
// Process:
//   - Запрашивает тип данных
//   - Запрашивает основные поля в зависимости от типа
//   - Предлагает добавить метаданные
//   - Шифрует и сохраняет данные
func (c *UIClient) createNewItem() {
	fmt.Println("\n=== Создание нового элемента ===")

	fmt.Println("Выберите тип данных:")
	fmt.Println("1. Логин/Пароль")
	fmt.Println("2. Текстовые данные")
	fmt.Println("3. Бинарные данные (файл)")
	fmt.Println("4. Банковская карта")
	fmt.Print("Ваш выбор [1]: ")

	typeChoice, _ := c.reader.ReadString('\n')
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
	name, _ := c.reader.ReadString('\n')
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
		login, _ := c.reader.ReadString('\n')
		login = strings.TrimSpace(login)

		fmt.Print("Введите пароль: ")
		password, _ := c.reader.ReadString('\n')
		password = strings.TrimSpace(password)

		loginData := map[string]string{
			"login":    login,
			"password": password,
		}
		jsonData, _ := json.Marshal(loginData)
		data = jsonData

	case protocol.DataTypeText:
		fmt.Print("Введите текст: ")
		text, _ := c.reader.ReadString('\n')
		text = strings.TrimSpace(text)
		data = []byte(text)

	case protocol.DataTypeBinary:
		fmt.Print("Введите путь к файлу: ")
		filePath, _ := c.reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			fmt.Println("Путь к файлу не может быть пустым")
			return
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Ошибка получения информации о файле: %v\n", err)
			return
		}

		if fileInfo.Size() > 500*1024 {
			fmt.Printf("Файл слишком большой (%d bytes). Максимальный размер: 500КB\n", fileInfo.Size())
			return
		}

		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Ошибка чтения файла: %v\n", err)
			return
		}

		data = fileData

		metadata[protocol.MetaOriginalFileName] = filepath.Base(filePath)
		metadata[protocol.MetaFileSize] = fmt.Sprintf("%d", len(fileData))
		metadata[protocol.MetaFileExtension] = filepath.Ext(filePath)

	case protocol.DataTypeBankCard:
		fmt.Print("Введите номер карты: ")
		cardNumber, _ := c.reader.ReadString('\n')
		cardNumber = strings.TrimSpace(cardNumber)

		fmt.Print("Введите срок действия (MM/YY): ")
		expiry, _ := c.reader.ReadString('\n')
		expiry = strings.TrimSpace(expiry)

		fmt.Print("Введите CVV: ")
		cvv, _ := c.reader.ReadString('\n')
		cvv = strings.TrimSpace(cvv)

		fmt.Print("Введите имя владельца: ")
		holder, _ := c.reader.ReadString('\n')
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

	fmt.Print("Хотите добавить дополнительное поле? Y/n: ")
	addMore, _ := c.reader.ReadString('\n')
	addMore = strings.TrimSpace(strings.ToLower(addMore))

	for addMore == "y" || addMore == "yes" {
		fmt.Print("Введите имя поля: ")
		fieldName, _ := c.reader.ReadString('\n')
		fieldName = strings.TrimSpace(fieldName)

		if fieldName == "" {
			fmt.Println("Имя поля не может быть пустым")
			continue
		}

		fmt.Print("Введите значение поля: ")
		fieldValue, _ := c.reader.ReadString('\n')
		fieldValue = strings.TrimSpace(fieldValue)

		metadata[fieldName] = fieldValue
		fmt.Printf("Добавлено поле: %s = %s\n", fieldName, fieldValue)

		fmt.Print("Хотите добавить еще одно поле? Y/n: ")
		addMore, _ = c.reader.ReadString('\n')
		addMore = strings.TrimSpace(strings.ToLower(addMore))
	}

	encryptedData, err := c.encryptData(data)
	if err != nil {
		log.Printf("Ошибка шифрования данных: %v\n", err)
		return
	}

	item := protocol.NewDataItem{
		Type:     dataType,
		Name:     name,
		Data:     encryptedData,
		Metadata: metadata,
	}

	log.Println("Сохраняем данные на сервере...")
	if err := c.SaveData(item); err != nil {
		log.Printf("Ошибка сохранения: %v\n", err)
		return
	}

	log.Println("Данные успешно сохранены!")
}

// encryptData шифрует данные.
//
// Parameters:
//
//	data - данные для шифрования
//
// Returns:
//
//	[]byte - зашифрованные данные
//	error  - ошибка шифрования
func (c *UIClient) encryptData(data []byte) ([]byte, error) {
	key := c.deriveSimpleKey()
	return crypto.Encrypt(data, key)
}

// decryptData дешифрует данные.
//
// Parameters:
//
//	data - зашифрованные данные
//
// Returns:
//
//	[]byte - расшифрованные данные
//	error  - ошибка дешифрования
func (c *UIClient) decryptData(data []byte) ([]byte, error) {
	key := c.deriveSimpleKey()
	return crypto.Decrypt(data, key)
}

// decryptItemData дешифрует данные элемента
//
// Parameters:
//
//	item - элемент с зашифрованными данными
//
// Returns:
//
//	[]byte - расшифрованные данные
//	error  - ошибка дешифрования
func (c *UIClient) decryptItemData(item protocol.DataItem) ([]byte, error) {
	key := c.deriveSimpleKey()
	return crypto.Decrypt(item.Data, key)
}

// deriveSimpleKey создает cryptographic key из пароля
//
// Returns:
//
//	[]byte - ключ длиной 32 байта
func (c *UIClient) deriveSimpleKey() []byte {
	hash := sha256.Sum256([]byte(c.GetUsername()))
	return hash[:]
}

// getDataTypeName возвращает человеко-читаемое имя типа данных.
//
// Parameters:
//
//	dataType - числовой тип данных из протокола
//
// Returns:
//
//	string - читаемое имя типа данных
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
