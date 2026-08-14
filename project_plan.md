# Perencanaan Proyek CLI: Auto Screenshot & Annotator

Dokumen ini berisi perencanaan proyek untuk aplikasi CLI menggunakan bahasa pemrograman Go (Golang) yang berfungsi untuk menangkap layar (screenshot), mencari teks tertentu, dan menambahkan anotasi (panah, kotak) secara otomatis.

## 1. Arsitektur & Alur Kerja (Workflow)

Aplikasi ini akan berjalan melalui Command Line Interface (CLI) dengan alur kerja berikut:
1. **Parsing Perintah**: CLI menerima argumen dari pengguna (contoh: `--window "Nama Window" --action "box:Teks Target"`).
2. **Screen Capture**: Mengambil gambar dari window yang dituju atau seluruh layar.
3. **Optical Character Recognition (OCR)**: Membaca teks di dalam gambar untuk menemukan koordinat (x, y) dari teks yang dicari.
4. **Image Processing & Drawing**: Menggambar anotasi (kotak merah, panah merah) pada koordinat yang ditemukan.
5. **Penyimpanan**: Menyimpan gambar hasil editan ke direktori target (`picture/screnshoot`).

## 2. Stack Teknologi & Library (Golang)

Untuk membangun proyek ini di Go, berikut adalah rekomendasi library yang digunakan:

*   **CLI Framework**: `github.com/spf13/cobra`
    *   *Fungsi*: Membuat struktur CLI yang rapi (mirip dengan `kubectl` atau `docker`).
*   **Screen Capture**: `github.com/kbinani/screenshot`
    *   *Fungsi*: Menangkap gambar dari layar monitor. (Catatan: Untuk menangkap spesifik *window*, mungkin diperlukan interaksi dengan Windows API menggunakan CGO atau package `golang.org/x/sys/windows`).
*   **OCR (Pendeteksi Teks)**: `github.com/otiai10/gosseract/v2`
    *   *Fungsi*: Wrapper untuk **Tesseract OCR**. Library ini digunakan untuk mengubah gambar menjadi teks beserta data posisi (bounding box) teks tersebut di dalam gambar.
*   **Image Drawing / Anotasi**: `github.com/fogleman/gg`
    *   *Fungsi*: Library rendering 2D yang sangat baik untuk menggambar bentuk (kotak, garis, panah) di atas gambar yang sudah ditangkap.

## 3. Struktur Direktori Proyek

```text
autoshot-cli/
├── cmd/
│   ├── root.go       # Konfigurasi utama CLI (Cobra)
│   ├── capture.go    # Sub-command untuk menangkap dan mengedit
├── internal/
│   ├── capture/      # Logika untuk screenshot layar/window
│   ├── ocr/          # Logika untuk mendeteksi teks dan bounding box
│   ├── draw/         # Logika untuk menggambar kotak/panah
│   └── storage/      # Logika untuk menyimpan file ke folder
├── picture/
│   └── screnshoot/   # Folder output default (akan dibuat otomatis jika belum ada)
├── main.go           # Entry point aplikasi
├── go.mod
└── go.sum
```

## 4. Fase Pengembangan (Roadmap)

*   **Fase 1: Setup & Screenshot Dasar**
    *   Inisialisasi Cobra CLI.
    *   Implementasi fitur screenshot seluruh layar atau area tertentu dan menyimpannya ke folder `picture/screnshoot`.
*   **Fase 2: Integrasi OCR**
    *   Mengintegrasikan `gosseract` untuk membaca teks dari screenshot.
    *   Membuat fungsi untuk mendapatkan koordinat (X, Y, Width, Height) dari kata/teks spesifik yang dicari.
*   **Fase 3: Rendering & Anotasi**
    *   Menggunakan library `gg` untuk menggambar kotak merah (`stroke`) di atas gambar berdasarkan koordinat dari Fase 2.
    *   Membuat kalkulasi geometri dasar untuk menggambar bentuk panah yang menunjuk ke koordinat teks.
*   **Fase 4: Integrasi & Polishing**
    *   Menggabungkan semua fase menjadi satu perintah CLI yang utuh.
    *   Menambahkan error handling (misal: jika teks tidak ditemukan oleh OCR).

---

## 5. Saran dan Kritik (Critical Review)

Terdapat beberapa tantangan teknis dalam ide proyek ini yang perlu Anda pertimbangkan sebelum memulai pengembangan:

### 1. Ketergantungan Terhadap Tesseract OCR (Kritik)
Mencari teks dalam sebuah gambar *mengharuskan* penggunaan teknologi OCR (Optical Character Recognition). Library seperti `gosseract` membutuhkan **Tesseract C++ library** untuk diinstal di OS pengguna.
*   **Dampak**: Aplikasi CLI Anda tidak akan menjadi sebuah *single binary* yang portable. Pengguna harus menginstal Tesseract secara manual di sistem Windows/Linux/Mac mereka agar CLI ini bisa berjalan.
*   **Saran Alternatif**: Jika aplikasi CLI ini spesifik untuk web atau GUI tertentu, Anda bisa mempertimbangkan menggunakan *Automation tools* (seperti Selenium/Playwright) yang mencari elemen HTML/UI Node dan mengambil koordinatnya secara langsung tanpa harus memproses gambar menggunakan OCR.

### 2. Tangkapan Layar Spesifik "Window" (Tantangan Windows)
Menangkap layar hanya untuk satu window aplikasi tertentu (bukan seluruh layar) cukup rumit dan sangat spesifik terhadap Sistem Operasi (OS).
*   **Dampak**: Library standar biasanya hanya bisa mengambil screenshot per *display/monitor*.
*   **Saran**: Pada tahap awal (MVP), ubah fungsionalitasnya menjadi menangkap *seluruh layar*, lalu gunakan OCR. Setelah fitur ini berjalan stabil, barulah Anda mengimplementasikan pemanggilan Windows API (`PrintWindow` atau enumerasi `EnumWindows`) untuk menangkap window spesifik.

### 3. Akurasi OCR
Akurasi Tesseract terkadang menurun pada teks yang kecil, font yang aneh, atau kontras warna yang rendah.
*   **Saran**: Anda mungkin perlu menambahkan fungsi pre-processing gambar (misal: mengubah gambar menjadi hitam putih / grayscale dan meningkatkan kontras) sebelum memasukkannya ke mesin OCR agar teks lebih mudah dideteksi.

### 4. Penamaan Folder
Anda menyebutkan `picture/screnshoot`. Sebaiknya gunakan penulisan bahasa Inggris yang baku yaitu `picture/screenshot` atau mengikuti standar folder OS seperti `Pictures/Screenshots` agar lebih profesional. Namun, program tentu saja bisa diatur untuk membuat folder sesuai input Anda.
