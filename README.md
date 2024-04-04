# ShortLink-go

![image](https://github.com/victortv7/shortlink-go/assets/9042203/e4ab591e-399c-452e-b26a-489a9c6af98d)

ShortLink-go is a Gin framework based URL shortening service, designed to be performant, horizontally scalable, and to work alongside PostgreSQL (or compatible databases e.g., CockroachDB, Aurora, Spanner) and Redis. See [How it Works](#how-it-works) for more details.

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Golang 1.22+

### Setup and Running Instructions

**Clone the Repository**

   ```bash
   git clone https://github.com/victortv7/shortlink-go.git
   cd shortlink-go
   ```

**Using Docker Compose**

   ```bash
   docker-compose up
   ```

**Using Makefile**

1. Start the PostgreSQL, run DB migrations, and start Redis:

   ```bash
   make db-up
   make db-migrate
   make redis-up
   ```

2. Configure the environment variables (see [.env.example](.env.example)) or use the default values.

3. Run the application:

   ```bash
   make run
   ```

### Testing

Run unit tests and generate a coverage report:

```bash
make test
make test-coverage
```

### Code Linting and Formatting

```bash
make lint
make fmt
```
## Usage

### Accessing the Swagger UI

```
http://localhost:8080/docs
```

### Creating a Short Link


```bash
curl -X 'POST' \
  'http://localhost:8080/create' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "long_url": "https://www.example.com"
}'
```

The response has the following format:

```json
{
   "short_link": "a4BhE"
}
```

### Redirecting a Short Link

To test the redirection functionality, simply navigate to the short link URL in your web browser or use a `curl` command like this:

```bash
curl -L 'http://localhost:8080/{short_link}'
```

### Accessing Link Stats

```bash
curl 'http://localhost:8080/stats/{short_link}'
```

## How It Works

ShortLink-go generates short links from long URLs and tracks their usage. Here's a brief overview of its core functionality:

- **Short Link Creation**: When a long URL is submitted, the application creates a new entry in the database with the URL (`long_url`) and an access count (`access_count`) set to zero. It then encodes the database entry's ID using [Base62](https://en.wikipedia.org/wiki/Base62) to generate a unique short link. This short link is also stored in Redis for quick access.

- **URL Redirection**: To redirect a short link to its original long URL, the application first checks Redis. If the short link is not found in Redis, it decodes the short link to retrieve the database ID, queries the database for the long URL, and updates Redis. This ensures subsequent accesses are faster. 

- **Access Count**: Each time a short link is accessed, its access count is incremented in the database to track how many times the short link has been used. This database write is done asynchronously in a background task to improve the latency of redirects.

### Cleaning Up

```bash
make db-down
make redis-down
make clean
```

## License

ShortLink-go is licensed under the MIT License. See [LICENSE](LICENSE) for details.
