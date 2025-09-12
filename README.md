# Password Manager

Secure password management system with client-server architecture.

### Quick Start

#### 1. Install and Setup PostgreSQL

**On Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql -c "CREATE DATABASE password_manager;"
sudo -u postgres psql -c "CREATE USER pm_user WITH PASSWORD 'password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE password_manager TO pm_user;"
```

**On MacOS:**
```bash
brew install postgresql
brew services start postgresql

# Create database and user
createdb password_manager
createuser pm_user
psql -c "ALTER USER pm_user WITH PASSWORD 'password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE password_manager TO pm_user;"
```

**2. Build the Application**
```bash
git clone git clone --single-branch --branch ver1 https://github.com/SemAnton-2007/go-password-manager.git
cd go-password-manager

# Build both server and client
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```
**3. Run the Server**
```bash
# Start the server with PostgreSQL connection
./bin/server -db-host=localhost -db-user=pm_user -db-password=password -db-name=password_manager
```
**4. Run the Client**
```bash
run interactively
./bin/client
```
**Usage Example**
Step-by-Step Navigation
1. Start the Client and Connect
```bash
$ ./bin/client
2025/09/12 19:37:38 === Password Manager Client ===
Введите адрес сервера [localhost]: 
Введите порт сервера [8080]: 
2025/09/12 19:37:39 Попытка подключения к localhost:8080...
2025/09/12 19:37:39 Подключение успешно!
```
**2. User Registration**
```bash
Выберите тип пользователя:
1. Новый пользователь
2. Зарегистрированный пользователь
Ваш выбор [1]: 1

Введите логин: test_user
Введите пароль: my_secure_password
2025/09/12 19:37:44 Регистрируем пользователя...
2025/09/12 19:37:44 Регистрация успешна!
2025/09/12 19:37:44 Авторизуем пользователя...
2025/09/12 19:37:44 Авторизация успешна!
```
**3. Main Menu Navigation**
```bash
=== Главное меню (пользователь: test_user) ===
1. Показать мои данные
2. Создать новый элемент
3. Выйти
Ваш выбор [3]: 2
```
**4. Create a Text Item**
```bash
=== Главное меню (пользователь: test) ===
1. Показать мои данные
2. Создать новый элемент
3. Выйти
Ваш выбор [3]: 2

=== Создание нового элемента ===
Выберите тип данных:
1. Логин/Пароль
2. Текстовые данные
3. Бинарные данные (файл)
4. Банковская карта
Ваш выбор [1]: 2
Введите название элемента: text
Введите текст: 12321
Хотите добавить дополнительное поле? Y/n: y
Введите имя поля: otp
Введите значение поля: 111
Добавлено поле: otp = 111
Хотите добавить еще одно поле? Y/n: n
2025/09/12 19:38:10 Сохраняем данные на сервере...
2025/09/12 19:38:10 Данные успешно сохранены!
```
**5. View Your Data**
```bash
=== Главное меню (пользователь: test) ===
1. Показать мои данные
2. Создать новый элемент
3. Выйти
Ваш выбор [3]: 1

=== Мои данные ===
2025/09/12 19:38:12 Синхронизируем данные...

Найдено 1 элементов:
1. text (Текст)

Действия:
0. Вернуться назад
1-9. Показать детали элемента
Ваш выбор [0]: 1

=== Детали элемента: text ===
Тип: Текст
Создан: 2025-09-12 19:38:10
Обновлен: 2025-09-12 19:38:10

--- Метаинформация ---
otp: 111

--- Текстовые данные ---
12321

Действия:
0. Вернуться назад
1. Удалить элемент
2. Редактировать элемент
Ваш выбор [0]: 0
```
