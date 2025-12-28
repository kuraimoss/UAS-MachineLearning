(() => {
  const $ = (sel) => document.querySelector(sel);
  const baseUrl = window.location.origin;

  const readJSON = (id) => {
    const el = $(id);
    if (!el) return {};
    try {
      const first = JSON.parse(el.textContent || "{}");
      if (typeof first === "string") {
        const s = first.trim();
        if ((s.startsWith("{") && s.endsWith("}")) || (s.startsWith("[") && s.endsWith("]"))) {
          try {
            return JSON.parse(s);
          } catch {
            return first;
          }
        }
      }
      return first;
    } catch {
      return {};
    }
  };

  const sampleSuccess = readJSON("#docsSampleSuccess");
  const sampleError = readJSON("#docsSampleError");

  const baseEl = $("#baseUrl");
  if (baseEl) baseEl.textContent = baseUrl;

  const curl = `curl -X POST "${baseUrl}/detect" \\\n  -F "image=@/path/to/plate.jpg"`;
  const health = `curl "${baseUrl}/health"`;

  const setText = (id, txt) => {
    const el = $(id);
    if (el) el.textContent = txt;
  };

  setText("#curlSnippet", curl);
  setText("#healthSnippet", health);
  setText("#postmanUrl", `${baseUrl}/detect`);

  const escapeHtml = (s) =>
    String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");

  const highlightJSON = (obj) => {
    const json = JSON.stringify(obj, null, 2);
    const esc = escapeHtml(json);
    return esc
      .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(?:\s*:)?)/g, (m) => {
        return m.endsWith(":") ? `<span class="k">${m}</span>` : `<span class="s">${m}</span>`;
      })
      .replace(/\b(true|false)\b/g, `<span class="b">$1</span>`)
      .replace(/\bnull\b/g, `<span class="nu">null</span>`)
      .replace(/(-?\d+(?:\.\d+)?)/g, `<span class="n">$1</span>`);
  };

  const setCode = (sel, obj) => {
    const el = $(sel);
    if (!el) return;
    el.innerHTML = highlightJSON(obj);
  };

  setCode("#sampleSuccess code", sampleSuccess);
  setCode("#successJson", sampleSuccess);
  setCode("#errorJson", sampleError);

  // Scroll reveal
  const revealEls = Array.from(document.querySelectorAll(".reveal"));
  const io = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) e.target.classList.add("in");
      }
    },
    { threshold: 0.12 }
  );
  revealEls.forEach((el) => io.observe(el));

  // Try it
  const fileInput = $("#fileInput");
  const cameraInput = $("#cameraInput");
  const fileName = $("#fileName");
  const preview = $("#preview");
  const previewImg = $("#previewImg");
  const sendBtn = $("#sendBtn");
  const resetBtn = $("#resetBtn");
  const timing = $("#timing");
  const reqid = $("#reqid");
  const panelStatus = $("#panelStatus");
  const skeleton = $("#skeleton");
  const responsePre = $("#responsePre");
  const responseCode = $("#responseCode");
  const copyBtn = $("#copyBtn");

  let selectedFile = null;
  let previewURL = "";

  const setLoading = (on) => {
    if (!sendBtn) return;
    sendBtn.classList.toggle("loading", on);
    sendBtn.disabled = on;
    if (skeleton) skeleton.classList.toggle("show", on);
    if (responsePre) responsePre.style.visibility = on ? "hidden" : "visible";
  };

  const setPanel = (obj, ok, httpStatus) => {
    if (panelStatus) panelStatus.textContent = ok ? `HTTP ${httpStatus}` : `HTTP ${httpStatus}`;
    if (responseCode) responseCode.innerHTML = highlightJSON(obj);
  };

  const resetUI = () => {
    if (fileInput) fileInput.value = "";
    if (cameraInput) cameraInput.value = "";
    if (fileName) fileName.textContent = "Belum ada file";
    selectedFile = null;
    if (preview) preview.classList.remove("has-img");
    if (previewImg) previewImg.removeAttribute("src");
    if (previewURL) {
      URL.revokeObjectURL(previewURL);
      previewURL = "";
    }
    if (timing) timing.textContent = "";
    if (reqid) reqid.textContent = "";
    if (panelStatus) panelStatus.textContent = "";
    if (responseCode) responseCode.textContent = "";
  };

  const onPick = (f) => {
    if (!f) return;
    selectedFile = f;
    if (fileName) fileName.textContent = f.name || "image";
    if (previewURL) URL.revokeObjectURL(previewURL);
    previewURL = URL.createObjectURL(f);
    if (previewImg) previewImg.src = previewURL;
    if (preview) preview.classList.add("has-img");
  };

  if (fileInput) {
    fileInput.addEventListener("change", () => onPick(fileInput.files?.[0] || null));
  }
  if (cameraInput) {
    cameraInput.addEventListener("change", () => onPick(cameraInput.files?.[0] || null));
  }

  if (resetBtn) resetBtn.addEventListener("click", resetUI);

  if (sendBtn) {
    sendBtn.addEventListener("click", async () => {
      const f = selectedFile;
      if (!f) {
        setPanel({ status: "error", error: { code: "MISSING_FILE", message: "Pilih file dulu" } }, false, 400);
        return;
      }

      const fd = new FormData();
      fd.append("image", f);

      setLoading(true);
      const t0 = performance.now();

      try {
        const resp = await fetch(`${baseUrl}/detect`, { method: "POST", body: fd });
        const httpStatus = resp.status;
        const text = await resp.text();
        let data;
        try {
          data = JSON.parse(text);
        } catch {
          data = { status: "error", error: { code: "BAD_RESPONSE", message: "Response bukan JSON" }, raw: text };
        }

        const ok = resp.ok && data?.status === "success";
        setPanel(data, ok, httpStatus);

        const t1 = performance.now();
        const apiMs = data?.processing_time_ms ?? null;
        if (timing) {
          timing.textContent = apiMs != null ? `processing_time_ms: ${apiMs} • browser: ${Math.round(t1 - t0)}ms` : `browser: ${Math.round(t1 - t0)}ms`;
        }
        const rid = data?.request_id || resp.headers.get("X-Request-Id") || "";
        if (reqid && rid) reqid.textContent = `request_id: ${rid}`;
      } catch (e) {
        setPanel({ status: "error", error: { code: "NETWORK_ERROR", message: String(e) } }, false, 0);
      } finally {
        setLoading(false);
      }
    });
  }

  if (copyBtn) {
    copyBtn.addEventListener("click", async () => {
      const txt = responseCode?.textContent || "";
      if (!txt) return;
      try {
        await navigator.clipboard.writeText(txt);
        copyBtn.textContent = "Copied";
        setTimeout(() => (copyBtn.textContent = "Copy JSON"), 900);
      } catch {
        // ignore
      }
    });
  }
})();
