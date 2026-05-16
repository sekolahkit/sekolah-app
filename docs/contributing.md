# Contributing

## Overview

Terima kasih atas ketertarikan Anda untuk berkontribusi ke SekolahApp! Panduan ini akan membantu Anda memulai.

---

## Getting Started

### 1. Fork & Clone

```bash
# Fork repository di GitHub
# Clone hasil fork
git clone https://github.com/YOUR_USERNAME/sekolah-app.git
cd sekolah-app

# Tambah upstream
git remote add upstream https://github.com/Sekolahkit/sekolah-app.git
```

### 2. Setup Development

```bash
# Backend
cd backend
go mod tidy

# Frontend
cd ../frontend
npm install
```

### 3. Buat Branch

```bash
# Sync dengan upstream
git fetch upstream
git checkout -b feature/my-feature upstream/main
```

---

## Development Workflow

### Branch Naming

| Prefix | Keterangan |
|--------|------------|
| `feature/` | Fitur baru |
| `fix/` | Bug fix |
| `docs/` | Dokumentasi |
| `refactor/` | Refactoring |
| `test/` | Testing |

Contoh: `feature/payment-gateway`, `fix/login-error`

### Commit Message

Gunakan format Conventional Commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

| Type | Keterangan |
|------|------------|
| feat | Fitur baru |
| fix | Bug fix |
| docs | Dokumentasi |
| style | Formatting |
| refactor | Refactoring |
| test | Testing |
| chore | Maintenance |

Contoh:
```
feat(payment): add Midtrans integration

- Add Midtrans adapter
- Add callback handler
- Update documentation

Closes #123
```

### Code Style

#### Go

- Gunakan `gofmt` untuk formatting
- Gunakan `golangci-lint` untuk linting
- Ikuti idiomatic Go

```bash
# Format
gofmt -w .

# Lint
golangci-lint run
```

#### TypeScript/React

- Gunakan ESLint + Prettier
- Ikuti style guide yang sudah ada

```bash
# Lint
npm run lint

# Format
npm run format
```

### Testing

#### Backend

```bash
cd backend

# Run semua test
go test ./...

# Run test dengan coverage
go test -cover ./...

# Run test tertentu
go test ./internal/siswa/...
```

#### Frontend

```bash
cd frontend

# Run semua test
npm test

# Run test dengan coverage
npm run test:coverage
```

---

## Pull Request

### Sebelum Submit

1. Pastikan semua test pass
2. Pastikan lint bersih
3. Update dokumentasi jika perlu
4. Test manual di browser

### Template PR

```markdown
## Deskripsi
Jelaskan perubahan yang dibuat

## Tipe Perubahan
- [ ] Fitur baru
- [ ] Bug fix
- [ ] Dokumentasi
- [ ] Refactoring

## Checklist
- [ ] Test pass
- [ ] Lint bersih
- [ ] Dokumentasi diupdate
- [ ] Tested di browser

## Screenshot (jika UI berubah)
[screenshot]

## Related Issues
Closes #123
```

### Review Process

1. Submit PR ke `main` branch
2. Tunggu review dari maintainer
3. Fix feedback jika ada
4. PR akan di-merge setelah approve

---

## Reporting Issues

### Bug Report

```markdown
## Deskripsi
Jelaskan bug yang terjadi

## Steps to Reproduce
1. Buka halaman ...
2. Klik ...
3. Lihat error

## Expected Behavior
Seharusnya ...

## Actual Behavior
Yang terjadi ...

## Environment
- OS: [e.g., Ubuntu 22.04]
- Browser: [e.g., Chrome 120]
- Version: [e.g., 1.0.0]

## Screenshot
[screenshot]
```

### Feature Request

```markdown
## Deskripsi
Jelaskan fitur yang diinginkan

## Use Case
Mengapa fitur ini dibutuhkan?

## Solusi yang Diusulkan
Bagaimana cara implementasinya?

## Alternatif
Alternatif lain yang dipertimbangkan
```

---

## Architecture

### Backend Structure

```
/internal
  /modul
    handler.go    → HTTP handler
    service.go    → Business logic
    repository.go → Database access
```

### Adding New Module

1. Buat folder di `/internal`
2. Buat `handler.go`, `service.go`, `repository.go`
3. Register route di `cmd/server/main.go`
4. Buat migration jika perlu tabel baru
5. Tambah dokumentasi

### Database Migration

```bash
# Buat migration baru
migrate create -ext sql -dir backend/internal/migration/migrations -seq add_new_table

# Akan menghasilkan:
# 000005_add_new_table.up.sql
# 000005_add_new_table.down.sql
```

---

## Community

- **GitHub Issues** — Bug reports & feature requests
- **GitHub Discussions** — Q&A dan diskusi

---

## License

Dengan berkontribusi, Anda setuju bahwa kontribusi Anda akan dilisensikan di bawah MIT License.
