# 🚚 FleetCommander

**FleetCommander** is a high-performance, distributed fleet management system. It simulates real-time vehicle tracking using Event Streaming, validates business transactions via Message Queues, and visualizes data on a Reactive Frontend.

![Architecture Diagram](diagram.png)

## 🌟 Key Features

* **Real-Time Event Streaming:** High-throughput GPS telemetry processing using **Redpanda (Kafka API)**.
* **Reactive Dashboard:** **Vue 3** frontend with **Leaflet** maps and **Socket.io** for live truck updates without polling.
* **Microservices Architecture:**
    * **Go (Golang):** High-performance services for Auth and Order management.
    * **Python:** Data simulation and FaaS (Function-as-a-Service) trigger for speeding detection.
    * **Node.js:** WebSocket relay service.
* **Reliable Messaging:** **RabbitMQ** ensures critical business orders are never lost.
* **Security:** **JWT** (JSON Web Token) authentication protected by an API Gateway (**Traefik**).
* **Scalability:** Services are containerized and ready for horizontal scaling.

---

## 📋 Prerequisites

Ensure you have the following installed on your machine:
* [Docker & Docker Compose](https://www.docker.com/) (Daemon must be running)
* [Go](https://go.dev/) (v1.20+)
* [Node.js](https://nodejs.org/) (v18+)
* [Python](https://www.python.org/) (v3.10+)

---

## 🛠️ Infrastructure Setup

This project uses Docker to host the data layer (Redpanda, RabbitMQ, Redis, Traefik).

1.  **Clone the repository:**
    ```bash
    git clone [https://github.com/YOUR_USERNAME/fleet-commander.git](https://github.com/YOUR_USERNAME/fleet-commander.git)
    cd fleet-commander
    ```

2.  **Start the containers:**
    ```bash
    docker compose up -d
    ```

3.  **Initialize Kafka Topics (One-time setup):**
    Redpanda auto-creation works, but manual creation prevents startup race conditions.
    ```bash
    docker exec fleetcommander-redpanda-1 rpk topic create vehicle-telemetry
    docker exec fleetcommander-redpanda-1 rpk topic create system-alerts
    ```

4.  **Verify Gateways:**
    * **Traefik Dashboard:** [http://localhost:8090](http://localhost:8090)
    * **RabbitMQ Dashboard:** [http://localhost:8090](http://localhost:8090) (via Traefik routing)

---

## 🚀 How to Run the Services

To simulate a full distributed environment, open **4 separate terminal tabs**.

### Terminal 1: Backend Services (Go)
This runs the Authentication and Order services.

```bash
# 1. Start Auth Service (Port 3002)
cd services/auth-service
go run main.go

# (Open a split pane or background the process to run the second one)

# 2. Start Order Service (Port 3001)
cd services/order-service
go run main.go
```

### Terminal 2: WebSocket Relay (Node.js)
This consumes the Kafka stream and broadcasts to the frontend.

```bash
cd services/websocket-service
npm install
node server.js
```

### Terminal 3: The Frontend (Vue.js)
This serves the interactive dashboard.

```bash
cd frontend/shell-app
npm install
npm run dev
```
> Open the link provided (usually [http://localhost:5173](http://localhost:5173)).

### Terminal 4: The Fleet Simulation (Python)
This generates the data. You can run this command in multiple tabs to create **multiple trucks**.

```bash
cd services/telemetry-service
# Create/Activate Virtual Env
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt  # or pip install aiokafka

# Start a truck
python3 main.py
```

*(Optional) Start the Speeding Detector FaaS:*
```bash
cd services/speed-function
python3 main.py
```

---

## 🧪 Demo Scenarios

Once everything is running, try these actions:

1.  **Live Tracking:**
    * Look at the map. You should see a truck icon moving.
    * Click the icon to see real-time speed/fuel stats in the sidebar.
    * Start a second python script to see a second truck appear instantly.

2.  **Secure Dispatch (JWT Flow):**
    * In the sidebar, find the **"Dispatch Command"** panel.
    * Log in with default credentials:
        * **User:** `admin`
        * **Pass:** `password123`
    * The panel will unlock. Enter a destination (e.g., "Cluj") and click **DISPATCH**.
    * Check **Terminal 1** (Go): You will see `📦 Order Received & Queued`.

3.  **Speeding Alert (FaaS):**
    * Watch the dashboard. If the truck simulation hits > 100km/h, a red alert will appear in the "Event Log" and the speed indicator will flash red.

---

## 📂 Project Structure

```
fleet-commander/
├── docker-compose.yml       # Infrastructure (Redpanda, Redis, RabbitMQ, Traefik)
├── README.md                # Documentation
├── frontend/
│   └── shell-app/           # Vue.js Dashboard (Leaflet Map + WebSockets)
│       ├── src/
│       │   ├── App.vue      # Main UI Logic
│       │   └── main.js      # Entry point
│       └── package.json
└── services/
    ├── auth-service/        # Go: JWT Authentication
    │   └── main.go
    ├── order-service/       # Go: Secure API -> RabbitMQ
    │   └── main.go
    ├── telemetry-service/   # Python: Data Producer (Truck Simulator)
    │   └── main.py
    ├── speed-function/      # Python: FaaS Trigger (Speed Detector)
    │   ├── handler.py       # Pure Business Logic
    │   └── main.py          # Trigger Logic
    └── websocket-service/   # Node.js: Kafka Consumer -> Socket.io Broadcast
        └── server.js
```
