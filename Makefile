.PHONY: build clean build-all build-linux build-windows build-darwin

# Переменные
BINARY_NAME=afk-fortnite
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=dist

# Основная цель - сборка для текущей платформы
build:
	go build -v -o $(BINARY_NAME) .

# Очистка
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

# Создание директории для сборки
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Сборка для всех платформ
build-all: build-linux build-windows build-darwin-arm64 build-darwin-amd64

# Сборка для Linux
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

# Сборка для Windows (требует mingw-w64)
build-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w -H=windowsgui" -v -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Сборка для macOS ARM64 (Apple Silicon)
build-darwin-arm64: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Сборка для macOS x86_64 (Intel)
build-darwin-amd64: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

# Обратная совместимость
build-darwin: build-darwin-arm64 build-darwin-amd64

# Установка зависимостей для кросс-компиляции (Ubuntu/Debian)
install-deps:
	sudo apt-get update
	sudo apt-get install -y libx11-dev xorg-dev libxtst-dev libpng++-dev
	sudo apt-get install -y xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev
	sudo apt-get install -y libxkbcommon-x11-dev libxkbcommon-dev
	sudo apt-get install -y gcc-mingw-w64

# Создание архивов
package: build-all
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64