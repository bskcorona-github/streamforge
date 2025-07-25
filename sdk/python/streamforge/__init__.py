"""
StreamForge Python SDK
Real-time monitoring and analytics for distributed streaming systems
"""

import asyncio
import json
import time
from datetime import datetime
from typing import Dict, List, Optional, Any, Callable
import aiohttp
import websockets
from dataclasses import dataclass, asdict
from enum import Enum


class LogLevel(Enum):
    DEBUG = "debug"
    INFO = "info"
    WARN = "warn"
    ERROR = "error"


@dataclass
class Metric:
    """Metric data point"""
    name: str
    value: float
    unit: str
    labels: Optional[Dict[str, str]] = None
    timestamp: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        data = asdict(self)
        if self.timestamp:
            data['timestamp'] = self.timestamp.isoformat()
        return data


@dataclass
class LogEntry:
    """Log entry"""
    level: str
    message: str
    fields: Optional[Dict[str, Any]] = None
    timestamp: Optional[datetime] = None

    def to_dict(self) -> Dict[str, Any]:
        data = asdict(self)
        if self.timestamp:
            data['timestamp'] = self.timestamp.isoformat()
        return data


@dataclass
class Alert:
    """Alert"""
    id: str
    severity: str
    message: str
    timestamp: datetime
    service: str
    metadata: Optional[Dict[str, Any]] = None

    def to_dict(self) -> Dict[str, Any]:
        data = asdict(self)
        data['timestamp'] = self.timestamp.isoformat()
        return data


@dataclass
class ServiceStatus:
    """Service status"""
    name: str
    status: str
    uptime: float
    response_time: float
    requests_per_second: float
    last_check: datetime

    def to_dict(self) -> Dict[str, Any]:
        data = asdict(self)
        data['last_check'] = self.last_check.isoformat()
        return data


@dataclass
class HealthCheck:
    """Health check response"""
    status: str
    timestamp: datetime

    def to_dict(self) -> Dict[str, Any]:
        data = asdict(self)
        data['timestamp'] = self.timestamp.isoformat()
        return data


class StreamForgeError(Exception):
    """StreamForge API error"""
    def __init__(self, message: str, status_code: Optional[int] = None, code: Optional[str] = None):
        self.message = message
        self.status_code = status_code
        self.code = code
        super().__init__(self.message)


