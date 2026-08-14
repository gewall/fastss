---
name: ss
description: Eksekusi tangkapan layar dan anotasi otomatis FastSS dengan cepat menggunakan perintah /ss.
---

# FastSS `/ss` Command Handler

Skill ini menangani pemanggilan shorthand `/ss` dari pengguna untuk mengambil tangkapan layar (*screenshot*) dan memberikan anotasi secara otomatis.

## Cara Eksekusi:
Ketika pengguna mengetik `/ss [argumen]` atau meminta screenshot/anotasi:
Jalankan biner `C:\Algi\fastss\ss.exe` atau `fastss.exe` menggunakan tool `run_command` dengan argumen yang diminta pengguna.

### Contoh Mapping Argumen:
- `/ss --box "teks"` -> `C:\Algi\fastss\ss.exe --box "teks"`
- `/ss -w "Chrome" --arrow "Login"` -> `C:\Algi\fastss\ss.exe -w "Chrome" --arrow "Login"`
- `/ss list` atau `/ss -l` -> `C:\Algi\fastss\ss.exe list`
- `/ss annotate "file.png" --box "teks"` -> `C:\Algi\fastss\ss.exe annotate "file.png" --box "teks"`
