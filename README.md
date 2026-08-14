# FastSS - Smart CLI Screenshot & Auto-Annotator in Go

**FastSS** adalah aplikasi Command Line Interface (CLI) berbasis Golang untuk menangkap gambar layar/window (*screenshot*) dan secara otomatis menambahkan anotasi visual (kotak merah, panah, highlight, blur/sensor, badge langkah) pada teks target menggunakan pendeteksian OCR bawaan Windows tanpa perlu menginstal dependensi eksternal C++.

---

## ✨ Fitur Utama

1. **Window & Screen Capture**:
   - Tangkap layar penuh (*fullscreen*) atau jendela aplikasi tertentu (`--window "Chrome"` / `--window "active"`).
   - Dukungan *countdown delay* (`--delay 3`) untuk memberi waktu berpindah jendela sebelum layar ditangkap.
   - Perintah `fastss list` untuk melihat daftar semua jendela aplikasi yang sedang terbuka beserta PID dan ukurannya.

2. **Deteksi Teks Otomatis (OCR + Fuzzy Matching)**:
   - Menggunakan engine OCR bawaan Windows (`Windows.Media.Ocr`) sehingga tidak memerlukan instalasi Tesseract manual.
   - Dilengkapi algoritma **Fuzzy Matching (Levenshtein Distance)** sehingga teks tetap terdeteksi meski terjadi sedikit *typo* / deviasi karakter dari pembacaan OCR.

3. **Anotasi Otomatis Lengkap**:
   - **Kotak (`--box`)**: Menggambar garis batas (*bounding box*) merah di sekeliling teks target.
   - **Panah (`--arrow`)**: Menggambar panah presisi yang menunjuk langsung ke teks target, dengan arah panah yang dapat diatur (`--arrow-from top-left`, `bottom`, dll).
   - **Highlight (`--highlight`)**: Menyorot teks target dengan warna transparan seperti stabilo.
   - **Blur / Sensor (`--blur`)**: Melakukan sensor pixelation pada teks sensitif (seperti password, token, nomor telepon).
   - **Badge (`--badge`)**: Menambahkan label penomoran langkah (misal: `fastss --badge "Submit:Langkah 1"`).

4. **Penyimpanan Otomatis**:
   - Gambar otomatis disimpan ke folder `picture/screnshoot/` dengan penamaan timestamp (misal: `screenshot_20260814_110500.png`).
   - Mendukung format output kustom PNG / JPEG.

5. **Anotasi Gambar Eksisting (`fastss annotate`)**:
   - Mampu menganotasi file gambar yang sudah ada tanpa harus mengambil screenshot baru.

---

## 🚀 Cara Menjalankan

### 1. Build Executable
```bash
go build -o fastss.exe .
```

---

## 📖 Contoh Penggunaan

### 1. Menangkap Jendela Tertentu & Menambahkan Panah Merah
```bash
# Menunjuk tombol "Login" pada jendela Chrome
.\fastss.exe -w "Chrome" --arrow "Login"
```

### 2. Menangkap Layar & Membuat Kotak Merah di Teks Tertentu
```bash
# Menggambar kotak merah di sekitar teks "Download"
.\fastss.exe --box "Download"
```

### 3. Kombinasi Anotasi (Kotak, Panah, Highlight, Blur)
```bash
# Tangkap jendela aktif setelah 3 detik, beri kotak pada 'Save', panah ke 'Save', highlight 'Online', dan blur token rahasia
.\fastss.exe -w "active" -d 3 \
  --box "Save Changes" \
  --arrow "Save Changes" \
  --highlight "Online" \
  --blur "secret-token"
```

### 4. Mengubah Warna dan Ketebalan Garis Anotasi
```bash
# Menggunakan warna hijau dengan ketebalan garis 6px
.\fastss.exe --box "Settings" --color green --stroke 6
```
Pilihan warna: `red`, `green`, `blue`, `yellow`, `orange`, `magenta`, `cyan`, atau kode HEX seperti `#FF5733`.

### 5. Melihat Daftar Jendela Aplikasi yang Terbuka
```bash
.\fastss.exe list
# atau
.\fastss.exe -l
```

### 6. Menganotasi File Gambar yang Sudah Ada
```bash
.\fastss.exe annotate "gambar_lama.png" --box "Submit" --arrow "Submit" -o "hasil.png"
```

---

## 📂 Struktur Proyek

```text
fastss/
├── cmd/
│   ├── root.go       # Konfigurasi flag dan runner utama CLI
│   ├── list.go       # Subcommand list open windows
│   └── annotate.go   # Subcommand anotasi gambar eksisting
├── internal/
│   ├── capture/      # Screen & window capture (Win32 API & Screenshot)
│   ├── ocr/          # Windows Media OCR & Fuzzy text matcher
│   ├── draw/         # Rendering 2D (Box, Arrow, Highlight, Blur, Badge)
│   └── storage/      # Penyimpanan file otomatis ke picture/screnshoot
├── picture/
│   └── screnshoot/   # Direktori output default
├── testimg/          # Script generator pengujian anotasi & OCR
├── go.mod
├── go.sum
└── main.go           # Entry point aplikasi
```
