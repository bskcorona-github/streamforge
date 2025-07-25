#!/bin/bash

# StreamForge Development Environment Setup Script
# This script sets up the development environment for the StreamForge project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${BLUE}[SETUP]${NC} $1"
}

success() {
    echo -e "${GREEN}✅${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

error() {
    echo -e "${RED}❌${NC} $1"
}

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        error "$1 is not installed. Please install it first."
        return 1
    fi
    success "$1 is installed"
}

# Check required tools
log "Checking required tools..."
check_tool "docker"
check_tool "docker-compose"
check_tool "git"
check_tool "make"

# Check optional tools
log "Checking optional tools..."
if command -v node &> /dev/null; then
    success "Node.js is installed"
else
    warning "Node.js is not installed (optional for local development)"
fi

if command -v go &> /dev/null; then
    success "Go is installed"
else
    warning "Go is not installed (optional for local development)"
fi

if command -v cargo &> /dev/null; then
    success "Rust is installed"
else
    warning "Rust is not installed (optional for local development)"
fi

if command -v python3 &> /dev/null; then
    success "Python is installed"
else
    warning "Python is not installed (optional for local development)"
fi

# Create necessary directories
log "Creating necessary directories..."
mkdir -p scripts
mkdir -p .devcontainer/grafana/provisioning
mkdir -p .devcontainer/prometheus
mkdir -p logs
mkdir -p tmp

# Create Prometheus configuration
log "Creating Prometheus configuration..."
cat > .devcontainer/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'api-gateway'
    static_configs:
      - targets: ['host.docker.internal:8080']

  - job_name: 'collector'
    static_configs:
      - targets: ['host.docker.internal:8081']

  - job_name: 'stream-processor'
    static_configs:
      - targets: ['host.docker.internal:8082']

  - job_name: 'ml-engine'
    static_configs:
      - targets: ['host.docker.internal:8083']

  - job_name: 'operator'
    static_configs:
      - targets: ['host.docker.internal:8084']
EOF

# Create Grafana datasource configuration
log "Creating Grafana datasource configuration..."
mkdir -p .devcontainer/grafana/provisioning/datasources
cat > .devcontainer/grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

# Create Grafana dashboard configuration
log "Creating Grafana dashboard configuration..."
mkdir -p .devcontainer/grafana/provisioning/dashboards
cat > .devcontainer/grafana/provisioning/dashboards/dashboard.yml << 'EOF'
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# Create basic dashboard
mkdir -p .devcontainer/grafana/provisioning/dashboards
cat > .devcontainer/grafana/provisioning/dashboards/streamforge-overview.json << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "StreamForge Overview",
    "tags": ["streamforge"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Services Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up",
            "refId": "A"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": null},
                {"color": "green", "value": 1}
              ]
            }
          }
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF

# Create environment file
log "Creating environment file..."
cat > .env.dev << 'EOF'
# Development Environment Variables
NODE_ENV=development
LOG_LEVEL=debug

# Database
POSTGRES_DB=streamforge
POSTGRES_USER=streamforge
POSTGRES_PASSWORD=streamforge
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Services
API_GATEWAY_PORT=8080
COLLECTOR_PORT=8081
STREAM_PROCESSOR_PORT=8082
ML_ENGINE_PORT=8083
OPERATOR_PORT=8084

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001

# JWT
JWT_SECRET=dev-secret-key-change-in-production
JWT_EXPIRY=24h

# OpenTelemetry
OTEL_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=streamforge-dev
EOF

# Create health check script
log "Creating health check script..."
cat > scripts/health-check.sh << 'EOF'
#!/bin/bash

# Health check script for StreamForge services

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[HEALTH]${NC} $1"
}

