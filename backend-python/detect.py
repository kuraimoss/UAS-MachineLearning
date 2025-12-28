import json
import os
import re
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
TROCR_PROCESSOR = None
TROCR_MODEL = None


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


def _load_trocr():
    global TROCR_PROCESSOR, TROCR_MODEL
    if TROCR_PROCESSOR is not None and TROCR_MODEL is not None:
        return TROCR_PROCESSOR, TROCR_MODEL

    buf_out = StringIO()
    buf_err = StringIO()
    with redirect_stdout(buf_out), redirect_stderr(buf_err):
        import torch  # type: ignore
        from transformers import (  # type: ignore
            TrOCRProcessor,
            VisionEncoderDecoderModel,
        )

        model_name = os.getenv("OCR_MODEL_NAME", "microsoft/trocr-base-printed").strip()
        device = os.getenv("OCR_DEVICE", "").strip() or ("cuda" if torch.cuda.is_available() else "cpu")

        processor = TrOCRProcessor.from_pretrained(model_name)
        model = VisionEncoderDecoderModel.from_pretrained(model_name)
        model.to(device)
        model.eval()

    TROCR_PROCESSOR, TROCR_MODEL = processor, model
    return TROCR_PROCESSOR, TROCR_MODEL


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

        # Baca gambar pakai PIL (hindari warning OpenCV imread)
        # iPhone sering menyimpan rotasi di EXIF orientation, jadi harus di-transpose dulu.
        try:
            from PIL import Image, ImageOps  # type: ignore

            img_pil = ImageOps.exif_transpose(Image.open(image_path)).convert("RGB")
            img = np.array(img_pil)
            img = cv2.cvtColor(img, cv2.COLOR_RGB2BGR)
        except Exception:
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
        best_idx = int(boxes.conf.argmax().item())

        x1, y1, x2, y2 = boxes.xyxy[best_idx].cpu().numpy().astype(int).tolist()
        confidence = float(boxes.conf[best_idx].item())

        # crop aman
        h, w = img.shape[:2]
        x1 = max(0, min(w - 1, x1))
        x2 = max(0, min(w, x2))
        y1 = max(0, min(h - 1, y1))
        y2 = max(0, min(h, y2))
        if x2 <= x1 or y2 <= y1:
            return {"error": "Crop plat tidak valid"}

        plate_img = img[y1:y2, x1:x2]

        # FIX WAJIB: perbesar plat sebelum OCR (sesuai snippet)
        plate_img = cv2.resize(plate_img, None, fx=2, fy=2, interpolation=cv2.INTER_CUBIC)

        # OCR TroCR (sesuai snippet kamu: processor + model.generate)
        processor, ocr_model = _load_trocr()
        device = next(ocr_model.parameters()).device

        from PIL import Image  # type: ignore
        import torch  # type: ignore

        plate_pil = Image.fromarray(cv2.cvtColor(plate_img, cv2.COLOR_BGR2RGB))
        pixel_values = processor(images=plate_pil, return_tensors="pt").pixel_values.to(device)

        with torch.no_grad():
            generated_ids = ocr_model.generate(
                pixel_values,
                max_length=16,
                num_beams=5,
                early_stopping=True,
            )

        text = processor.batch_decode(generated_ids, skip_special_tokens=True)[0]

        cleaned_text = (text or "").upper()
        cleaned_text = re.sub(r"[^A-Z0-9]", "", cleaned_text)

        # Final cleaning untuk plat Indonesia (siap dipakai Go scraper)
        _raw, cleaned_plate = clean_ocr_text([cleaned_text])

        return {
            "plate_raw": text,
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
