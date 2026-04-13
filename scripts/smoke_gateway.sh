#!/usr/bin/env bash
# Smoke-тест HTTP API (grpc-gateway) User + Train proxy.
# Запуск из каталога UserService:
#   ./scripts/smoke_gateway.sh
# Другой хост:
#   BASE_URL=http://localhost:8080 ./scripts/smoke_gateway.sh
# Уже есть пользователь — только логин:
#   SKIP_SIGNUP=1 TEST_EMAIL='you@example.com' TEST_PASSWORD='secret123' ./scripts/smoke_gateway.sh
# Лог всех HTTP-ответов (append):
#   LOG_FILE=/path/to/run.log ./scripts/smoke_gateway.sh
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="${LOG_FILE:-$ROOT_DIR/smoke_gateway.log}"

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
PASSWORD="${TEST_PASSWORD:-secret123}"
# SKIP_SIGNUP=1 — не создавать пользователя, только логин (нужны TEST_EMAIL и TEST_PASSWORD).
# Иначе: новый email + случайный российский номер (11 цифр, с 7), пока не удастся или 5 попыток.
SKIP_SIGNUP="${SKIP_SIGNUP:-0}"

REDC='\033[0;31m'
GREENC='\033[0;32m'
YELLOWC='\033[1;33m'
NC='\033[0m'

fail() {
  echo -e "${REDC}FAIL${NC} $*" >&2
  if [[ -n "${LOG_FILE:-}" ]]; then
    echo "======== $(ts_log) SCRIPT FAIL: $* ========" >> "$LOG_FILE"
  fi
  exit 1
}

ok() {
  echo -e "${GREENC}OK${NC}  $*"
}

info() {
  echo -e "${YELLOWC}==>${NC} $*"
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "Нужна команда: $1"
}

json_get() {
  local json="$1"
  local key="$2"
  python3 -c "import sys,json; print(json.load(sys.stdin).get('$key',''))" <<<"$json"
}

# 11 цифр, начинается с 7 (как ожидает валидация UserService)
random_phone() {
  python3 -c "import random; print('7' + ''.join(str(random.randint(0,9)) for _ in range(10)))"
}

ts_log() {
  date +"%Y-%m-%dT%H:%M:%S%z"
}

# Полный ответ API в лог (в том числе JWT — не коммитьте .log)
log_http_response() {
  local method="$1"
  local path="$2"
  local http_code="$3"
  local resp_body="$4"
  local req_note="${5:-}"
  {
    echo "======== $(ts_log) $method ${BASE_URL}${path} ========"
    [[ -n "$req_note" ]] && echo "request: $req_note"
    echo "HTTP $http_code"
    echo "$resp_body"
    echo
  } >> "$LOG_FILE"
}

init_log_file() {
  mkdir -p "$(dirname "$LOG_FILE")"
  {
    echo
    echo ">>>>>>>> Session start $(ts_log) PID=$$ BASE_URL=$BASE_URL LOG_FILE=$LOG_FILE"
    echo "NOTE: ответы могут содержать токены и персональные данные."
    echo
  } >> "$LOG_FILE"
}

curl_json() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local auth="${4:-}"
  local url="${BASE_URL}${path}"
  local req_note=""
  if [[ -n "$data" ]]; then
    req_note="body=${data}"
  fi
  if [[ -n "$auth" ]]; then
    req_note="${req_note}${req_note:+; }Authorization=Bearer ***"
  fi
  local tmp
  tmp="$(mktemp)"
  local code
  if [[ -n "$auth" ]]; then
    if [[ -n "$data" ]]; then
      code=$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" "$url" \
        -H 'Content-Type: application/json' \
        -H "Authorization: Bearer $auth" \
        -d "$data")
    else
      code=$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" "$url" \
        -H "Authorization: Bearer $auth")
    fi
  else
    if [[ -n "$data" ]]; then
      code=$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" "$url" \
        -H 'Content-Type: application/json' \
        -d "$data")
    else
      code=$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" "$url")
    fi
  fi
  body="$(cat "$tmp")"
  rm -f "$tmp"
  log_http_response "$method" "$path" "$code" "$body" "$req_note"
  echo "$code|$body"
}

init_log_file
info "BASE_URL=$BASE_URL (ответы пишутся в $LOG_FILE)"
need_cmd curl
need_cmd python3

# --- Train health (без JWT)
info "GET /api/v1/train/health"
raw=$(curl_json GET /api/v1/train/health)
code="${raw%%|*}"
body="${raw#*|}"
[[ "$code" == "200" ]] || fail "train health HTTP $code: $body"
[[ "$body" == *'"ok":true'* ]] || fail "train health body: $body"
ok "train health"

uid=""
EMAIL=""
TOKEN=""

