import re
from typing import Iterable, Tuple


_PLATE_RE = re.compile(r"([A-Z]{1,2})(\d{1,4})([A-Z]{1,3})")


def _normalize(s: str) -> str:
    s = s.upper()
    s = re.sub(r"[^A-Z0-9]", "", s)
    return s


def _fix_common_ocr(s: str) -> str:
    # Heuristik ringan: perbaiki karakter yang sering ketukar.
    # Tidak dipaksakan untuk semua posisi (biar tidak merusak huruf depan/belakang).
    return (
        s.replace("O", "0")
        .replace("Q", "0")
        .replace("I", "1")
        .replace("L", "1")
        .replace("Z", "2")
        .replace("S", "5")
        .replace("B", "8")
        .replace("G", "6")
    )


def clean_ocr_text(texts: Iterable[str]) -> Tuple[str, str]:
    texts = [t for t in texts if isinstance(t, str) and t.strip()]
    raw = " ".join(t.strip() for t in texts).strip()
    if not raw:
        return "", ""

    candidates = []
    for t in texts:
        n = _normalize(t)
        if n:
            candidates.append(n)

    joined = _normalize(raw)
    if joined:
        candidates.append(joined)

    # Cari kandidat yang match pola plat Indonesia.
    best = ""
    for c in candidates:
        m = _PLATE_RE.search(c)
        if m:
            plate = "".join(m.groups())
            best = plate
            break

    if not best:
        # Fallback: coba setelah perbaikan OCR umum
        for c in candidates:
            f = _fix_common_ocr(c)
            m = _PLATE_RE.search(f)
            if m:
                best = "".join(m.groups())
                break

    if not best:
        # Fallback terakhir: hasil normalisasi saja (dibatasi panjang).
        best = candidates[-1] if candidates else joined
        best = best[:9]

    return raw, best

