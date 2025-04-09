# Auth Handlers

## Register (POST)
### Принимает:
```json
{
  "email": "string",
  "password": "string"
}
```
### Возвращает:
- 201 Created при успешной регистрации
- 400 Bad Request если данные некорректны
```json
{
  "message": "User registered successfully",
  "user_id": "uuid"
}
```
- 409 Conflict если пользователь уже существует
```json
{
  "error": "User already exists"
}
```