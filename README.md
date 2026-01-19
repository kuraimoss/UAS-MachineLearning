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

## Langkah Deployment ke GCP

1. Buat VM di Google Cloud Platform (Ubuntu), lalu buka firewall untuk port 22 (SSH), 80 (HTTP), dan 8080 (app).
2. SSH ke VM, lalu install kebutuhan:
   ```bash
   sudo apt update
   sudo apt install -y git python3 python3-pip nginx
   sudo snap install go --classic
   ```
3. Clone project ke VM:
   ```bash
   git clone <repo-url>
   cd UAS
   ```
4. Install dependency Python:
   ```bash
   cd backend-python
   python3 -m pip install -r requirements.txt
   ```
5. Pastikan model YOLO tersedia di `backend-python/model/best.pt`.
6. Set environment variable agar Go bisa memanggil Python:
   ```bash
   export YOLO_PY_SCRIPT="/home/<user>/UAS/backend-python/detect.py"
   export PYTHON_BIN="/usr/bin/python3"
   export ADDR=":8080"
   ```
7. Build dan jalankan backend Go:
   ```bash
   cd /home/<user>/UAS/backend-go
   go build -o server ./cmd/server
   ./server
   ```
8. (Opsional) Jalankan sebagai service agar auto-start:
   ```bash
   sudo tee /etc/systemd/system/plate.service > /dev/null <<'EOF'
   [Unit]
   Description=Plate Detection API
   After=network.target

   [Service]
   WorkingDirectory=/home/<user>/UAS/backend-go
   Environment=YOLO_PY_SCRIPT=/home/<user>/UAS/backend-python/detect.py
   Environment=PYTHON_BIN=/usr/bin/python3
   Environment=ADDR=:8080
   ExecStart=/home/<user>/UAS/backend-go/server
   Restart=always
   RestartSec=3

   [Install]
   WantedBy=multi-user.target
   EOF
   sudo systemctl daemon-reload
   sudo systemctl enable --now plate.service
   ```
9. Konfigurasi Nginx sebagai reverse proxy (akses lewat domain atau IP publik):
   ```bash
   sudo tee /etc/nginx/sites-available/plate.conf > /dev/null <<'EOF'
   server {
       listen 80;
       server_name <domain-atau-ip>;

       location / {
           proxy_pass http://127.0.0.1:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }
   }
   EOF
   sudo ln -s /etc/nginx/sites-available/plate.conf /etc/nginx/sites-enabled/plate.conf
   sudo nginx -t
   sudo systemctl restart nginx
   ```
10. Saat presentasi, buka `http://<domain-atau-ip>/` untuk halaman docs dan lakukan uji endpoint `/detect` dengan upload gambar.

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
