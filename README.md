### Open Image Registry


<img src="docs/assets/logo.png" width="150" height="150" alt="Logo">

A powerful, open-source Docker image registry with an intuitive web interface, built for scalability and simplicity. Host, proxy, and cache Docker images all in one place.

> ⚠️ **Work in Progress** - This project is actively under development. APIs and features may change. We welcome contributions and feedback!

## Features

- **Built-in Web UI** - Manage your Docker images through a modern, user-friendly dashboard
- **Horizontally Scalable** - Deploy multiple instances with proper database configuration for unlimited growth
- **Docker Image Hosting** - Store and distribute your own Docker images securely
- **Registry Proxying & Caching** - Cache images from Docker Hub, Quay.io, and other popular registries to reduce bandwidth and improve performance
- **Multiple Storage Backends** - Support for local filesystem, S3, and cloud storage solutions
- **API-First Design** - Complete REST API for programmatic access and integration
- **Authentication & Authorization** - Secure access control for private registries
- **Multi-Tenancy Ready** - Organize images across multiple namespaces and teams

## Use Cases

- **Private Docker Registry** - Host proprietary images for your organization
- **Image Caching Proxy** - Speed up CI/CD pipelines by caching public images locally
- **Air-Gapped Deployments** - Mirror registries for isolated environments
- **Multi-Tenant SaaS** - Provide registry services to multiple customers
- **Edge Computing** - Deploy lightweight registries closer to your infrastructure

## Quick Start

### Prerequisites

- Go 1.19+ (for server)
- Node.js 16+ (for WebUI)

### Clone the Repository

```bash
git clone git@github.com:ksankeerth/open-image-registry.git
cd open-image-registry
```

### Run the Server

```bash
make run-server
```

The server will start on `http://localhost:8000` by default.

### Run the WebUI (Development Mode)

In a separate terminal:

```bash
cd webapp
npm install
npm start
```

The webpack development server will start on `http://localhost:3000`.

## Architecture & Scaling

Open Image Registry is designed to scale horizontally:

- **Database Layer** - Use managed PostgreSQL (AWS RDS, GCP Cloud SQL, etc.)
- **Storage Layer** - Use S3, Azure Blob Storage, or cloud-native storage
- **Registry Instances** - Deploy multiple instances behind a load balancer

## Contributing

We're actively developing Open Image Registry and welcome contributions! 

