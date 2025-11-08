# /speckit.constitution — Fitur Constitution (Backend Go 1.24)

Tujuan
- Menetapkan prinsip pengembangan, aturan arsitektur, dan quality gates.
- Fokus: DDD + Clean Architecture + TDD untuk backend Go 1.24.

Ringkasan Fitur
- Nama fitur: Get Available Date For Playback Personel
- Deskripsi singkat: Mendapatkan ada di tanggal berapa saja yang memiliki data di influx di bulan ini (param date)

Prinsip Inti
- Domain-first: logika bisnis murni di lapisan domain.
- Use case–oriented: aplikasi berpusat pada kasus penggunaan.
- Dependencies ke dalam: kontrak didefinisikan dari sisi domain/application.
- Ports & Adapters: boundary jelas; implementasi di tepi (adapter/infra).
- Pure domain: tanpa I/O, logging, atau context di domain.
- TDD disiplin: red → green → refactor untuk setiap perubahan.

Prinsip SOLID
- Single Responsibility Principle (SRP): setiap modul/class/package memiliki satu alasan untuk berubah. Domain entity fokus pada aturan bisnis; adapter fokus pada I/O/presentasi.
- Open/Closed Principle (OCP): komponen terbuka untuk ekstensi, tertutup untuk modifikasi. Tambahkan adapter baru tanpa mengubah use case/ports; perluas lewat implementasi baru, bukan edit inti.

Praktik Implementasi untuk SOLID
- Gunakan injeksi dependensi via konstruktor; hindari global mutable state.
- Konfigurasi diisolasi dalam `Config` dan diteruskan ke layer luar (adapters), bukan domain.
- Error pakai sentinel dan dibungkus di adapter dengan konteks; mapping ke katalog error di interface REST.
- Jaga boundary: mapping DTO ↔ entity dilakukan di adapters; domain tidak mengetahui format transport.