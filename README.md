# UAS-MachineLearning — Plat Detection System

Backend API untuk deteksi & pembacaan plat nomor kendaraan (motor/mobil) dari foto menggunakan deep learning, lalu lookup wilayah Samsat berdasarkan plat.

## Fitur

- Upload gambar (`multipart/form-data`)
- Deteksi area plat (YOLOv8 custom)
- OCR teks plat (TrOCR)
- Normalisasi plat (contoh: `BK 4272 AMQ`)
- Lookup wilayah Samsat (source: samsat.info)
- UI dokumentasi API + Try It (served langsung dari Go)

## Struktur

- `backend-go/` REST API (Golang)
- `backend-python/` YOLO + OCR worker (Python, dieksekusi via `os/exec`)

## Prasyarat

- Windows / Linux / macOS
- Go (>= 1.22)
- Python (>= 3.10) + pip

## Setup Python

```bash
cd backend-python
python -m pip install -r requirements.txt
```

Pastikan model YOLO ada di `backend-python/model/best.pt`.

## Menjalankan Server (Go)

PowerShell (Windows):

```powershell
$env:YOLO_PY_SCRIPT="e:\RPL\Semester 7\ML\UAS\backend-python\detect.py"
$env:PYTHON_BIN="C:\Users\kurai\AppData\Local\Programs\Python\Python310\python.exe"
cd "e:\RPL\Semester 7\ML\UAS\backend-go"
go run ./cmd/server
```

Buka docs:
- `http://localhost:8080/` (Docs + Try It)

## Testing dengan Postman

Request:
- Method: `POST`
- URL: `http://localhost:8080/detect`
- Body: `form-data`
  - key: `image` (type: File) pilih gambar

## Environment Variables (opsional)

- `ADDR` (default `:8080`)
- `PYTHON_BIN` (default `python`)
- `YOLO_PY_SCRIPT` (wajib): path ke `backend-python/detect.py`
- `YOLO_TIMEOUT_SECONDS` (default `120`)
- `SCRAPE_TIMEOUT_SECONDS` (default `15`)
- `MAX_UPLOAD_MB` (default `15`)
- `MIN_PLATE_CONFIDENCE` (default `0`)
- `SAMSAT_FIRESTORE_API_KEY` (default: mengikuti value dari samsat.info)
- `SAMSAT_PAGE_URL` (default: halaman samsat)
- `SAMSAT_FIRESTORE_BASE_URL` (default: firestore REST base)

## Catatan

- Python tidak membuka port apa pun; hanya dieksekusi oleh Go.
- Untuk akses kamera di browser: `capture` lebih konsisten di mobile; desktop butuh WebRTC jika ingin kamera inline.
