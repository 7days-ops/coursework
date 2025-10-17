# Integrity Monitor

Система мониторинга целостности системных утилит для Linux. Программа отслеживает изменения в исполняемых файлах в критических системных директориях и оповещает всех активных пользователей о потенциальной угрозе безопасности.

## Возможности

- ✅ Вычисление и хранение контрольных сумм (SHA256) системных утилит
- ✅ Мониторинг в реальном времени с использованием inotify (fsnotify)
- ✅ Периодическое сканирование всех утилит
- ✅ Отправка предупреждений на все активные TTY/PTS терминалы
- ✅ Логирование всех событий
- ✅ SQLite база данных для хранения эталонных checksums
- ✅ Настраиваемые пути мониторинга

## Архитектура проекта

```
coursework/
├── cmd/
│   └── integrity-monitor/
│       └── main.go              # Главный файл приложения
├── internal/
│   ├── scanner/                 # Сканирование директорий
│   │   ├── scanner.go
│   │   └── paths.go
│   ├── checksum/                # Вычисление контрольных сумм
│   │   ├── calculator.go
│   │   └── comparator.go
│   ├── database/                # Работа с БД
│   │   ├── storage.go
│   │   └── sqlite.go
│   ├── notifier/                # Система уведомлений
│   │   ├── notifier.go
│   │   ├── tty.go
│   │   └── logger.go
│   ├── watcher/                 # Мониторинг файловой системы
│   │   ├── watcher.go
│   │   └── events.go
│   └── config/                  # Конфигурация
│       └── config.go
├── pkg/
│   └── models/                  # Модели данных
│       └── utility.go
├── configs/
│   └── config.yaml              # Файл конфигурации
└── scripts/
    ├── install.sh               # Скрипт установки
    └── integrity-monitor.service # Systemd сервис
```

## Установка

### Предварительные требования

- Go 1.21 или выше
- Linux система
- Root доступ для установки

### Шаги установки

1. **Клонировать репозиторий:**
```bash
cd ~/coursework
```

2. **Установить зависимости:**
```bash
go mod download
```

3. **Собрать проект:**
```bash
go build -o integrity-monitor cmd/integrity-monitor/main.go
```

4. **Установить в систему (требует root):**
```bash
cd scripts
chmod +x install.sh
sudo ./install.sh
```

## Инициализация базы данных

**ВАЖНО:** Перед первым запуском необходимо создать базу данных с эталонными контрольными суммами.

### Откуда берутся checksums?

Контрольные суммы **вычисляются из текущего состояния вашей системы**. Программа:
1. Сканирует все исполняемые файлы в `/bin`, `/sbin`, `/usr/bin`, `/usr/sbin`, и т.д.
2. Вычисляет SHA256 хэш каждого файла
3. Сохраняет эти хэши в базу данных как "эталонные"

**Поэтому инициализацию нужно проводить на чистой, доверенной системе!**

### Команда инициализации:

```bash
sudo integrity-monitor -init
```

Это создаст файл `/var/lib/integrity-monitor/checksums.db` со всеми контрольными суммами.

Пример вывода:
```
Initializing database with current system state...
Found 2847 utilities to process
Progress: 100/2847 utilities processed
Progress: 200/2847 utilities processed
...
Initialization complete! Stored checksums for 2847/2847 utilities
```

## Использование

### Режимы работы

#### 1. Инициализация базы данных
```bash
sudo integrity-monitor -init
```

#### 2. Одноразовое сканирование
```bash
sudo integrity-monitor -scan
```

#### 3. Непрерывный мониторинг (по умолчанию)
```bash
sudo integrity-monitor
```

или

```bash
sudo integrity-monitor -monitor
```

### Конфигурация

Отредактируйте `/etc/integrity-monitor/config.yaml`:

```yaml
database:
  path: /var/lib/integrity-monitor/checksums.db

monitored_paths:
  - /bin
  - /sbin
  - /usr/bin
  - /usr/sbin
  - /usr/local/bin
  - /usr/local/sbin

scan_interval: 300  # секунды (5 минут)
enable_watcher: true
log_file: /var/log/integrity-monitor.log
```

### Запуск как системный сервис

1. **Скопировать systemd unit файл:**
```bash
sudo cp scripts/integrity-monitor.service /etc/systemd/system/
```

