# FastSS - Smart CLI Screenshot & Auto-Annotator in Go

**FastSS** adalah aplikasi Command Line Interface (CLI) berbasis Golang untuk menangkap gambar layar/window (*screenshot*) dan secara otomatis menambahkan anotasi visual (kotak merah, panah, highlight, blur/sensor, badge langkah) pada teks target menggunakan pendeteksian OCR bawaan Windows tanpa perlu menginstal dependensi eksternal C++.

---

## ⚡ Instalasi Global (Agar `ss` Bisa Digunakan di Seluruh Folder)

Perintah `go build` biasa **tidak** otomatis mendaftarkan aplikasi ke PATH Windows. Agar Anda bisa mengetik **`ss`** dari folder mana saja di CMD atau PowerShell, gunakan salah satu cara berikut:

### Opsi A: Menggunakan Script Instalasi Otomatis (Direkomendasikan)
Jalankan file installer yang sudah disediakan di folder ini:
```powershell
.\install.bat
```
*Script ini otomatis meng-compile kode dan menaruh biner `ss.exe` ke folder global PATH (`%USERPROFILE%\go\bin`), sehingga langsung dapat digunakan dari seluruh direktori.*

---

### Opsi B: Manual via PowerShell (1 Baris)
Jika ingin mendaftarkan folder proyek saat ini secara permanen ke User PATH Windows:
```powershell
# Jalankan perintah ini di PowerShell (Cukup sekali):
[Environment]::SetEnvironmentVariable("Path", [Environment]::GetEnvironmentVariable("Path", "User") + ";$PWD", "User")
```
*(Setelah menjalankan perintah di atas, buka jendela CMD/PowerShell baru agar PATH ter-refresh).*

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

## 🎯 Seleksi Teks Spesifik (Targeting Akurat)

Jika terdapat banyak teks yang sama di layar (misal tombol "Edit" atau "Delete" di setiap baris tabel), Anda bisa menargetkan teks yang tepat dengan berbagai cara:

### 1. Menggunakan Indeks Urutan (`[1]`, `[2]`, `[last]`)
```powershell
# Menandai kemunculan pertama saja (default):
ss --box "Edit[1]"

# Menandai kemunculan ke-2:
ss --box "Edit[2]"

# Menandai kemunculan terakhir:
ss --box "Edit[last]"

# Menandai SEMUA kemunculan teks 'Edit':
ss --box "Edit[all]"
# atau menggunakan flag --all:
ss --box "Edit" --all
```

### 2. Berdasarkan Konteks / Teks Terdekat (`near`, `below`, `above`, `right-of`)
```powershell
# Menandai tombol "Edit" yang berada di dekat tulisan "Billing":
ss --box "Edit near Billing"

# Menandai tombol "Save" yang posisinya di bawah "Settings":
ss --arrow "Save below Settings"

# Menandai tombol "Delete" yang berada di sebelah kanan "Account":
ss --box "Delete right-of Account"
```

### 3. Berdasarkan Area Layar (`top`, `bottom`, `left`, `right`, `top-right`, dll)
```powershell
# Menandai tombol 'Save' yang berada di pojok kanan atas:
ss --box "top-right:Save"

# Menandai tombol 'Submit' di bagian bawah layar:
ss --box "bottom:Submit"
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
├── install.bat       # Script installer otomatis ke global PATH
├── go.mod
├── go.sum
└── main.go           # Entry point aplikasi
```