class StreamForgeClient:
    """StreamForge client"""
    
    def __init__(self, config: Dict[str, Any]):
        self.api_url = config.get('api_url', 'http://localhost:8080')
        self.ws_url = config.get('ws_url', 'ws://localhost:8080')
        self.api_key = config.get('api_key')
        self.timeout = config.get('timeout', 30)
        self.retries = config.get('retries', 3)
        self.session: Optional[aiohttp.ClientSession] = None
        self.ws: Optional[websockets.WebSocketServerProtocol] = None

    async def __aenter__(self):
        self.session = aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(total=self.timeout)
        )
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        if self.session:
            await self.session.close()
        if self.ws:
            await self.ws.close()

    async def send_metrics(self, metrics: List[Metric]) -> None:
        """Send metrics to the StreamForge API"""
        payload = {
            'metrics': [metric.to_dict() for metric in metrics]
        }
        
        await self._make_request('POST', '/api/v1/metrics', payload)

    async def send_logs(self, logs: List[LogEntry]) -> None:
        """Send logs to the StreamForge API"""
        payload = {
            'logs': [log.to_dict() for log in logs]
        }
        
        await self._make_request('POST', '/api/v1/logs', payload)

    async def get_metrics(self, filters: Optional[Dict[str, str]] = None) -> List[Metric]:
        """Get metrics from the StreamForge API"""
        params = filters or {}
        response = await self._make_request('GET', '/api/v1/metrics', None, params)
        
        metrics_data = response.get('metrics', [])
        metrics = []
        
        for metric_data in metrics_data:
            timestamp = None
            if 'timestamp' in metric_data:
                timestamp = datetime.fromisoformat(metric_data['timestamp'].replace('Z', '+00:00'))
            
            metric = Metric(
                name=metric_data['name'],
                value=metric_data['value'],
                unit=metric_data['unit'],
                labels=metric_data.get('labels'),
                timestamp=timestamp
            )
            metrics.append(metric)
        
        return metrics

    async def get_alerts(self, filters: Optional[Dict[str, str]] = None) -> List[Alert]:
        """Get alerts from the StreamForge API"""
        params = filters or {}
        response = await self._make_request('GET', '/api/v1/alerts', None, params)
        
        alerts_data = response.get('alerts', [])
        alerts = []
        
        for alert_data in alerts_data:
            timestamp = datetime.fromisoformat(alert_data['timestamp'].replace('Z', '+00:00'))
            
            alert = Alert(
                id=alert_data['id'],
                severity=alert_data['severity'],
                message=alert_data['message'],
                timestamp=timestamp,
                service=alert_data['service'],
                metadata=alert_data.get('metadata')
            )
            alerts.append(alert)
        
        return alerts

    async def get_service_status(self) -> List[ServiceStatus]:
        """Get service status from the StreamForge API"""
        response = await self._make_request('GET', '/api/v1/services/status')
        
        services_data = response.get('services', [])
        services = []
        
        for service_data in services_data:
            last_check = datetime.fromisoformat(service_data['last_check'].replace('Z', '+00:00'))
            
            service = ServiceStatus(
                name=service_data['name'],
                status=service_data['status'],
                uptime=service_data['uptime'],
                response_time=service_data['response_time'],
                requests_per_second=service_data['requests_per_second'],
                last_check=last_check
            )
            services.append(service)
        
        return services

    async def health_check(self) -> HealthCheck:
        """Perform a health check"""
        response = await self._make_request('GET', '/api/v1/health')
        
        timestamp = datetime.fromisoformat(response['timestamp'].replace('Z', '+00:00'))
        
        return HealthCheck(
            status=response['status'],
            timestamp=timestamp
        )

    async def connect_websocket(self, callbacks: Dict[str, Callable]) -> None:
        """Connect to WebSocket"""
        ws_url = self.ws_url
        if self.api_key:
            ws_url += f"?api_key={self.api_key}"

        try:
            self.ws = await websockets.connect(ws_url)
            
            if 'on_connect' in callbacks:
                callbacks['on_connect']()

            # Start listening for messages
            asyncio.create_task(self._listen_websocket(callbacks))
            
        except Exception as e:
            if 'on_error' in callbacks:
                callbacks['on_error'](f"Failed to connect WebSocket: {e}")
            raise

    async def disconnect_websocket(self) -> None:
        """Disconnect from WebSocket"""
        if self.ws:
            await self.ws.close()
            self.ws = None

    async def _listen_websocket(self, callbacks: Dict[str, Callable]) -> None:
        """Listen for WebSocket messages"""
        try:
            async for message in self.ws:
                try:
                    data = json.loads(message)
                    message_type = data.get('type')
                    
                    if message_type == 'metrics' and 'on_metrics' in callbacks:
                        metrics_data = data.get('metrics', [])
                        metrics = []
                        for metric_data in metrics_data:
                            timestamp = None
                            if 'timestamp' in metric_data:
                                timestamp = datetime.fromisoformat(metric_data['timestamp'].replace('Z', '+00:00'))
                            
                            metric = Metric(
                                name=metric_data['name'],
                                value=metric_data['value'],
                                unit=metric_data['unit'],
                                labels=metric_data.get('labels'),
                                timestamp=timestamp
                            )
                            metrics.append(metric)
                        
                        callbacks['on_metrics'](metrics)
                    
                    elif message_type == 'alerts' and 'on_alerts' in callbacks:
                        alerts_data = data.get('alerts', [])
                        alerts = []
                        for alert_data in alerts_data:
                            timestamp = datetime.fromisoformat(alert_data['timestamp'].replace('Z', '+00:00'))
                            
                            alert = Alert(
                                id=alert_data['id'],
                                severity=alert_data['severity'],
                                message=alert_data['message'],
                                timestamp=timestamp,
                                service=alert_data['service'],
                                metadata=alert_data.get('metadata')
                            )
                            alerts.append(alert)
                        
                        callbacks['on_alerts'](alerts)
                    
                    elif message_type == 'service_status' and 'on_service_status' in callbacks:
                        services_data = data.get('services', [])
                        services = []
                        for service_data in services_data:
                            last_check = datetime.fromisoformat(service_data['last_check'].replace('Z', '+00:00'))
                            
                            service = ServiceStatus(
                                name=service_data['name'],
                                status=service_data['status'],
                                uptime=service_data['uptime'],
                                response_time=service_data['response_time'],
                                requests_per_second=service_data['requests_per_second'],
                                last_check=last_check
                            )
                            services.append(service)
                        
                        callbacks['on_service_status'](services)
                
                except json.JSONDecodeError as e:
                    if 'on_error' in callbacks:
                        callbacks['on_error'](f"Failed to parse WebSocket message: {e}")
                except Exception as e:
                    if 'on_error' in callbacks:
                        callbacks['on_error'](f"WebSocket message error: {e}")
        
        except websockets.exceptions.ConnectionClosed:
            if 'on_disconnect' in callbacks:
                callbacks['on_disconnect']()
        except Exception as e:
            if 'on_error' in callbacks:
                callbacks['on_error'](f"WebSocket error: {e}")

    async def _make_request(
        self, 
        method: str, 
        path: str, 
        payload: Optional[Dict[str, Any]] = None,
        params: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """Make HTTP request"""
        if not self.session:
            raise RuntimeError("Client not initialized. Use async context manager.")

        url = f"{self.api_url}{path}"
        headers = {'Content-Type': 'application/json'}
        
        if self.api_key:
            headers['Authorization'] = f"Bearer {self.api_key}"

        for attempt in range(self.retries):
            try:
                if method == 'GET':
                    async with self.session.get(url, headers=headers, params=params) as response:
                        return await self._handle_response(response)
                elif method == 'POST':
                    async with self.session.post(url, headers=headers, json=payload) as response:
                        return await self._handle_response(response)
                else:
                    raise ValueError(f"Unsupported HTTP method: {method}")
            
            except Exception as e:
                if attempt == self.retries - 1:
                    raise StreamForgeError(f"Request failed after {self.retries} attempts: {e}")
                await asyncio.sleep(2 ** attempt)  # Exponential backoff

    async def _handle_response(self, response: aiohttp.ClientResponse) -> Dict[str, Any]:
        """Handle HTTP response"""
        if response.status >= 400:
            try:
                error_data = await response.json()
                raise StreamForgeError(
                    message=error_data.get('message', 'Unknown error'),
                    status_code=response.status,
                    code=error_data.get('code')
                )
            except:
                raise StreamForgeError(
                    message=f"Request failed with status {response.status}",
                    status_code=response.status
                )
        
        return await response.json()


# Helper functions
def create_metric(name: str, value: float, unit: str, labels: Optional[Dict[str, str]] = None) -> Metric:
    """Create a metric"""
    return Metric(
        name=name,
        value=value,
        unit=unit,
        labels=labels,
        timestamp=datetime.now()
    )


def create_log_entry(level: LogLevel, message: str, fields: Optional[Dict[str, Any]] = None) -> LogEntry:
    """Create a log entry"""
    return LogEntry(
        level=level.value,
        message=message,
        fields=fields,
        timestamp=datetime.now()
    )


# Convenience functions for synchronous usage
def create_sync_client(config: Dict[str, Any]) -> 'SyncStreamForgeClient':
    """Create a synchronous StreamForge client"""
    return SyncStreamForgeClient(config)


class SyncStreamForgeClient:
    """Synchronous StreamForge client wrapper"""
    
    def __init__(self, config: Dict[str, Any]):
        self.config = config
        self.loop = None

    def _get_loop(self):
        """Get or create event loop"""
        try:
            return asyncio.get_running_loop()
        except RuntimeError:
            if self.loop is None:
                self.loop = asyncio.new_event_loop()
                asyncio.set_event_loop(self.loop)
            return self.loop

    def send_metrics(self, metrics: List[Metric]) -> None:
        """Send metrics synchronously"""
        async def _send():
            async with StreamForgeClient(self.config) as client:
                await client.send_metrics(metrics)
        
        self._get_loop().run_until_complete(_send())

    def send_logs(self, logs: List[LogEntry]) -> None:
        """Send logs synchronously"""
        async def _send():
            async with StreamForgeClient(self.config) as client:
                await client.send_logs(logs)
        
        self._get_loop().run_until_complete(_send())

    def get_metrics(self, filters: Optional[Dict[str, str]] = None) -> List[Metric]:
        """Get metrics synchronously"""
        async def _get():
            async with StreamForgeClient(self.config) as client:
                return await client.get_metrics(filters)
        
        return self._get_loop().run_until_complete(_get())

    def get_alerts(self, filters: Optional[Dict[str, str]] = None) -> List[Alert]:
        """Get alerts synchronously"""
        async def _get():
            async with StreamForgeClient(self.config) as client:
                return await client.get_alerts(filters)
        
        return self._get_loop().run_until_complete(_get())

    def get_service_status(self) -> List[ServiceStatus]:
        """Get service status synchronously"""
        async def _get():
            async with StreamForgeClient(self.config) as client:
                return await client.get_service_status()
        
        return self._get_loop().run_until_complete(_get())

    def health_check(self) -> HealthCheck:
        """Perform health check synchronously"""
        async def _check():
            async with StreamForgeClient(self.config) as client:
                return await client.health_check()
        
        return self._get_loop().run_until_complete(_check())


# Version
__version__ = "0.1.0" 