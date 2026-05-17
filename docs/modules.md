# Modul Aplikasi

## Daftar Modul

| Modul | Status | Keterangan |
|-------|--------|------------|
| [Autentikasi](#autentikasi) | Wajib | Login, register, RBAC |
| [Siswa](#siswa) | Wajib | Data siswa |
| [Kelas](#kelas) | Wajib | Data kelas |
| [Pembayaran](#pembayaran) | Wajib | Tagihan & pembayaran |
| [PPDB](#ppdb) | Wajib | Pendaftaran siswa baru |
| [Notifikasi](#notifikasi) | Opsional | WhatsApp, Telegram, Email |
| [Laporan](#laporan) | Wajib | Dashboard & export |
| [Upload](#upload--file-management) | Wajib | File upload & signed URL |
| [Backup](#backup) | Wajib | Backup & restore |

---

## Autentikasi

### Fitur
- Login email/password
- Login dengan Google (OAuth2)
- Role-based access control (RBAC)
- Ganti password

### Role
| Role | Akses |
|------|-------|
| admin | Semua fitur |
| operator | Pembayaran, PPDB, Siswa |
| guru | Data kelas yang diajar |
| siswa | Data sendiri, tagihan, pembayaran |
| orangtua | Data anak, tagihan, pembayaran |

### Alur Login
```
1. User masukkan email & password
2. Backend validasi credentials
3. Generate JWT token
4. Set httpOnly cookie
5. Return user data & redirect berdasarkan role
```

### Alur Login Google
```
1. User klik "Login dengan Google"
2. Frontend redirect ke Google OAuth
3. Google return authorization code
4. Backend exchange code untuk user info
5. Cari/buat user di database
6. Generate JWT token
7. Set httpOnly cookie
```

---

## Siswa

### Fitur
- CRUD data siswa
- Import dari Excel
- Export ke Excel
- Pencarian & filter
- Upload foto
- Status siswa (aktif, lulus, pindah, keluar)
- Riwayat kelas per tahun ajaran
- Riwayat pembayaran

### Data Siswa
- NIS (Nomor Induk Siswa)
- Nama lengkap
- Jenis kelamin
- Tempat, tanggal lahir
- Agama
- Alamat
- No HP, email
- Foto
- Data orangtua (nama, no HP, email)
- Tahun ajaran masuk
- Status

### Import Excel
Format kolom Excel untuk import:

| Kolom | Field | Wajib |
|-------|-------|-------|
| A | NIS | Ya |
| B | Nama | Ya |
| C | Jenis Kelamin (L/P) | Ya |
| D | Tempat Lahir | - |
| E | Tanggal Lahir (YYYY-MM-DD) | - |
| F | Agama | - |
| G | Alamat | - |
| H | No HP | - |
| I | Email | - |
| J | Nama Orangtua | - |
| K | No HP Orangtua | - |
| L | Email Orangtua | - |

---

## Kelas

### Fitur
- CRUD data kelas
- Tambah/hapus siswa dari kelas
- Daftar siswa per kelas
- Wali kelas

### Data Kelas
- Nama kelas (contoh: "7A", "X TKJ 1")
- Tingkat: SD 1-6, SMP 7-9, SMA/SMK 10-12
- Jurusan (opsional, untuk SMK)
- Wali kelas
- Ruangan
- Kapasitas
- Shift (pagi/sore)
- Tahun ajaran

### Jurusan (untuk SMK)
- TKJ (Teknik Komputer & Jaringan)
- RPL (Rekayasa Perangkat Lunak)
- Akuntansi
- Administrasi Perkantoran
- dll (bisa dikustomisasi)

---

## Pembayaran

### Fitur
- Kategori pembayaran fleksibel
- Tagihan per siswa
- Support cicilan (banyak pembayaran per tagihan)
- Upload bukti bayar
- Verifikasi manual oleh operator
- Payment gateway (adapter pattern)
- Export ke Excel
- Cetak kwitansi

### Alur Pembayaran Manual
```
1. Admin/Operator buat tagihan (per siswa atau massal)
2. Siswa/Orangtua lihat tagihan
3. Sistem tampilkan rekening sekolah aktif (GET /rekening-sekolah/aktif)
4. Siswa/Orangtua transfer ke salah satu rekening
5. Siswa/Orangtua upload bukti bayar + pilih rekening tujuan (rekening_sekolah_id)
6. Operator verifikasi bukti bayar (cross-check dengan rekening tujuan)
7. Status tagihan otomatis update (belum_bayar → sebagian → lunas)
```

Catatan:
- Metode `cash`: tidak butuh rekening, `rekening_sekolah_id` = NULL
- Metode `transfer`: wajib pilih rekening tujuan
- Metode `midtrans`/`xendit`: diproses via gateway, `rekening_sekolah_id` = NULL

### Alur Payment Gateway
```
1. Admin/Operator buat tagihan
2. Siswa/Orangtua pilih "Bayar Online" atau admin/operator inisiasi via API
3. Backend validasi:
   a. Tagihan ada dan milik sekolah yang sama (tenant isolation)
   b. Tagihan belum lunas
   c. Provider tersedia dan dikonfigurasi
   d. Siswa/Orangtua: tagihan milik siswa yang terhubung (pengguna_siswa)
4. Backend cek idempotency: jika sudah ada transaksi pending untuk tagihan+provider → return yang sudah ada
5. Backend generate order_id dan panggil gateway.CreateTransaction()
6. Simpan gateway_transaksi (status: pending)
7. Return payment_url ke frontend
8. User buka payment_url di browser
9. User selesaikan pembayaran di gateway
10. Gateway kirim callback ke backend
11. Backend validasi signature callback
12. Backend cek idempotency (payment_gateway_id sudah ada? → skip, return 200)
13. Backend cek overpay invariant dalam DB transaction
14. Simpan pembayaran + update status tagihan + update gateway_transaksi status (atomic)
```

### Payment Invariants

| Invariant | Keterangan |
|-----------|------------|
| No overpay | `SUM(pembayaran.jumlah WHERE status='verified' AND tagihan_id=X) <= tagihan.nominal` |
| Idempotency | Callback dengan `(provider, payment_gateway_id)` yang sudah ada → skip, return 200 OK |
| Initiation idempotency | Transaksi gateway dengan `(tagihan_id, provider, status='pending')` yang sudah ada → return yang sudah ada |
| Atomic update | Insert pembayaran + update status tagihan + update gateway_transaksi status dalam satu DB transaction |
| Signature validation | Setiap callback harus divalidasi signature sebelum diproses |
| No premature verification | Tidak ada pembayaran yang dibuat verified/lunas saat inisiasi — hanya setelah callback valid |

### Signature Validation per Gateway

| Gateway | Metode |
|---------|--------|
| Midtrans | SHA-512: `SHA512(order_id + status_code + gross_amount + server_key)` |
| Xendit | Header `x-callback-token` dibandingkan dengan token di konfigurasi |

### Idempotency Flow

```
1. Callback masuk dengan payment_gateway_id = "trx-001", provider = "midtrans"
2. Cek DB: SELECT * FROM pembayaran WHERE provider='midtrans' AND payment_gateway_id='trx-001'
3. Jika sudah ada → return 200 (tidak proses ulang)
4. Jika belum ada → lanjut proses dalam transaction:
   a. Hitung total verified: SELECT SUM(jumlah) FROM pembayaran WHERE tagihan_id=X AND status='verified'
   b. Jika total + jumlah_baru > nominal → tolak, log anomaly
   c. Jika OK → INSERT pembayaran, UPDATE tagihan.status
   d. COMMIT
```

### Adapter Payment Gateway
```go
type PaymentGateway interface {
    CreateTransaction(ctx context.Context, req TransactionRequest) (TransactionResponse, error)
    HandleCallback(ctx context.Context, payload []byte) (CallbackResult, error)
    ValidateSignature(ctx context.Context, payload []byte, headers http.Header) error
    GetStatus(ctx context.Context, transactionID string) (StatusResult, error)
}

// implementasi
- MidtransAdapter
- XenditAdapter
```

### Kategori Pembayaran Contoh
| Kategori | Keterangan |
|----------|------------|
| SPP | Sumbangan Pembinaan Pendidikan |
| Uang Pangkal | Pembayaran sekali saat masuk |
| Uang Ujian | Biaya ujian |
| Uang Kegiatan | Kegiatan sekolah |
| Seragam | Pembelian seragam |
| Buku | Pembelian buku |

### Cicilan
Satu tagihan bisa dibayar beberapa kali:
```
Tagihan: SPP Januari = Rp 500.000
  - Cicilan 1: Rp 200.000 (verified)
  - Cicilan 2: Rp 200.000 (verified)
  - Cicilan 3: Rp 100.000 (pending)
  - Status: sebagian (400.000 / 500.000)
```

---

## PPDB

### Fitur Modular

| Modul | Status | Keterangan |
|-------|--------|------------|
| Pendaftaran | Wajib | Form online, upload berkas |
| Verifikasi | Wajib | Admin cek kelengkapan |
| Ujian | Opsional | Input nilai ujian |
| Perankingan | Opsional | Hitung ranking |
| Pengumuman | Wajib | Status diterima/ditolak |
| Daftar Ulang | Opsional | Konfirmasi diterima |

### Alur PPDB
```
1. Pendaftar mengisi form online (publik)
2. Pendaftar upload berkas (publik)
3. Admin verifikasi berkas
4. [Opsional] Admin input nilai ujian
5. Admin jalankan ranking (POST /ppdb/ranking/run)
   - dry_run=true untuk preview tanpa menyimpan
   - dry_run=false untuk menyimpan hasil
6. Admin publish pengumuman (POST /ppdb/ranking/publish)
7. Pendaftar cek pengumuman (publik)
8. Pendaftar diterima → admin konfirmasi daftar ulang
```

### Perankingan

Modul perankingan bisa dikonfigurasi per sekolah per tahun ajaran. Admin memilih metode ranking dan mengatur kuota.

#### Metode Ranking

| Metode | Keterangan | Cocok Untuk |
|--------|------------|-------------|
| `nilai_ujian` | Ranking berdasarkan rata-rata nilai ujian | SMP, SMA, SMK |
| `umur` | Ranking berdasarkan umur (lebih muda = skor lebih tinggi) | SD |
| `jarak` | Ranking berdasarkan jarak rumah ke sekolah | SD, SMP (Permendikbud) |
| `kombinasi` | Bobot nilai ujian + umur + jarak | SMP, SMA |

#### Konfigurasi Bobot (Metode Kombinasi)

```json
{
    "nilai_ujian": 60,
    "umur": 20,
    "jarak": 20
}
```

Bobot diisi di field `bobot_json` pada konfigurasi ranking. Nilai menentukan proporsi kontribusi masing-masing faktor terhadap skor akhir.

#### Alur Eksekusi Ranking
```
1. Admin POST /ppdb/ranking/run dengan { tahun_ajaran_id, dry_run: true }
2. Sistem hitung skor untuk semua pendaftar tahun ajaran tersebut
3. Sistem sort berdasarkan skor DESC, lalu tanggal_lahir ASC (deterministik tiebreak)
4. Sistem assign ranking 1, 2, 3, ...
5. Sistem assign status berdasarkan kuota:
   - ranking <= kuota → diterima
   - ranking <= kuota + cadangan → cadangan
   - sisanya → tidak_diterima
6. Jika dry_run=true → return preview, tidak ada perubahan di DB
7. Jika dry_run=false → reset ranking sebelumnya, simpan skor/ranking/status baru
8. Simpan log ke ppdb_ranking_log untuk audit
```

#### Idempotency
Re-ranking menggantikan hasil sebelumnya secara atomik. Ranking sebelumnya di-reset sebelum hasil baru diterapkan. Setiap eksekusi di-log ke `ppdb_ranking_log`.

#### Publish Pengumuman
Setelah ranking dijalankan, admin bisa publish pengumuman (POST /ppdb/ranking/publish). Ini membuat record `ppdb_pengumuman` untuk setiap pendaftar yang memiliki ranking. Pengumuman dapat dilihat oleh pendaftar melalui halaman publik.

### Daftar Ulang

Setelah pendaftar dinyatakan diterima, admin dapat mengkonfirmasi daftar ulang:

```
1. Pendaftar melihat pengumuman "Diterima"
2. Admin membuka detail pendaftar
3. Admin klik "Konfirmasi Daftar Ulang"
4. Status berubah menjadi "daftar_ulang"
5. daftar_ulang_status = "sudah", daftar_ulang_at diisi
```

**Status Daftar Ulang:**
| Status | Keterangan |
|--------|------------|
| `belum` | Belum konfirmasi |
| `sudah` | Sudah konfirmasi |
| `batal` | Pembatalan |

Hanya pendaftar dengan status `diterima` yang dapat daftar ulang. Admin dapat memfilter daftar pendaftar berdasarkan status daftar ulang.

### Form Pendaftaran
- Nama lengkap
- NIK
- Tempat, tanggal lahir
- Jenis kelamin
- Agama
- Alamat
- Asal sekolah
- No HP, email
- Data orangtua (nama, no HP, pekerjaan)
- Foto

### Berkas yang Diupload
| Jenis | Format | Wajib |
|-------|--------|-------|
| Ijazah | PDF/JPG | Ya |
| SKHUN | PDF/JPG | Ya |
| Akta Lahir | PDF/JPG | Ya |
| Kartu Keluarga | PDF/JPG | Ya |
| Foto | JPG/PNG | Ya |

### Status Pendaftar
```
menunggu → berkas_lengkap → diterima
         → berkas_ditolak
         → cadangan → diterima (jika kuota terbuka)
                    → tidak_diterima (jika periode habis)

menunggu → diterima (tanpa verifikasi berkas)
menunggu → tidak_diterima
menunggu → cadangan
diterima → daftar_ulang
```

---

## Notifikasi

### Fitur
- WhatsApp (whatsmeow)
- Telegram (Bot API)
- Email (SMTP)
- Queue dengan retry
- Konfigurasi per sekolah

### Adapter Pattern
```go
type Notifier interface {
    Send(ctx context.Context, recipient string, message string) error
}

// implementasi
- WhatsAppNotifier (whatsmeow)
- TelegramNotifier
- EmailNotifier (SMTP)
```

### Queue
Notifikasi dikirim via queue untuk reliability:
```
1. Buat notifikasi → simpan ke notifikasi_antrian (status: pending)
2. Worker proses queue
3. Kirim via adapter yang aktif
4. Update status (sent / failed)
5. Kalau failed → retry (max 3x)
6. Kalau tetap failed → log error, admin bisa manual retry
```

### Trigger Notifikasi
| Event | Notifikasi |
|-------|------------|
| Tagihan baru | Siswa/Orangtua |
| Pembayaran verified | Siswa/Orangtua |
| Pembayaran rejected | Siswa/Orangtua |
| Jatuh tempo besok | Siswa/Orangtua |
| PPDB: berkas diterima | Pendaftar |
| PPDB: berkas ditolak | Pendaftar |
| PPDB: pengumuman | Pendaftar |

### Template Notifikasi
```
Tagihan Baru:
"Tagihan {kategori} sebesar Rp {nominal} untuk {nama_siswa}. Jatuh tempo: {jatuh_tempo}. Silakan bayar melalui aplikasi."

Pembayaran Verified:
"Pembayaran {kategori} sebesar Rp {jumlah} untuk {nama_siswa} telah diverifikasi. Terima kasih."

Pengumuman PPDB:
"Hasil PPDB {nama_sekolah}: {nama_pendaftar} {status}. Silakan cek aplikasi untuk detail."
```

---

## Laporan

### Fitur
- Dashboard per role
- Export Excel
- Cetak kwitansi
- Rekap pembayaran
- Rekap PPDB
- Rekap siswa

### Dashboard Admin
- Total siswa aktif
- Total pembayaran bulan ini
- Pendaftar PPDB baru
- Tagihan jatuh tempo
- Grafik pembayaran per bulan

### Dashboard Operator Keuangan
- Tagihan belum bayar
- Pembayaran pending verifikasi
- Pembayaran hari ini
- Tagihan jatuh tempo

### Dashboard Operator PPDB
- Pendaftar baru hari ini
- Berkas pending verifikasi
- Total pendaftar
- Status pengumuman

### Dashboard Guru
- Kelas yang diajar
- Jumlah siswa per kelas

### Dashboard Siswa/Orangtua
- Tagihan belum bayar
- Riwayat pembayaran
- Status PPDB

### Export Excel
Semua laporan bisa di-export ke Excel:
- Rekap pembayaran per periode
- Rekap siswa per kelas
- Rekap PPDB
- Daftar tagihan

### Cetak
- Kwitansi pembayaran
- Surat keterangan
- Daftar siswa

---

## Upload & File Management

### Fitur
- Upload file (JPEG, PNG, PDF, max 5 MB)
- Kategori: bukti_bayar, berkas_ppdb, foto_siswa, foto_ppdb, general
- Download via auth handler (tenant-scoped)
- Signed URL untuk akses sementara tanpa auth

### Struktur File
```
./uploads/
  {sekolah_id}/
    bukti_bayar/
    berkas_ppdb/
    foto_siswa/
    foto_ppdb/
    general/
```

### Signed URL
Untuk keperluan notifikasi (WhatsApp/Telegram/Email) atau sharing sementara:

```go
// Generate signed URL (dari backend/internal code)
result, err := upload.SignPath(secret, sekolahID, filePath, 5*time.Minute)
// result.URL → "/api/v1/upload/signed/eyJzaWQi..."

// Atau via API
POST /api/v1/upload/signed
{ "path": "1/bukti_bayar/abc.pdf", "ttl_seconds": 300 }
```

- Default TTL: 5 menit
- Max TTL: 15 menit
- Token: HMAC-SHA256 signed, tamper-proof
- Token reusable sampai expired (tidak single-use)

---

## Backup

### Fitur
- Auto backup harian
- Backup manual dari UI
- Download backup
- Restore backup

### Struktur Backup
```
/backups
  /2024-01-15_020000
    sekolah.db
    uploads.tar.gz
  /2024-01-16_020000
    sekolah.db
    uploads.tar.gz
```

### Konfigurasi Backup
```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"    # jam 2 pagi
  retention: 7             # simpan 7 hari terakhir
  path: "./backups"
```
