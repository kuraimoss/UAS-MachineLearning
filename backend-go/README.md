# backend-go

Backend REST API (Go) untuk:

- menerima request deteksi plat
- memanggil inferensi YOLO (Python)
- (opsional) scraping data SAMSAT

## Menjalankan

1. Pastikan Go sudah ter-install.
2. Jalankan server:

```bash
cd backend-go
go run ./cmd/server
```

Server default di `http://localhost:8080`.

## Endpoint

- `GET /healthz`
- `POST /detect`
  - `multipart/form-data` (disarankan): upload file pakai key `image` (atau `file`)
  - Response: envelope JSON (`status`, `request_id`, `processing_time_ms`, `data`, `meta`)

## Integrasi Python (opsional)

`/detect` akan mencoba memanggil script Python kalau env var `YOLO_PY_SCRIPT` di-set ke path script inferensi (mis. `..\backend-python\detect.py`).

## Environment Variables

- `ADDR` (default `:8080`)
- `YOLO_PY_SCRIPT` (wajib): path ke `backend-python/detect.py`
- `PYTHON_BIN` (default `python`): path binary python
- `YOLO_TIMEOUT_SECONDS` (default `120`)
- `SCRAPE_TIMEOUT_SECONDS` (default `15`)
- `MAX_UPLOAD_MB` (default `15`)
- `MIN_PLATE_CONFIDENCE` (default `0`)
- `SAMSAT_FIRESTORE_API_KEY` (default sesuai samsat.info)
- `SAMSAT_PAGE_URL` (default halaman samsat)
- `SAMSAT_FIRESTORE_BASE_URL` (default firestore REST base)

## Contoh Postman

Request:
- Method: `POST`
- URL: `http://localhost:8080/detect`
- Body: `form-data`
  - key: `image` (type: File) pilih gambar

Contoh error response:

```json
{
  "status": "error",
  "request_id": "uuid",
  "processing_time_ms": 123,
  "error": { "code": "MISSING_FILE", "message": "field file tidak ditemukan (key: image)" }
}
```
