.PHONY: build dev clean test frontend backend

build: frontend backend

frontend:
	cd ../frontend && npm run build
	rm -rf pkg/frontend/dist
	cp -r ../frontend/dist pkg/frontend/dist

backend:
	go build -o sekolah-app cmd/server/main.go

dev:
	go run cmd/server/main.go

test:
	go test ./... -count=1

clean:
	rm -f sekolah-app
	rm -rf pkg/frontend/dist
	mkdir -p pkg/frontend/dist
	echo '<!DOCTYPE html><html><body>build frontend first</body></html>' > pkg/frontend/dist/placeholder.html
