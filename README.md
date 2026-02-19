# 🐙 Octo
**The Intelligent, Distributed HTTP Monitoring Platform**

![Octo Banner](assets/banner.png)

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-AGPLv3-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](Dockerfile)
[![Status](https://img.shields.io/badge/status-active-success.svg)](#)

Octo is a modern, high-performance monitoring solution designed for developers and SREs who need reliability without the bloat. Built with **Go**, **React**, and **TimescaleDB**, Octo delivers real-time insights, historical analysis, and seamless configuration management in a lightweight package.

---

## 🚀 Key Features

*   **⚡ High-Performance Monitoring**: Execute thousands of concurrent checks with minimal resource footprint.
*   **📊 Real-Time Dashboard**: Visualize uptime, latency, and health status instantly via a modern React UI.
*   **🔒 SSL/TLS Monitoring**: Automatically track certificate expiration and get alerted before they expire.
*   **💾 Long-Term Storage**: Leverage **TimescaleDB** for powerful time-series queries and historical data retention.
*   **⚙️ UI-Based Configuration**: manage your endpoints directly from the browser with a full-featured editor.
*   **🚨 Alerting System**: Flexible alerting via Webhooks (Slack, Discord, PagerDuty) with tag-based routing.
*   **🛠️ Hot-Reload**: Update your monitoring targets on the fly via API or UI—no restarts required.
*   **🐳 Container Native**: Deploy effortlessly with Docker or Podman.
*   **🌍 Multi-Satellite Monitoring**: Run distributed checks from multiple geographic locations to verify global availability and latency.

---

## 🌍 Multi-Satellite Monitoring (New!)

Octo now supports **Distributed Monitoring** via Satellites. Deploy lightweight satellite nodes in different regions (e.g., AWS us-east-1, DigitalOcean fra1) to:
*   **Verify Global Availability**: Ensure your service is accessible from around the world.
*   **Monitor Latency**: Track response times from different geographic edge locations.
*   **Geo-Redundancy**: Avoid false positives caused by local network issues.

### Satellite Configuration
Easily assign endpoints to specific satellites via the configuration UI or YAML.
![Satellite Config](/home/manu/.gemini/antigravity/brain/349300b6-f128-449e-ac40-a91e3509fed3/satellite_config.png)

### Global Health View
See the status of your endpoints from every active satellite directly on the dashboard.
![Dashboard Details](/home/manu/.gemini/antigravity/brain/349300b6-f128-449e-ac40-a91e3509fed3/dashboard_endpoint_details.png)

---

## 📸 Screenshots

### Operational Dashboard
Get a bird's-eye view of all your services with real-time status indicators.
![Dashboard](assets/dashboard.png)

### Deep Dive Analytics & SSL Info
Analyze historical performance trends, view SSL certificate details, and spot anomalies with synchronized charts.
![Details](/home/manu/.gemini/antigravity/brain/474e469c-9b7c-4452-bdbf-4167263b6816/endpoint_details_verification_1771347979164.png)

### Endpoint Management
Effortlessly add, edit, and remove endpoints using the built-in configuration editor.
![Configuration](/home/manu/.gemini/antigravity/brain/474e469c-9b7c-4452-bdbf-4167263b6816/endpoints_list_gear_icons_final_1771349462941.png)

### Powerful Editor
Configure advanced settings like Request Headers, Validation Rules, and SSL Alerts with ease.
![Editor](/home/manu/.gemini/antigravity/brain/474e469c-9b7c-4452-bdbf-4167263b6816/filled_endpoint_form_bottom_1771349992199.png)

---

## 🏗️ Architecture

Octo follows a clean, modular architecture separating the core monitoring engine from the storage and presentation layers.

```mermaid
graph TD
    User["User / Browser"] -->|"HTTP/WebSockets"| API["API Server (Go)"]
    API -->|"Serve UI"| UI["React Frontend"]
    API -->|"Read/Write"| Config["Config Manager"]
    Config -->|Load| SCH[Scheduler]
    SCH -->|Spawn| Workers["Worker Pool"]
    Workers -->|"HTTP Check"| Target["External Targets"]
    Workers -->|"Store Result"| DB[("TimescaleDB")]
    API -->|"Query Metrics"| DB
    
    subgraph Core Engine
        Config
        SCH
        Workers
    end

    subgraph Satellites ["Distributed Satellites"]
        Sat1["Satellite (US-East)"]
        Sat2["Satellite (EU-West)"]
    end

    Sat1 -->|"Poll Config / Push Results"| API
    Sat2 -->|"Poll Config / Push Results"| API
    Sat1 -->|"HTTP Check"| Target
    Sat2 -->|"HTTP Check"| Target
```

---

## ⚙️ Configuration
Octo uses a YAML configuration file located at `config/config.yml`.

### Volume Mounting
When running with Docker or Podman, the local `config/` directory is mounted into the container at `/config`.
This means you can edit `config/config.yml` on your host machine, and the changes will be reflected in the container (requires restart for some changes, though hot-reload is supported for endpoints).

**Note:** The `config.yml` file is NOT baked into the image. It is injected at runtime via the volume mount.

## 🗺️ Project Roadmap

| Version | Status | Features |
| :--- | :--- | :--- |
| **v0.1.0** | ✅ **Released** | MVP Core Engine, InfluxDB Support, Basic API |
| **v0.2.0** | ✅ **Released** | React Frontend, TimescaleDB Migration, Config UI, Historical Charts |
| **v0.3.0** | ✅ **Released** | Alerting Engine, SSL Monitoring, Config Editor, Aggregation Metrics |
| **v0.4.0** | 📅 **Planned** | User Authentication (OAuth2/OIDC), Multi-user Support |
| **v1.0.0** | 📅 **Future** | High Availability Clustering, Plugin System, Public Release |

---

## 🏁 Getting Started

### Prerequisites
*   **Docker** or **Podman**
*   **Go 1.22+** (only for manual builds)

### ⚡ Quick Start (Podman)
The fastest way to get up and running:

```bash
./scripts/run_podman.sh
```
Open **[http://localhost:8080](http://localhost:8080)** to start monitoring.

### 🐳 Quick Start (Docker Compose)
Standard deployment for Docker users:

```bash
docker-compose up -d --build
```

### 🔧 Manual Setup
For development or custom deployments:

1.  **Start Database**: Ensure TimescaleDB is running on `localhost:5432`.
2.  **Build & Run**:
    ```bash
    # Build Frontend
    cd web && npm install && npm run build && cd ..

    # Build Backend
    go build -o master cmd/master/main.go

    # Run
    ./master
    ```

---

## 📡 API Reference

Octo exposes a RESTful API for automation and integration.

*   `GET /health` - System health check
*   `GET /api/v1/config` - Retrieve current configuration
*   `POST /api/v1/config/endpoints` - Create new endpoint
*   `GET /api/v1/endpoints` - List all endpoints
*   `GET /api/v1/endpoints/{id}/history` - Retrieve historical metrics

---

<p align="center">
  Made with ❤️ by the Octo Team
</p>
