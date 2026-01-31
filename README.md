# Gator

Gator is a command-line RSS feed aggregator that allows you to follow and read RSS feeds from the terminal. It stores feeds and posts in a PostgreSQL database, allowing you to track multiple feeds, follow/unfollow them, and browse posts from your followed feeds.

## Prerequisites

Before using Gator, ensure you have the following installed:

- **Go** (version 1.25.4 or higher) - [Download here](https://go.dev/dl/)
- **PostgreSQL** - [Download here](https://www.postgresql.org/download/)

## Installation

Install Gator using `go install`:

```bash
go install github.com/aHardReset/gator@latest
```

This will install the `gator` binary to your `$GOPATH/bin` directory. Make sure this directory is in your system's PATH.

## Configuration

### 1. Set up PostgreSQL Database

Create a PostgreSQL database for Gator:

```bash
createdb gator
```

### 2. Run Database Migrations

Navigate to the project directory and run the SQL schema files located in `sql/schema/` to set up your database tables.

### 3. Create Configuration File

Gator requires a configuration file at `~/.gatorconfig.json`. Create this file with the following structure:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `username` and `password` with your PostgreSQL credentials. The `current_user_name` field will be automatically updated when you log in.

## Usage

### User Management

**Register a new user:**

```bash
gator register <username>
```

**Login as an existing user:**

```bash
gator login <username>
```

**List all users:**

```bash
gator users
```

Shows all registered users, with `(current)` indicator next to the logged-in user.

### Feed Management

**Add a new RSS feed:**

```bash
gator addfeed <feed_name> <feed_url>
```

Automatically follows the feed after creation.

**List all feeds:**

```bash
gator feeds
```

Shows all feeds with their names, URLs, and creators.

**Follow a feed:**

```bash
gator follow <feed_url>
```

**Unfollow a feed:**

```bash
gator unfollow <feed_url>
```

**List feeds you're following:**

```bash
gator following
```

### Reading Posts

**Browse posts from your followed feeds:**

```bash
gator browse [limit]
```

Display recent posts from feeds you follow. Defaults to 2 posts if no limit is specified.

Example:

```bash
gator browse 10
```

### Feed Aggregation

**Start the feed aggregator:**

```bash
gator agg <duration>
```

Continuously fetches and updates feeds at the specified interval.

Example:

```bash
gator agg 1m   # Fetch feeds every 1 minute
gator agg 30s  # Fetch feeds every 30 seconds
gator agg 1h   # Fetch feeds every 1 hour
```

Duration format follows Go's time.Duration syntax (e.g., `1m`, `30s`, `1h30m`).

### Database Management

**Reset the database:**

```bash
gator reset
```

� Warning: This flushes all tables and deletes all data.

## Example Workflow

```bash
# Register and login
gator register john
gator login john

# Add some feeds
gator addfeed TechNews https://example.com/tech/rss
gator addfeed DevBlog https://example.com/dev/feed.xml

# List your followed feeds
gator following

# Start aggregating feeds in the background
gator agg 5m &

# Browse recent posts
gator browse 5
```

## Project Structure

- `main.go` - Entry point and command registration
- `appLogic.go` - Command handler implementations
- `rss.go` - RSS feed fetching and parsing logic
- `middlewares.go` - Authentication middleware
- `internal/config/` - Configuration file management
- `internal/database/` - Database queries and models
- `sql/schema/` - Database schema migrations
- `sql/queries/` - SQL query definitions
