# Bank Transfer API

Go va PostgreSQL yordamida yozilgan oddiy bank hisoblari va hisoblar orasida pul o'tkazish REST API loyihasi.

## Imkoniyatlar

- Barcha bank hisoblarini ko'rish.
- ID orqali bitta hisobni ko'rish.
- Bir hisobdan boshqa hisobga pul o'tkazish.
- Barcha pul o'tkazmalar tarixini ko'rish.
- ID orqali bitta pul o'tkazmani ko'rish.
- Pul o'tkazishni PostgreSQL transaction ichida bajarish.
- Parallel pul o'tkazishlarda hisob qatorlarini `FOR UPDATE` bilan bloklash.
- Kiritilgan ma'lumotlarni tekshirish va mos HTTP status kodlarini qaytarish.
- Unit testlar (`httptest` yordamida, PostgreSQL talab qilmaydi).

## Texnologiyalar

- Go 1.23+
- `net/http`
- PostgreSQL
- `github.com/jackc/pgx/v5`

## Loyihani ishga tushirish

### 1. Talablar

Kompyuterda Go 1.23 yoki undan yangi versiya va PostgreSQL o'rnatilgan bo'lishi kerak.

### 2. Database yaratish

PostgreSQL terminalida:

```sql
CREATE DATABASE bank_transfer;
```

Keyin `bank_transfer` bazasiga ulanib, loyihadagi `sql.sql` faylini bajaring. Masalan:

```bash
psql -U postgres -d bank_transfer -f sql.sql
```

Agar PostgreSQL foydalanuvchingiz `postgres` bo'lmasa, buyruqdagi foydalanuvchi nomini o'zgartiring.

### 3. Environment sozlash

`.env.example` faylidan nusxa olib `.env` yarating:

```bash
cp .env.example .env
```

`.env` ichidagi `DATABASE_URL` qiymatini o'zingizning PostgreSQL foydalanuvchi nomingiz, parolingiz va portingizga moslang.

`.env` Git'ga qo'shilmaydi. Haqiqiy parolni GitHub'ga yuklamang.

### 4. Dependency va ishga tushirish

```bash
go mod tidy
go run .
```

Server standart holatda `http://localhost:8080` manzilida ishlaydi. Portni `.env` ichidagi `SERVER_ADDR` orqali o'zgartirish mumkin.

Health check:

```http
GET http://localhost:8080/health
```

## API endpointlar

| Method | Endpoint | Tavsif |
|---|---|---|
| GET | `/health` | Server holati |
| GET | `/accounts` | Barcha hisoblar |
| GET | `/accounts/{id}` | Bitta hisob |
| POST | `/transfer` | Pul o'tkazish |
| GET | `/transfers` | Barcha o'tkazmalar |
| GET | `/transfers/{id}` | Bitta o'tkazma |

### Pul o'tkazish namunasi

`POST /transfer`

Header: `Content-Type: application/json`

```json
{
  "sender_id": 1,
  "receiver_id": 2,
  "amount": 15000.50
}
```

Muvaffaqiyatli javob (`200 OK`):

```json
{
  "message": "Pul muvaffaqiyatli o'tkazildi"
}
```

## HTTP status kodlari

- `200 OK` — so'rov muvaffaqiyatli.
- `400 Bad Request` — noto'g'ri ID, summa, JSON yoki balans yetarli emas.
- `404 Not Found` — hisob yoki o'tkazma topilmadi.
- `405 Method Not Allowed` — endpoint uchun noto'g'ri HTTP method.
- `500 Internal Server Error` — kutilmagan server yoki database xatosi.

Xatolar JSON ko'rinishida qaytadi:

```json
{
  "error": "Balans yetarli emas"
}
```

## Testlar

Unit testlarni ishga tushirish:

```bash
go test ./...
```

Race detector bilan:

```bash
go test -race ./...
```

Handler testlari `httptest` va soxta repositorylardan foydalanadi, shu sababli unit testlar uchun PostgreSQL ishga tushirilishi shart emas.

## Eslatma

Bu loyiha o'quv maqsadida yaratilgan. Haqiqiy moliyaviy tizimlarda pul qiymatlari uchun `float64` o'rniga aniq decimal turidan foydalanish, autentifikatsiya/avtorizatsiya, audit, idempotency, rate limiting, monitoring va qo'shimcha xavfsizlik tekshiruvlarini joriy etish kerak.
# bank-transfer
