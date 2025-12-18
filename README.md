# GophKeeper — Password and Private Data Manager

Educational client–server application for secure storage of users private data (passwords, text notes, bank cards, and files).

## Project Status

![Go](https://img.shields.io/badge/go-1.22+-blue)
![gRPC](https://img.shields.io/badge/gRPC-enabled-green)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-supported-blue)

---

## Architecture

The project follows a **client–server** architecture.

### Server

- implemented in Go
- communicates with clients via **gRPC**
- stores user data in **PostgreSQL**
- responsible for:
  - user registration
  - authentication
  - secure storage and management of user data
  - JWT validation and lifecycle management

### Client

- CLI application implemented using **cobra**
- communicates with the server via gRPC
- stores the JWT token locally in the user’s working directory
- provides commands for authentication and vault operations

---

## Security

### Authentication and Authorization

- **JWT** is used for authentication
- the token:
  - is issued by the server after successful registration or login
  - is sent with every gRPC request via metadata: `Authorization: Bearer <token>`
  - is validated by the server for integrity and expiration
- when an `Unauthenticated` response is received, the client automatically removes the locally stored token

### Data Encryption

- the server uses a **master key** to protect user encryption keys
- user data is stored in encrypted form
- the master key and JWT signing key **must not be committed** to the repository and are provided via configuration

---

## Supported Data Types

| Type        | Description                     |
|-------------|---------------------------------|
| `PASSWORD`  | Passwords (login, secret, meta) |
| `TEXT`      | Text notes                      |
| `BANK_CARD` | Bank card data                  |
| `FILE`      | Arbitrary files                 |

---

## Configuration

### Server Configuration

The server configuration is stored in `server.yaml`.

```yaml
server:
  address: ":8080"
  master_key: "CHANGE_ME_MASTER_KEY"
  token_key: "CHANGE_ME_TOKEN_KEY"
  token_lifetime: 180
  use_tls: false

database:
  host: "localhost"
  port: "5432"
  name: "passkeeper"
  user: "postgres"
  password: "CHANGE_ME_DB_PASS"

logging:
  level: "debug"
```

### Server Configuration Parameters

- `server.address` — server bind address
- `server.master_key` — master key used to encrypt user data
- `server.token_key` — key used to sign JWTs
- `server.token_lifetime` — JWT lifetime in minutes
- `server.use_tls` — enable or disable TLS
- `database.*` — PostgreSQL connection settings
- `logging.level` — logging level

---

### Client Configuration

The client configuration is stored in `client.yaml`.

```yaml
server_address: ":8080"
work_dir: "CHANGE_ME_WORK_DIR"
use_tls: false
```

### Client Configuration Parameters

- `server_address` — gRPC server address
- `work_dir` — client working directory:
  - stores the JWT token (`token.jwt`)
  - used to save downloaded files
- `use_tls` — enable or disable TLS

---

## CLI Commands

The CLI is divided into two top-level command groups:

- `user` — authentication and registration
- `vault` — operations on stored data

---

## `user` Commands

### Register a new user

```sh
gophkeeper user register -l "login" -p "password"
```

### Login

```sh
gophkeeper user login -l "login" -p "password"
```

---

## `vault` Commands

### Get all items of a specific type

```sh
gophkeeper vault getall -t PASSWORD
```

### Get an item by ID

```sh
gophkeeper vault get --id 00c15ce5-b86d-47ce-8298-710d875acbfd
```

### Delete an item by ID

```sh
gophkeeper vault delete --id 00c15ce5-b86d-47ce-8298-710d875acbfd
```

### Add a password

```sh
gophkeeper vault add password \
  -p "password" \
  -l "login" \
  -r "Resource name" \
  -c "Comment"
```

### Add a text note

```sh
gophkeeper vault add text \
  -t "Text" \
  -n "Note title" \
  -c "Comment"
```

### Add a bank card

```sh
gophkeeper vault add bcard \
  -o "Card holder" \
  -n "0000 0000 0000 0000" \
  -s "CSV" \
  -m 2 \
  -y 25 \
  -b "Bank" \
  -c "Comment"
```

### Add a file

```sh
gophkeeper vault add file \
  -p "file.text" \
  -c "Comment"
```

---

## Project Purpose

This project is **educational** and is intended for:

- practicing client–server interaction in Go
- working with gRPC
- implementing JWT-based authentication
- designing CLI applications
- learning secure data storage principles