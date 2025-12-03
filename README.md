# Auto Sender – Golang | MySQL | Redis | Webhook.site  
Automated Message Dispatching System (with Scheduler + DB + Cache + API)

This project implements an **automatic message sending system** in Golang.  
It retrieves **2 pending messages every 2 minutes**, sends them to a webhook endpoint,  
marks them as **sent**, and **caches the messageId** in Redis.

Designed per assignment requirements:

- Clean architecture & folder structure  
- Domain-driven packages  
- MySQL for persistence  
- Redis for bonus cache  
- Background scheduler (no cron packages)  
- 2 API endpoints (start/stop scheduler + list sent messages)  
- Docker-ready deployment  
- Swagger support  

---

## Features

### Automatic message processing
- Every 2 minutes (configurable)
- Sends **2 pending messages** at a time
- Messages sent **only once**
- Newly inserted messages processed automatically

### Webhook integration
- Sends messages to any webhook endpoint (e.g., webhook.site)
- Generates **unique messageId** using UUID
- Stores the messageId in DB and Redis (bonus)

### Storage
- **MySQL** – primary message store  
- **Redis** – stores messageId + timestamp (optional bonus)

### Proper architecture
- Http layer  
- Scheduler  
- Domain services  
- Repositories  
- Config  
- Redis client  
- Clean code with interfaces  

---

### How to Run
**1. To build and run the application using Docker**

- docker compose up --build
- This will start all required services: auto-sender-app (Go application) , auto-sender-db (MySQL), auto-sender-redis (Redis)
- Once containers are up, you can access the APIs and Swagger documentation as below.

**2. API Documentation (Swagger)**

- Swagger UI is served directly by the application. Access it at: http://localhost:8080/swagger/index.html

**3. API Endpoints**

- Start/Stop scheduler  
- POST /scheduler  { "action": "start" | "stop" }

- List sent messages
- GET /messages/sent?limit=5

**4 Checking Redis Stored Message Data**

- The application stores message metadata (messageId, sentAt) in Redis using HASH format.
- To check values: Enter Redis container: docker exec -it auto-sender-redis redis-cli
- Run: HGETALL message:4 OR get a single field: HGET message:4 messageId


