# backend-python

Tempat untuk model YOLO dan script inferensi.

## Struktur

- `model/` taruh file model (mis. `best.pt`)
- `detect.py` entry point untuk dipanggil dari Go (set `YOLO_PY_SCRIPT`)

## Menjalankan (manual)

1. Install dependency:

```bash
cd backend-python
pip install -r requirements.txt
```

2. Jalankan deteksi:

```bash
python detect.py path\\ke\\gambar.jpg
```

Output ke stdout (JSON):

```json
{"plate_raw":"...","plate_cleaned":"...","confidence":0.9}
```

## OCR Model (TroCR)

Default OCR pakai HuggingFace TroCR (`microsoft/trocr-base-printed`). Bisa diubah lewat env var:

- `OCR_MODEL_NAME` (contoh: `microsoft/trocr-large-printed` atau path lokal model)
- `OCR_DEVICE` (contoh: `cuda` atau `cpu`)