2. **Перезагрузить systemd и запустить сервис:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable integrity-monitor
sudo systemctl start integrity-monitor
```

3. **Проверить статус:**
```bash
sudo systemctl status integrity-monitor
```

4. **Просмотр логов:**
```bash
sudo journalctl -u integrity-monitor -f
```

## Принцип работы

### 1. Инициализация
- Программа сканирует все указанные директории
- Для каждого исполняемого файла вычисляется SHA256 хэш
- Хэши сохраняются в SQLite базе данных

### 2. Мониторинг

Программа использует два механизма:

**A. Real-time мониторинг (fsnotify/inotify):**
- Отслеживает события изменения файлов в реальном времени
- Срабатывает мгновенно при модификации файла

**B. Периодическое сканирование:**
- Каждые N секунд (по умолчанию 300) сканирует все файлы
- Дополнительная защита на случай пропуска событий

### 3. Обнаружение изменений

При обнаружении изменения файла:
1. Вычисляется новый checksum
2. Сравнивается с эталонным из БД
3. Если не совпадает - создается alert

### 4. Оповещение

При обнаружении подмены утилиты:
- Запись в `/var/log/integrity-monitor.log`
- Запись в базу данных (таблица alerts)
- **Отправка сообщения на все активные TTY/PTS терминалы**

Пример сообщения на TTY:
```
╔══════════════════════════════════════════════════════════════╗
║              ⚠️  SECURITY ALERT - UTILITY MODIFIED  ⚠️        ║
╠══════════════════════════════════════════════════════════════╣
║ Path:          /usr/bin/ls
║ Severity:      CRITICAL
║ Old Checksum:  a1b2c3d4e5f6g7h8...
║ New Checksum:  x9y8z7w6v5u4t3s2...
║ Detected At:   2025-10-17 20:30:45
║
║ WARNING: A system utility has been modified!
║ This could indicate a security breach or malicious activity.
║ Please investigate immediately!
╚══════════════════════════════════════════════════════════════╝
```

## Тестирование

### Проверка работы системы:

1. **Инициализировать БД:**
```bash
sudo integrity-monitor -init
```

2. **Запустить мониторинг в фоне:**
```bash
sudo integrity-monitor &
```

3. **Изменить какую-то утилиту (для теста):**
```bash
# Создать копию оригинала
sudo cp /bin/ls /bin/ls.backup

# Изменить файл (добавить один байт)
sudo sh -c 'echo " " >> /bin/ls'
```

4. **Наблюдать alert на всех TTY**

5. **Восстановить оригинал:**
```bash
sudo mv /bin/ls.backup /bin/ls
```

## База данных

### Структура БД (SQLite)

**Таблица `utilities`:**
- `id` - PRIMARY KEY
- `path` - полный путь к файлу
- `checksum` - SHA256 хэш
- `last_modified` - время изменения файла
- `size` - размер файла
- `created_at` - время добавления в БД
- `updated_at` - время последнего обновления

**Таблица `alerts`:**
- `id` - PRIMARY KEY
- `utility_path` - путь к измененному файлу
- `old_checksum` - старый хэш
- `new_checksum` - новый хэш
- `detected_at` - время обнаружения
- `severity` - уровень критичности

### Просмотр данных в БД:

```bash
sqlite3 /var/lib/integrity-monitor/checksums.db

# Показать все утилиты
SELECT * FROM utilities LIMIT 10;

# Показать все alerts
SELECT * FROM alerts;

# Количество утилит
SELECT COUNT(*) FROM utilities;
```

## Безопасность

⚠️ **Важные замечания:**

1. **Инициализация на чистой системе:** Запускайте `-init` только на системе, которой вы полностью доверяете
2. **Root доступ:** Программа требует root для доступа к системным директориям
3. **Защита БД:** Файл `checksums.db` критически важен - защитите его от изменений
4. **Ложные срабатывания:** Обновления системы изменят checksums - потребуется реинициализация

## Возможные улучшения

- [ ] Белый список процессов, которые могут изменять утилиты (apt, yum, etc.)
- [ ] Интеграция с package managers для автоматического обновления checksums
- [ ] Email уведомления
- [ ] Web интерфейс для просмотра alerts
- [ ] Экспорт checksums в стандартные форматы (AIDE, Tripwire)
- [ ] Поддержка цифровых подписей

## Требования к системе

- Linux kernel 2.6.13+ (для inotify)
- SQLite3
- Минимум 50MB свободного места для БД
- Root привилегии

## Лицензия

Курсовая работа - MIT License

## Автор

Кирилл Путсеев
