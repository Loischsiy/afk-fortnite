#!/usr/bin/env bash
# Готовит vendor/ для офлайн-сборки (используется NixOS-сборкой, vendorHash=null).
#
# Проблема: go mod vendor переносит только директории, import-аемые Go-кодом.
# C-заголовки robotgo и gohook лежат в директориях без .go-файлов
# (base/, event/, screen/ и т.д.), поэтому в vendor/ они не попадают и
# CGO-сборка падает. Скрипт докладывает эти директории из модульного кэша
# (те же версии, что в go.mod).
#
# Запуск при каждом изменении зависимостей: go mod tidy && bash scripts/vendor.sh
set -euo pipefail
cd "$(dirname "$0")/.."

go mod tidy
go mod vendor

cache=$(go env GOMODCACHE)

# module_path  список директорий с C-заголовками
while read -r mod dirs; do
  [ -z "$mod" ] && continue
  ver=$(go list -m -f '{{.Version}}' "$mod")
  for sub in $dirs; do
    if [ -d "$cache/$mod@$ver/$sub" ]; then
      rm -rf "vendor/$mod/$sub"
      cp -R "$cache/$mod@$ver/$sub" "vendor/$mod/$sub"
      echo "vendor: добавлен $mod/$sub"
    else
      echo "vendor: в кэше нет $mod@$ver/$sub (проверьте версию)" >&2
      exit 1
    fi
  done
done <<'EOF'
github.com/go-vgo/robotgo base cv event key mouse screen window
github.com/robotn/gohook chan event hook
EOF

echo "vendor/ готов"
