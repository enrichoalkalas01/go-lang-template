```

## 3. Perbandingan Output Logs

### Echo Output:
```

2025-10-30T21:43:03.240+0700 INFO Starting Echo application...
2025-10-30T21:43:03.241+0700 INFO Configuration loaded successfully
2025-10-30T21:43:03.242+0700 INFO Starting Echo server

---

/ **/\_**/ / **_
/ _// **/ _ \/ _ \
/**\_/\__/_//\_/\_**/ v4.x.x
High performance, minimalist Go web framework
https://echo.labstack.com
****************\_\_\_\_****************O/**\_\_\_**
O\
⇨ http server started on 0.0.0.0:8080

2025-10-30T21:43:03.243+0700 INFO Echo server started successfully

```

### Fiber Output:
```

2025-10-30T21:43:03.240+0700 INFO Starting Fiber application...
2025-10-30T21:43:03.241+0700 INFO Configuration loaded successfully
2025-10-30T21:43:03.242+0700 INFO Starting Fiber server
⚡ Fiber server starting on http://0.0.0.0:8081

┌───────────────────────────────────────────────────┐
│ Fiber v2.x.x │
│ http://0.0.0.0:8081 │
│ │
│ Handlers ............. 15 Processes ........... 1 │
│ Prefork ....... Disabled PID ............. 12345 │
└───────────────────────────────────────────────────┘

2025-10-30T21:43:03.243+0700 INFO Fiber server started successfully

```

### Gin Output:
```

2025-10-30T21:43:03.240+0700 INFO Starting Gin application...
2025-10-30T21:43:03.241+0700 INFO Configuration loaded successfully
2025-10-30T21:43:03.242+0700 INFO Starting Gin server
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.

-   using env: export GIN_MODE=release
-   using code: gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET /health --> service-golang/pkg/server/gin.(*GinServer).healthCheck-fm (5 handlers)
[GIN-debug] GET /api/v1/users --> service-golang/pkg/server/gin.(*GinServer).getUsers-fm (5 handlers)
[GIN-debug] GET /api/v1/users/:id --> service-golang/pkg/server/gin.(*GinServer).getUserByID-fm (5 handlers)
[GIN-debug] POST /api/v1/users --> service-golang/pkg/server/gin.(*GinServer).createUser-fm (5 handlers)
[GIN-debug] PUT /api/v1/users/:id --> service-golang/pkg/server/gin.(*GinServer).updateUser-fm (5 handlers)
[GIN-debug] DELETE /api/v1/users/:id --> service-golang/pkg/server/gin.(*GinServer).deleteUser-fm (5 handlers)
[GIN-debug] Listening and serving HTTP on 0.0.0.0:8082
2025-10-30T21:43:03.243+0700 INFO Gin server started successfully
