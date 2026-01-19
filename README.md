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

## Langkah Deployment ke GCP (versi laporan siswa)

1. Siapkan VM di Google Cloud Platform, lalu buka port aplikasi di firewall (misalnya 8080).
2. Install Go, Python, dan pip di VM, lalu download source code project.
3. Masuk ke `backend-python` dan install dependency:
   ```bash
   python -m pip install -r requirements.txt
   ```
4. Pastikan model YOLO ada di `backend-python/model/best.pt`.
5. Set environment variable:
   - `YOLO_PY_SCRIPT` (path ke `backend-python/detect.py`)
   - `PYTHON_BIN` (path ke python)
   - opsional: `ADDR`, `YOLO_TIMEOUT_SECONDS`, dll.
6. Jalankan backend Go dari `backend-go`:
   ```bash
   go run ./cmd/server
   ```
7. Pasang Nginx sebagai reverse proxy supaya project bisa diakses lewat domain.
8. Saat presentasi, tampilkan halaman docs di `http://<IP-VM>/` (via Nginx) dan uji endpoint `/detect`.

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