if [[ "$SKIP_SIGNUP" == "1" ]]; then
  [[ -n "${TEST_EMAIL:-}" ]] || fail "SKIP_SIGNUP=1: задайте TEST_EMAIL"
  EMAIL="$TEST_EMAIL"
  info "SKIP_SIGNUP=1: только логин email=$EMAIL"
else
  max_attempts=5
  for ((attempt = 1; attempt <= max_attempts; attempt++)); do
    if [[ -n "${TEST_EMAIL:-}" ]]; then
      EMAIL="$TEST_EMAIL"
    else
      EMAIL="smoke-$(date +%s)-${attempt}-$RANDOM@example.test"
    fi
    if [[ -n "${TEST_PHONE:-}" ]]; then
      PHONE="$TEST_PHONE"
    else
      PHONE="$(random_phone)"
    fi

    info "POST /api/v1/users (SignUp attempt $attempt/$max_attempts email=$EMAIL phone=$PHONE)"
    payload=$(printf '{"email":"%s","password":"%s","phone":"%s"}' "$EMAIL" "$PASSWORD" "$PHONE")
    raw=$(curl_json POST /api/v1/users "$payload")
    code="${raw%%|*}"
    body="${raw#*|}"

    if [[ "$code" == "200" || "$code" == "201" ]]; then
      uid="$(json_get "$body" userId)"
      [[ -n "$uid" ]] || fail "SignUp: нет userId в ответе: $body"
      ok "SignUp userId=$uid"
      break
    fi

    if [[ "$code" == "409" && "$attempt" -lt "$max_attempts" ]]; then
      if [[ -n "${TEST_EMAIL:-}" && -n "${TEST_PHONE:-}" ]]; then
        fail "SignUp HTTP 409 (email и телефон заданы вручную — конфликт в БД). Используйте SKIP_SIGNUP=1 или смените TEST_EMAIL/TEST_PHONE: $body"
      fi
      echo -e "${YELLOWC}==>${NC} SignUp 409, пробуем другой email/телефон..."
      continue
    fi

    fail "SignUp HTTP $code: $body"
  done
  [[ -n "$uid" ]] || fail "SignUp: не удалось после $max_attempts попыток"
fi

# --- Login
info "POST /api/v1/auth/login"
raw=$(curl_json POST /api/v1/auth/login "$(printf '{"email":"%s","password":"%s"}' "$EMAIL" "$PASSWORD")")
code="${raw%%|*}"
body="${raw#*|}"
[[ "$code" == "200" ]] || fail "login HTTP $code: $body"
TOKEN="$(json_get "$body" accessToken)"
[[ ${#TOKEN} -gt 50 ]] || fail "login: короткий/пустой accessToken (положи полный токен из JSON): $body"
ok "login, accessToken length=${#TOKEN}"

# uid из логина, если пропускали SignUp
if [[ -z "$uid" ]]; then
  uid="$(json_get "$body" userId)"
  [[ -n "$uid" ]] || fail "login: нет userId в ответе: $body"
  ok "userId из login: $uid"
fi

# --- /users/me/id
info "GET /api/v1/users/me/id"
raw=$(curl_json GET /api/v1/users/me/id "" "$TOKEN")
code="${raw%%|*}"
body="${raw#*|}"
[[ "$code" == "200" ]] || fail "me/id HTTP $code: $body"
me="$(json_get "$body" userId)"
[[ "$me" == "$uid" ]] || fail "me/id userId mismatch: got=$me expected=$uid"
ok "users/me/id"

# --- GET user by id
info "GET /api/v1/users/{user_id}"
raw=$(curl_json "GET" "/api/v1/users/${uid}" "" "$TOKEN")
code="${raw%%|*}"
body="${raw#*|}"
[[ "$code" == "200" ]] || fail "get user HTTP $code: $body"
ok "get user"

# --- Train plans (proxy)
info "GET /api/v1/train/plans"
raw=$(curl_json GET /api/v1/train/plans "" "$TOKEN")
code="${raw%%|*}"
body="${raw#*|}"
[[ "$code" == "200" ]] || fail "train plans HTTP $code: $body"
[[ "$body" == *"plans"* ]] || fail "train plans body: $body"
ok "train/plans"

info "Все проверки прошли."
{
  echo "======== $(ts_log) SCRIPT DONE SUCCESS ========"
  echo
} >> "$LOG_FILE"

echo "Для ручных curl: EMAIL=$EMAIL PASSWORD=$PASSWORD"
echo "Повторный прогон без регистрации: SKIP_SIGNUP=1 TEST_EMAIL=$EMAIL TEST_PASSWORD=$PASSWORD ./scripts/smoke_gateway.sh"
echo "Лог ответов: $LOG_FILE"
