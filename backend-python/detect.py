import json
import os
import sys
from contextlib import redirect_stderr, redirect_stdout
from io import StringIO
from pathlib import Path

from utils.ocr_cleaner import clean_ocr_text


def _script_dir() -> Path:
    return Path(__file__).resolve().parent


def _model_path() -> Path:
    return _script_dir() / "model" / "best.pt"


def _print_json(obj: dict) -> None:
    sys.stdout.write(json.dumps(obj, ensure_ascii=False))


def _parse_image_path(argv: list[str]) -> str | None:
    if len(argv) >= 3 and argv[1] == "--image":
        return argv[2]
    if len(argv) >= 2:
        return argv[1]
    return None


YOLO_MODEL = None
PADDLE_OCR = None


def _load_yolo():
    global YOLO_MODEL
    if YOLO_MODEL is not None:
        return YOLO_MODEL

    buf_out = StringIO()
    buf_err = StringIO()
    with redirect_stdout(buf_out), redirect_stderr(buf_err):
        from ultralytics import YOLO  # type: ignore

        YOLO_MODEL = YOLO(str(_model_path()))
    return YOLO_MODEL


def _load_paddle_ocr():
    global PADDLE_OCR
    if PADDLE_OCR is not None:
        return PADDLE_OCR

    buf_out = StringIO()
    buf_err = StringIO()
    with redirect_stdout(buf_out), redirect_stderr(buf_err):
        # Force CPU-only Paddle runtime
        os.environ.setdefault("CUDA_VISIBLE_DEVICES", "-1")
        from paddleocr import PaddleOCR  # type: ignore

        PADDLE_OCR = PaddleOCR(lang="en", use_angle_cls=True, use_gpu=False, show_log=False)

    return PADDLE_OCR


def detect_plate(image_path: str) -> dict:
    try:
        # Import di dalam fungsi supaya kalau ada error, tetap bisa return JSON (bukan crash tanpa stdout)
        os.environ.setdefault("OPENCV_LOG_LEVEL", "SILENT")
        import cv2  # type: ignore
        import numpy as np  # type: ignore

        # Matikan log OpenCV (biar tidak ada output selain JSON)
        try:
            if hasattr(cv2, "utils") and hasattr(cv2.utils, "logging"):
                cv2.utils.logging.setLogLevel(cv2.utils.logging.LOG_LEVEL_ERROR)
            elif hasattr(cv2, "setLogLevel"):
                cv2.setLogLevel(3)  # ERROR
        except Exception:
            pass

        # Baca gambar dengan OpenCV (selaras dengan paddleOCR.py)
        img = cv2.imread(image_path)
        if img is None:
            return {"error": "Gambar tidak ditemukan atau tidak bisa dibaca"}

        model = _load_yolo()

        # YOLO detection (seperti snippet kamu)
        buf_out = StringIO()
        buf_err = StringIO()
        with redirect_stdout(buf_out), redirect_stderr(buf_err):
            results = model(img, conf=0.25, verbose=False)

        if not results or results[0].boxes is None or len(results[0].boxes) == 0:
            return {"error": "Plat tidak terdeteksi"}

        boxes = results[0].boxes
        # Pakai box pertama agar konsisten dengan paddleOCR.py
        x1, y1, x2, y2 = boxes.xyxy[0].cpu().numpy().astype(int).tolist()
        confidence = float(boxes.conf[0].item())

        # crop aman + padding supaya huruf tepi tidak terpotong
        h, w = img.shape[:2]
        pad_x = int((x2 - x1) * 0.08)
        pad_y = int((y2 - y1) * 0.12)
        x1 = max(0, min(w - 1, x1 - pad_x))
        x2 = max(0, min(w, x2 + pad_x))
        y1 = max(0, min(h - 1, y1 - pad_y))
        y2 = max(0, min(h, y2 + pad_y))
        if x2 <= x1 or y2 <= y1:
            return {"error": "Crop plat tidak valid"}

        plate_img = img[y1:y2, x1:x2]

        # FIX WAJIB: perbesar plat sebelum OCR (sesuai snippet)
        plate_img = cv2.resize(plate_img, None, fx=3, fy=3, interpolation=cv2.INTER_CUBIC)
        # Debug: simpan crop terakhir untuk dibandingkan
        try:
            cv2.imwrite(str(_script_dir() / "last_crop.jpg"), plate_img)
        except Exception:
            pass

        # OCR PaddleOCR
        plate_gray = cv2.cvtColor(plate_img, cv2.COLOR_BGR2GRAY)
        plate_gray = cv2.GaussianBlur(plate_gray, (3, 3), 0)
        _, plate_th = cv2.threshold(plate_gray, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
        ocr_img = cv2.cvtColor(plate_th, cv2.COLOR_GRAY2BGR)
        # Debug: simpan hasil threshold (selaras dengan paddleOCR.py)
        try:
            cv2.imwrite(str(_script_dir() / "last_crop_th.jpg"), ocr_img)
        except Exception:
            pass

        ocr = _load_paddle_ocr()
        # OCR dua jalur: warna asli + threshold, pilih hasil terbaik
        results = []
        for oimg in (plate_img, ocr_img):
            result = ocr.ocr(oimg, cls=True)
            texts = []
            for line in result:
                for word in line:
                    texts.append(word[1][0])
            results.append(" ".join(texts).strip())

        # Final cleaning untuk plat Indonesia (siap dipakai Go scraper)
        raw_text, cleaned_plate = clean_ocr_text([r for r in results if r])

        return {
            "plate_raw": raw_text,
            "plate_cleaned": cleaned_plate,
            "confidence": round(confidence, 4),
        }
    except Exception as e:
        return {"error": str(e).strip() or "Unknown error"}


if __name__ == "__main__":
    image_path = _parse_image_path(sys.argv)
    if not image_path:
        _print_json({"error": "Path gambar belum diberikan"})
        raise SystemExit(1)

    result = detect_plate(image_path)
    _print_json(result)
    raise SystemExit(0 if "error" not in result else 1)
