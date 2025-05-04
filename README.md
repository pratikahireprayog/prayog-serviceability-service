# Prayog Serviceability Service

A service to determine if a customer's location is serviceable for delivery.

## What is this repository for?

* Providing fast, accurate serviceability checks for delivery to customer locations
* Integration with logistics systems and pin code databases
* Supporting different delivery timeframes (standard, express, same-day)

## How do I get set up?

### Prerequisites
* Go 1.24.2 or later
* PostgreSQL 14 or later (for development)
* Redis (optional, for caching)

### Installing Go 1.24.2
This project requires Go 1.24.2. Here are several ways to install it:

#### Using the official installer
1. Download Go 1.24.2 from [golang.org/dl](https://golang.org/dl/)
2. Follow the installation instructions for your operating system

#### Using a package manager
- **macOS with Homebrew**:
  ```bash
  brew install go@1.24
  ```
- **Linux with apt**:
  ```bash
  wget https://golang.org/dl/go1.24.2.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
  source ~/.profile
  ```

#### Using asdf version manager
We provide a `.tool-versions` file that specifies Go 1.24.2 for asdf users:
```bash
asdf plugin add golang
asdf install
```

### Setup Steps
1. Clone the repository
```bash
git clone https://github.com/prayog/serviceability.git
cd serviceability
```

2. Install dependencies
```bash
make deps
```

3. Configure the application
```bash
cp configs/example.env configs/.env
# Edit .env file with your configuration
```

4. Run the application
```bash
make run
```

### Available Make Commands
* `make build` - Build the application
* `make test` - Run tests
* `make run` - Run the application
* `make clean` - Clean build artifacts
* `make tidy` - Tidy go modules
* See `make help` for all available commands

## Task Management

This project uses Task Master AI for task management. Task Master helps break down complex requirements into manageable tasks and track their implementation.

For detailed instructions on setting up and using Task Master, see [README-task-master.md](README-task-master.md).

## Contribution guidelines

* Follow Go coding standards and project structure
* Write tests for new features
* Update documentation

## Who do I talk to?

* Repo owner or admin
* Other community or team contact

## Database Setup

### Create a PostgreSQL Database
```bash
# Create the database
psql -U postgres -c "CREATE DATABASE \"serviceability-dev\";"

# Run migrations
migrate -path cmd/migrations/migrations -database "postgresql://postgres:postgres@localhost:5432/serviceability-dev?sslmode=disable" up
```

The migration will create and populate the following tables:
- Location entities: countries, administrative_regions, cities, areas, postal_codes
- Location aliases: country_aliases, administrative_region_aliases, city_aliases, area_aliases
- Service entities: order_types, service_types, service_availabilities

### Database Schema
The schema documentation is available in the [docs/db-schema.md](docs/db-schema.md) file.

## Environment Setup
Copy the `.env-local/dev.env` file to `.env` in the project root:
```bash
cp .env-local/dev.env .env
```

## Run the Application
```bash
go run cmd/serviceability/main.go
```