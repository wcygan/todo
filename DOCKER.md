# Docker

[Frontend Dockerfile](frontend/Dockerfile)
[Backend Dockerfile](backend/Dockerfile)

```bash
docker build -f backend/Dockerfile -t todo-backend .
docker build -f frontend/Dockerfile -t todo-frontend .
docker run -p 8080:8080 todo-backend
docker run -p 3000:3000 todo-frontend
```