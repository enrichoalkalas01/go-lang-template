start:
	@go run cmd/web/server.go

start-gin:
	@go run cmd/web/gin/server.go

start-fiber:
	@go run cmd/web/fiber/server.go

start-echo:
	@go run cmd/web/echo/server.go

start-chi:
	@go run cmd/web/chi/server.go