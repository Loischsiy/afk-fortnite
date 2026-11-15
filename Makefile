.PHONY: build build-gui icon tidy clean build-all build-linux build-windows build-darwin build-darwin-arm64 build-darwin-amd64 install-deps package appimage

# Переменные
BINARY_NAME=afk-fortnite
VERSION=$(shell git describe --tags --always --dirty)
BUILD_DIR=dist
ICON_SYSO=rsrc_windows_amd64.syso

# AppImage: linuxdeploy сам подтянет разделяемые библиотеки (X11, GL и т.д.)
LINUXDEPLOY_URL=https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage
LINUXDEPLOY=$(BUILD_DIR)/linuxdeploy-x86_64.AppImage

# Основная цель - сборка для текущей платформы
build:
	go build -v -o $(BINARY_NAME) .

# Генерация ресурса с иконкой для Windows .exe (файл rsrc_windows_amd64.syso
# автоматически вкомпилируется в бинарник только для GOOS=windows GOARCH=amd64)
$(ICON_SYSO): icon.ico
	go install github.com/akavel/rsrc@latest
	rsrc -ico icon.ico -arch amd64 -o $(ICON_SYSO)

icon: $(ICON_SYSO)

# Сборка GUI-версии без окна консоли + иконка (актуально для Windows)
build-gui: $(ICON_SYSO)
	go build -ldflags="-H=windowsgui" -v -o $(BINARY_NAME).exe .

# Обновление зависимостей (нужно после добавления Fyne)
tidy:
	go mod tidy

# Очистка
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(ICON_SYSO)

# Создание директории для сборки
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Сборка для всех платформ
build-all: build-linux build-windows build-darwin-arm64 build-darwin-amd64

# Сборка для Linux
build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

# Сборка для Windows (требует mingw-w64), GUI без консоли + иконка
build-windows: $(BUILD_DIR) $(ICON_SYSO)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags="-s -w -H=windowsgui" -v -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Сборка для macOS ARM64 (Apple Silicon) - Fyne и robotgo требуют CGO
build-darwin-arm64: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .

# Сборка для macOS x86_64 (Intel) - Fyne и robotgo требуют CGO
build-darwin-amd64: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -v -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .

# Обратная совместимость
build-darwin: build-darwin-arm64 build-darwin-amd64

# Установка зависимостей для сборки (Ubuntu/Debian)
install-deps:
	sudo apt-get update
	sudo apt-get install -y libx11-dev xorg-dev libxtst-dev libpng++-dev
	sudo apt-get install -y xcb libxcb-xkb-dev x11-xkb-utils libx11-xcb-dev
	sudo apt-get install -y libxkbcommon-x11-dev libxkbcommon-dev
	sudo apt-get install -y libgl1-mesa-dev
	sudo apt-get install -y gcc-mingw-w64

# Создание архивов
package: build-all
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64

# Сборка AppImage (только на Linux; под macOS не работает)
$(LINUXDEPLOY):
	curl -fL -o $@ $(LINUXDEPLOY_URL)
	chmod +x $@

$(BUILD_DIR)/AppDir/afk-fortnite.desktop: $(BUILD_DIR)
	mkdir -p $(BUILD_DIR)/AppDir/usr/bin
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(BUILD_DIR)/AppDir/usr/bin/$(BINARY_NAME)
	cp packaging/linux/$(BINARY_NAME).desktop $(BUILD_DIR)/AppDir/
	cp packaging/linux/$(BINARY_NAME).png $(BUILD_DIR)/AppDir/
	ln -sf $(BINARY_NAME).png $(BUILD_DIR)/AppDir/.DirIcon
	touch $@

# Итоговый файл: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.AppImage
# APPIMAGE_EXTRACT_AND_RUN=1 - запуск linuxdeploy без FUSE (нужно в CI)
appimage: build-linux $(BUILD_DIR)/AppDir/afk-fortnite.desktop $(LINUXDEPLOY)
	cd $(BUILD_DIR) && APPIMAGE_EXTRACT_AND_RUN=1 ARCH=x86_64 ./linuxdeploy-x86_64.AppImage \
		--appdir AppDir -d $(BINARY_NAME).desktop -i $(BINARY_NAME).png --output appimage
	mv $(BUILD_DIR)/AFK_Fortnite-x86_64.AppImage $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64.AppImage
