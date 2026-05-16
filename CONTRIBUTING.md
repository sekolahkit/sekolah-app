# Kontribusi

Terima kasih atas minat Anda untuk berkontribusi pada SekolahApp!

Panduan lengkap tersedia di [docs/contributing.md](docs/contributing.md), mencakup:

- Fork & clone workflow
- Branch naming convention
- Commit message format (Conventional Commits)
- Code style & linting
- Testing requirements
- Pull request template

## Quick Reference

```bash
# Fork dan clone
git clone https://github.com/<username>/sekolah-app.git
cd sekolah-app

# Buat branch
git checkout -b feat/nama-fitur

# Pastikan CI checks lolos sebelum PR
cd backend && go test ./... && go build ./...
cd frontend && npm run lint && npm run build

# Commit dengan conventional commits
git commit -m "feat(module): deskripsi singkat"
```

## Code of Conduct

Kami mengharapkan semua kontributor menjaga lingkungan yang ramah dan profesional.