success() {
    echo -e "${GREEN}✅${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

error() {
    echo -e "${RED}❌${NC} $1"
}

# Check if service is responding
check_service() {
    local name=$1
    local url=$2
    local timeout=${3:-5}
    
    if curl -s --max-time $timeout $url > /dev/null 2>&1; then
        success "$name is healthy"
        return 0
    else
        error "$name is not responding"
        return 1
    fi
}

# Check Docker containers
check_containers() {
    log "Checking Docker containers..."
    
    local containers=("postgres" "redis" "prometheus" "grafana")
    local healthy=0
    local total=${#containers[@]}
    
    for container in "${containers[@]}"; do
        if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$container.*Up"; then
            success "$container is running"
            ((healthy++))
        else
            error "$container is not running"
        fi
    done
    
    echo "Container health: $healthy/$total healthy"
}

# Check services
check_services() {
    log "Checking services..."
    
    local services=(
        "API Gateway:http://localhost:8080/health"
        "Collector:http://localhost:8081/health"
        "Stream Processor:http://localhost:8082/health"
        "ML Engine:http://localhost:8083/health"
        "Operator:http://localhost:8084/health"
        "Prometheus:http://localhost:9090/-/healthy"
        "Grafana:http://localhost:3001/api/health"
    )
    
    local healthy=0
    local total=${#services[@]}
    
    for service in "${services[@]}"; do
        IFS=':' read -r name url <<< "$service"
        if check_service "$name" "$url"; then
            ((healthy++))
        fi
    done
    
    echo "Service health: $healthy/$total healthy"
}

# Main execution
main() {
    echo "StreamForge Health Check"
    echo "======================="
    
    check_containers
    echo
    check_services
    
    echo
    log "Health check completed"
}

main "$@"
EOF

chmod +x scripts/health-check.sh

# Create demo data generation script
log "Creating demo data generation script..."
cat > scripts/generate-demo-data.sh << 'EOF'
#!/bin/bash

# Demo data generation script for StreamForge

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[DEMO]${NC} $1"
}

success() {
    echo -e "${GREEN}✅${NC} $1"
}

# Generate sample metrics
generate_metrics() {
    log "Generating sample metrics..."
    
    # Simulate API requests
    for i in {1..50}; do
        curl -s -X POST http://localhost:8080/api/v1/metrics \
            -H "Content-Type: application/json" \
            -d "{
                \"service\": \"demo-service-$((i % 5 + 1))\",
                \"metric\": \"request_duration\",
                \"value\": $((RANDOM % 1000 + 100)),
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"labels\": {
                    \"endpoint\": \"/api/v1/data\",
                    \"method\": \"POST\",
                    \"status_code\": \"200\"
                }
            }" > /dev/null
    done
    
    success "Generated 50 sample metrics"
}

# Generate sample traces
generate_traces() {
    log "Generating sample traces..."
    
    # Simulate distributed traces
    for i in {1..20}; do
        trace_id=$(uuidgen | tr '[:upper:]' '[:lower:]')
        
        # Root span
        curl -s -X POST http://localhost:8080/api/v1/traces \
            -H "Content-Type: application/json" \
            -d "{
                \"trace_id\": \"$trace_id\",
                \"span_id\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"operation_name\": \"process_request\",
                \"start_time\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"duration\": $((RANDOM % 1000 + 100)),
                \"service_name\": \"api-gateway\",
                \"tags\": {
                    \"http.method\": \"GET\",
                    \"http.url\": \"/api/v1/data\",
                    \"http.status_code\": \"200\"
                }
            }" > /dev/null
        
        # Child span
        curl -s -X POST http://localhost:8080/api/v1/traces \
            -H "Content-Type: application/json" \
            -d "{
                \"trace_id\": \"$trace_id\",
                \"span_id\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"parent_span_id\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"operation_name\": \"database_query\",
                \"start_time\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"duration\": $((RANDOM % 500 + 50)),
                \"service_name\": \"database\",
                \"tags\": {
                    \"db.type\": \"postgresql\",
                    \"db.statement\": \"SELECT * FROM metrics\"
                }
            }" > /dev/null
    done
    
    success "Generated 20 sample traces"
}

