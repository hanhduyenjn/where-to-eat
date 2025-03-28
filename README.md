# Project Setup Guide

This guide will walk you through the process of setting up the project with Docker Compose, running various jobs, and accessing the API.

## 1. Docker Compose Setup

### Prerequisites:
- Ensure Docker and Docker Compose are installed on your system.

### Steps:
1. **Start PostgreSQL and MongoDB**:
   - In the root directory of the project, run the following command to start the PostgreSQL and MongoDB containers:

     ```bash
     docker-compose up
     ```

   This will start both databases with the configurations defined in the `docker-compose.yml` file.

2. **Import Category Configuration into MongoDB**:
   - The `category_config.json` file should be imported into MongoDB's collection `<MONGO_COLLECTION_CONFIG>`. You can do this by running the following MongoDB command or using a script.

## 2. Environment Variables

Make sure to set the following environment variables:

```bash
GOOGLE_API_KEY=<your-google-api-key>

MONGO_DB=<your-mongo-database-name>
MONGO_USER=<your-mongo-username>
MONGO_PASSWORD=<your-mongo-password>
MONGO_HOST=<your-mongo-host>
MONGO_PORT=<your-mongo-port>
MONGO_AUTH_SOURCE=<your-mongo-auth-source>

MONGO_COLLECTION_AREAS=<your-mongo-areas-collection-name>
MONGO_COLLECTION_SEARCH_RESULTS=<your-mongo-search-results-collection-name>
MONGO_COLLECTION_CONFIG=<your-mongo-config-collection-name>

POSTGRES_DB=<your-postgres-database-name>
POSTGRES_USER=<your-postgres-username>
POSTGRES_PASSWORD=<your-postgres-password>
POSTGRES_HOST=<your-postgres-host>
POSTGRES_PORT=<your-postgres-port>
POSTGRES_SSL=<your-postgres-ssl-mode>
```

## 3. Running the Jobs
### 3.1. Fetch Images (Crawling)
To fetch images, run the following command:

```bash
go run ./cmd/job/main.go run-fetch-images <limit> <offset>
```
Example:

```bash
go run ./cmd/job/main.go run-fetch-images 1000 0
```

### 3.2. Fetch Areas (One-time API Call)
To fetch areas for a given location (e.g., "Quận 11"), run:

```bash
go run ./cmd/job/main.go run-fetch-areas "<area-name>"
```

Example:

```bash
go run ./cmd/job/main.go run-fetch-areas "Quận 11"
```

### 3.3. Fetch Places (Multiple API Calls)
To fetch places in a given area (latitude/longitude coordinates), run:

```bash
go run ./cmd/job/main.go run-fetch-places <minLat> <maxLat> <minLng> <maxLng> <category>
```

Example:

```bash
go run ./cmd/job/main.go run-fetch-places 11.753392196944883 10.77841976253177 10.63416287050883 106.6615346564114 restaurants
```

Note: The available categories are stored in the `category_config` table in MongoDB.

### 3.4. ETL (Transform Data)
To transform the fetched data, run the following command:

```bash
go run ./cmd/pipeline/main.go
```

## 4. Running the Server
To start the server, run:

```bash
go run ./cmd/server/main.go
```

The server will start on [http://localhost:8081](http://localhost:8081).

## 5. API Usage
Once the server is running, you can test the API by sending a GET request to the following endpoint:

Example cURL Command:

```bash
curl --location 'http://localhost:8081/nearby-places?lat=10.770413699999999&lng=106.6699414&radius=2000&searchString=Ti%E1%BB%87m%20B%C3%A1nh%Kem' \
--header 'Content-Type: application/json' \
--header 'X-Goog-Api-Key: <your-google-api-key>' \
--header 'X-Goog-FieldMask: *'
```

Parameters:
- `lat`: Latitude of the location.
- `lng`: Longitude of the location.
- `radius`: Search radius (in meters).
- `searchString`: Search query (e.g., place or business name).