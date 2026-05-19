# wayt-core

Shared library untuk platform Wayt. Di-publish ke GitHub dan diimpor oleh wayt-admin, wayt-owner, wayt-customer.

**Module**: `github.com/wayt-app/wayt-core`

## Struktur

```
config/         — config.Load() dari env vars
model/          — domain structs (GORM models)
repository/     — database queries (interface + postgres impl)
service/        — business logic
migrations/     — SQL migration files
pkg/
  email/        — Resend email sender
  middleware/   — Gin middleware (JWT auth, rate limit, dll)
  response/     — standar HTTP response helper
  sse/          — Server-Sent Events hub
  storage/      — Supabase storage
  whatsapp/     — Fonnte WhatsApp sender
```

## Domain Models

`AdminUser`, `Booking`, `Branch`, `BusinessOwner`, `Customer`, `Media`, `Notification`, `Plan`, `PlanOrder`, `Restaurant`, `Staff`, `Subscription`, `TableType`, `Slot`

## Catatan

- Ini library, bukan aplikasi — tidak ada `cmd/main.go`
- Setiap breaking change → bump versi di go.mod tiap app yang menggunakannya
- Versi saat ini dipakai: v0.3.0 (admin), v0.4.0 (owner), v0.5.0 (customer)
