#!/usr/bin/env zsh

check() {
  local name="$1"
  local label="$2"
  if eval "$name"; then 
    echo "✅ $label"
  else
    echo "❌ $name"
    echo ">>>"
    echo "$output"
    echo "<<<"
  fi
}

has_go() { go version | grep -q 'go1.23' }
install_go() {
  echo "Please download and install Go from here:"
  echo "https://go.dev/dl/"
}

has_tailwind() {
  [[ -f ./tailwindcss ]]
}

install_tailwind() {
  local arch=$(uname -m)
  local os=$(uname -s)
  local executable=''
  case "$os-$arch" in
  Linux-x86_64) executable='tailwindcss-linux-x64' ;;
  Darwin-x86_64) executable='tailwindcss-macos-x64' ;;
  Darwin-arm64) executable='tailwindcss-macos-arm64' ;;
  *)
    echo "Unsupported platform: $os-$arch"
    return 1
    ;;
  esac
  curl --fail -L -o tailwindcss "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/$executable"
  chmod +x tailwindcss
}

ensure() {
  local name="$1"
  local label="$2"
  if ! has_$name; then
    install_$name
  fi
  check "has_$name" "$label"
}

main() {
  local -A modules=(
    [go]="Go v1.23.0"
    [tailwind]=tailwindcss
  )
  case "$1" in
    check)
      for module label in "${(@kv)modules[@]}"; do
        check "has_$module" "$label"
      done
      ;;
    *)
      for module label in "${(@kv)modules[@]}"; do
        ensure $module "$label"
      done
      ;;
  esac
}

main "$@"
