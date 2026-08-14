# FastSS - Smart CLI Screenshot & Auto-Annotator in Go

**FastSS** adalah aplikasi Command Line Interface (CLI) berbasis Golang untuk menangkap gambar layar/window (*screenshot*) dan secara otomatis menambahkan anotasi visual (kotak merah, panah, highlight, blur/sensor, badge langkah) pada teks target menggunakan pendeteksian OCR bawaan Windows tanpa perlu menginstal dependensi eksternal C++.

---

## ✨ Fitur Utama

1. **Window & Screen Capture**:
   - Tangkap layar penuh (*fullscreen*) atau jendela aplikasi tertentu (`ss "Chrome"` / `ss -w "active"`).
   - **Force Foreground**: Otomatis mengangkat aplikasi target ke posisi paling depan di atas terminal/CMD saat pengambilan screenshot.
   - Dukungan *countdown delay* (`-d 3`) untuk memberi waktu berpindah jendela sebelum layar ditangkap.
   - Perintah `ss list` untuk melihat daftar semua jendela aplikasi yang sedang terbuka beserta PID dan ukurannya.

2. **Deteksi Teks Otomatis (OCR + Fuzzy Matching)**:
   - Menggunakan engine OCR bawaan Windows (`Windows.Media.Ocr`) sehingga tidak memerlukan instalasi Tesseract manual.
   - Dilengkapi algoritma **Fuzzy Matching (Levenshtein Distance)** sehingga teks tetap terdeteksi meski terjadi deviasi pembacaan OCR.

3. **Anotasi Otomatis Lengkap**:
   - **Kotak (`--box`)**: Menggambar garis batas (*bounding box*) di sekeliling teks target.
   - **Panah (`--arrow`)**: Menggambar panah presisi yang menunjuk langsung ke teks target (`--arrow-from top-left`, `bottom`, `right`, dll).
   - **Highlight (`--highlight`)**: Menyorot teks target dengan warna transparan seperti stabilo.
   - **Blur / Sensor (`--blur`)**: Melakukan sensor pixelation pada teks sensitif (seperti password, token, data rahasia).
   - **Badge (`--badge`)**: Menambahkan label penomoran langkah (misal: `--badge "Submit:Langkah 1"`).

4. **Shorthand Execution (`ss` & `/ss`)**:
   - Eksekusi instan di terminal menggunakan perintah **`ss`**.
   - Integrasi AI chat command menggunakan **`/ss`**.

5. **Penyimpanan & Konfigurasi Fleksibel (`.env`)**:
   - Mendukung kustomisasi folder penyimpanan melalui file konfigurasi `.env`.
   - Default tersimpan ke `picture/screnshoot/`.

---

## ⚡ Shorthand Command (`ss` & `/ss`)

### 1. Terminal / PowerShell / CMD
Anda dapat menggunakan perintah singkat **`ss`** dari terminal mana saja:

```powershell
# Tangkap Chrome dan beri panah merah ke teks "Review"
ss "Chrome" --arrow "Review" --color red

# Tangkap fullscreen dan buat kotak di teks "Login"
ss --box "Login"
```

### 2. Antigravity Chat Slash Command
Anda juga dapat menjalankan perintah langsung melalui kolom chat AI:

```text
/ss "Chrome" --arrow "Review" --color red
/ss --box "Save Changes"
/ss list
```

---

## ⚙️ Konfigurasi Environment (`.env`)

Anda dapat menentukan folder penyimpanan default untuk semua screenshot melalui file `.env`.

1. Salin file template `.env.example` menjadi `.env`:
   ```powershell
   Copy-Item .env.example .env
   ```

2. Ubah variabel sesuai kebutuhan Anda di `.env`:
   ```env
   # Path folder penyimpanan screenshot (relatif atau absolut)
   FASTSS_OUTPUT_DIR=picture/screnshoot

   # Warna default anotasi
   FASTSS_DEFAULT_COLOR=red

   # Ketebalan garis (pixel)
   FASTSS_STROKE_WIDTH=4.0
   ```

---

## 📖 Contoh Penggunaan Lengkap

### 1. Menangkap Jendela Spesifik & Menambahkan Panah
```powershell
# Langsung gunakan nama aplikasi sebagai argumen pertama:
ss "Chrome" --arrow "Submit" --color red

# Atau menggunakan flag -w:
ss -w "Notepad" --arrow "File"
```

### 2. Menangkap Layar & Membuat Kotak di Teks Tertentu
```powershell
ss --box "Download"
```

### 3. Kombinasi Anotasi (Kotak, Panah, Highlight, Blur)
```powershell
# Tangkap jendela aktif setelah 3 detik, beri kotak pada 'Save', panah ke 'Save', highlight 'Online', dan blur token rahasia
ss -w "active" -d 3 \
  --box "Save Changes" \
  --arrow "Save Changes" \
  --highlight "Online" \
  --blur "secret-token"
```

### 4. Mengubah Warna dan Ketebalan Garis Anotasi
```powershell
# Menggunakan warna hijau dengan ketebalan 6px
ss --box "Settings" --color green --stroke 6
```
*Pilihan warna: `red`, `green`, `blue`, `yellow`, `orange`, `magenta`, `cyan`, atau kode HEX seperti `#FF5733`.*

### 5. Melihat Daftar Jendela Aplikasi yang Terbuka
```powershell
ss list
# atau
ss -l
```

### 6. Menganotasi File Gambar yang Sudah Ada
```powershell
ss annotate "gambar_lama.png" --box "Submit" --arrow "Submit" -o "hasil.png"
```

---

## 📂 Struktur Proyek

```text
fastss/
├── .agents/
│   └── skills/ss/    # Antigravity chat slash command integration (/ss)
├── cmd/
│   ├── root.go       # Konfigurasi flag dan runner utama CLI
│   ├── list.go       # Subcommand list open windows
│   └── annotate.go   # Subcommand anotasi gambar eksisting
├── internal/
│   ├── capture/      # Screen & window capture (Win32 API & Screenshot)
│   ├── ocr/          # Windows Media OCR & Fuzzy text matcher
│   ├── draw/         # Rendering 2D (Box, Arrow, Highlight, Blur, Badge)
│   └── storage/      # Penyimpanan file & loader konfigurasi .env
├── picture/
│   └── screnshoot/   # Direktori output default
├── .env.example      # Template konfigurasi environment
├── go.mod
├── go.sum
└── main.go           # Entry point aplikasi
```