# Generate sample logs
generate_logs() {
    log "Generating sample logs..."
    
    # Simulate application logs
    for i in {1..100}; do
        level=$(["$((RANDOM % 10))" -lt 8 ] && echo "INFO" || echo "ERROR")
        service="service-$((i % 3 + 1))"
        
        curl -s -X POST http://localhost:8080/api/v1/logs \
            -H "Content-Type: application/json" \
            -d "{
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"level\": \"$level\",
                \"service\": \"$service\",
                \"message\": \"Sample log message $i\",
                \"trace_id\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"span_id\": \"$(uuidgen | tr '[:upper:]' '[:lower:]')\",
                \"attributes\": {
                    \"user_id\": \"user-$((i % 10 + 1))\",
                    \"request_id\": \"req-$(uuidgen | tr '[:upper:]' '[:lower:]')\"
                }
            }" > /dev/null
    done
    
    success "Generated 100 sample logs"
}

# Main execution
main() {
    log "Starting demo data generation..."
    
    # Wait for services to be ready
    sleep 10
    
    generate_metrics
    generate_traces
    generate_logs
    
    success "Demo data generation completed!"
    echo
    echo "You can now view the data in:"
    echo "  - Dashboard: http://localhost:3000"
    echo "  - Grafana: http://localhost:3001 (admin/admin)"
    echo "  - Prometheus: http://localhost:9090"
}

main "$@"
EOF

chmod +x scripts/generate-demo-data.sh

# Create tool check script
log "Creating tool check script..."
cat > scripts/check-tools.sh << 'EOF'
#!/bin/bash

# Tool check script for StreamForge development

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[TOOLS]${NC} $1"
}

success() {
    echo -e "${GREEN}✅${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

error() {
    echo -e "${RED}❌${NC} $1"
}

# Check tool version
check_tool_version() {
    local tool=$1
    local command=$2
    local version_flag=${3:-"--version"}
    
    if command -v $command &> /dev/null; then
        local version=$($command $version_flag 2>&1 | head -n 1)
        success "$tool: $version"
        return 0
    else
        error "$tool is not installed"
        return 1
    fi
}

# Required tools
log "Checking required tools..."
required_tools=(
    "Docker:docker"
    "Docker Compose:docker-compose"
    "Git:git"
    "Make:make"
)

for tool_info in "${required_tools[@]}"; do
    IFS=':' read -r name command <<< "$tool_info"
    check_tool_version "$name" "$command"
done

# Optional tools
log "Checking optional tools..."
optional_tools=(
    "Node.js:node"
    "Go:go"
    "Rust:cargo"
    "Python:python3"
    "kubectl:kubectl"
    "Helm:helm"
)

for tool_info in "${optional_tools[@]}"; do
    IFS=':' read -r name command <<< "$tool_info"
    if check_tool_version "$name" "$command" > /dev/null 2>&1; then
        success "$name is available"
    else
        warning "$name is not installed (optional)"
    fi
done

# Check Docker
log "Checking Docker setup..."
if docker info > /dev/null 2>&1; then
    success "Docker is running"
else
    error "Docker is not running"
fi

# Check available ports
log "Checking required ports..."
ports=(3000 8080 8081 8082 8083 8084 9090 3001 5432 6379)
for port in "${ports[@]}"; do
    if netstat -tuln 2>/dev/null | grep -q ":$port "; then
        warning "Port $port is already in use"
    else
        success "Port $port is available"
    fi
done

success "Tool check completed"
EOF

chmod +x scripts/check-tools.sh

# Set permissions
log "Setting script permissions..."
chmod +x scripts/*.sh

success "Development environment setup completed!"
echo
echo "Next steps:"
echo "1. Run 'make start-dev' to start all services"
echo "2. Run 'make demo' to generate sample data"
echo "3. Visit http://localhost:3000 for the dashboard"
echo "4. Visit http://localhost:3001 for Grafana (admin/admin)" 