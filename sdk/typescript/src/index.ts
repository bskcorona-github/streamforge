/**
 * StreamForge TypeScript SDK
 * Real-time monitoring and analytics for distributed streaming systems
 */

export interface Metric {
  name: string
  value: number
  unit: string
  labels?: Record<string, string>
  timestamp?: Date
}

export interface LogEntry {
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  fields?: Record<string, any>
  timestamp?: Date
}

export interface Alert {
  id: string
  severity: 'critical' | 'warning' | 'info'
  message: string
  timestamp: Date
  service: string
  metadata?: Record<string, any>
}

export interface ServiceStatus {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  uptime: number
  responseTime: number
  requestsPerSecond: number
  lastCheck: Date
}

export interface StreamForgeConfig {
  apiUrl: string
  wsUrl: string
  apiKey?: string
  timeout?: number
  retries?: number
}

export class StreamForgeError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public code?: string
  ) {
    super(message)
    this.name = 'StreamForgeError'
  }
}

export class StreamForgeClient {
  private config: StreamForgeConfig
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  constructor(config: StreamForgeConfig) {
    this.config = {
      timeout: 30000,
      retries: 3,
      ...config
    }
  }

  /**
   * メトリクスを送信
   */
  async sendMetrics(metrics: Metric[]): Promise<void> {
    try {
      const response = await this.makeRequest('/api/v1/metrics', {
        method: 'POST',
        body: JSON.stringify({ metrics })
      })

      if (!response.ok) {
        throw new StreamForgeError(
          `Failed to send metrics: ${response.statusText}`,
          response.status
        )
      }
    } catch (error) {
      throw new StreamForgeError(
        `Failed to send metrics: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  /**
   * ログを送信
   */
  async sendLogs(logs: LogEntry[]): Promise<void> {
    try {
      const response = await this.makeRequest('/api/v1/logs', {
        method: 'POST',
        body: JSON.stringify({ logs })
      })

      if (!response.ok) {
        throw new StreamForgeError(
          `Failed to send logs: ${response.statusText}`,
          response.status
        )
      }
    } catch (error) {
      throw new StreamForgeError(
        `Failed to send logs: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  /**
   * メトリクスを取得
   */
  async getMetrics(filters?: {
    name?: string
    labels?: Record<string, string>
    startTime?: Date
    endTime?: Date
    limit?: number
  }): Promise<Metric[]> {
    try {
      const params = new URLSearchParams()
      if (filters?.name) params.append('name', filters.name)
      if (filters?.startTime) params.append('startTime', filters.startTime.toISOString())
      if (filters?.endTime) params.append('endTime', filters.endTime.toISOString())
      if (filters?.limit) params.append('limit', filters.limit.toString())
      if (filters?.labels) {
        Object.entries(filters.labels).forEach(([key, value]) => {
          params.append(`labels.${key}`, value)
        })
      }

      const response = await this.makeRequest(`/api/v1/metrics?${params.toString()}`)
      
      if (!response.ok) {
        throw new StreamForgeError(
          `Failed to get metrics: ${response.statusText}`,
          response.status
        )
      }

      const data = await response.json()
      return data.metrics.map((m: any) => ({
        ...m,
        timestamp: new Date(m.timestamp)
      }))
    } catch (error) {
      throw new StreamForgeError(
        `Failed to get metrics: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  /**
   * アラートを取得
   */
  async getAlerts(filters?: {
    severity?: string
    service?: string
    startTime?: Date
    endTime?: Date
    limit?: number
  }): Promise<Alert[]> {
    try {
      const params = new URLSearchParams()
      if (filters?.severity) params.append('severity', filters.severity)
      if (filters?.service) params.append('service', filters.service)
      if (filters?.startTime) params.append('startTime', filters.startTime.toISOString())
      if (filters?.endTime) params.append('endTime', filters.endTime.toISOString())
      if (filters?.limit) params.append('limit', filters.limit.toString())

      const response = await this.makeRequest(`/api/v1/alerts?${params.toString()}`)
      
      if (!response.ok) {
        throw new StreamForgeError(
          `Failed to get alerts: ${response.statusText}`,
          response.status
        )
      }

      const data = await response.json()
      return data.alerts.map((a: any) => ({
        ...a,
        timestamp: new Date(a.timestamp)
      }))
    } catch (error) {
      throw new StreamForgeError(
        `Failed to get alerts: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  /**
   * サービスステータスを取得
   */
  async getServiceStatus(): Promise<ServiceStatus[]> {
    try {
      const response = await this.makeRequest('/api/v1/services/status')
      
      if (!response.ok) {
        throw new StreamForgeError(
          `Failed to get service status: ${response.statusText}`,
          response.status
        )
      }

      const data = await response.json()
      return data.services.map((s: any) => ({
        ...s,
        lastCheck: new Date(s.lastCheck)
      }))
    } catch (error) {
      throw new StreamForgeError(
        `Failed to get service status: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  /**
   * WebSocket接続を開始
   */
  connectWebSocket(callbacks: {
    onMetrics?: (metrics: Metric[]) => void
    onAlerts?: (alerts: Alert[]) => void
    onServiceStatus?: (status: ServiceStatus[]) => void
    onError?: (error: Error) => void
    onConnect?: () => void
    onDisconnect?: () => void
  }): void {
    try {
      this.ws = new WebSocket(this.config.wsUrl)
      
      this.ws.onopen = () => {
        this.reconnectAttempts = 0
        callbacks.onConnect?.()
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          
          switch (data.type) {
            case 'metrics':
              callbacks.onMetrics?.(data.metrics.map((m: any) => ({
                ...m,
                timestamp: new Date(m.timestamp)
              })))
              break
            case 'alerts':
              callbacks.onAlerts?.(data.alerts.map((a: any) => ({
                ...a,
                timestamp: new Date(a.timestamp)
              })))
              break
            case 'service_status':
              callbacks.onServiceStatus?.(data.services.map((s: any) => ({
                ...s,
                lastCheck: new Date(s.lastCheck)
              })))
              break
          }
        } catch (error) {
          callbacks.onError?.(new Error(`Failed to parse WebSocket message: ${error}`))
        }
      }

      this.ws.onerror = (error) => {
        callbacks.onError?.(new Error(`WebSocket error: ${error}`))
      }

      this.ws.onclose = () => {
        callbacks.onDisconnect?.()
        this.attemptReconnect(callbacks)
      }
    } catch (error) {
      callbacks.onError?.(new Error(`Failed to connect WebSocket: ${error}`))
    }
  }

  /**
   * WebSocket接続を切断
   */
  disconnectWebSocket(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * ヘルスチェック
   */
  async healthCheck(): Promise<{ status: string; timestamp: Date }> {
    try {
      const response = await this.makeRequest('/api/v1/health')
      
      if (!response.ok) {
        throw new StreamForgeError(
          `Health check failed: ${response.statusText}`,
          response.status
        )
      }

      const data = await response.json()
      return {
        status: data.status,
        timestamp: new Date(data.timestamp)
      }
    } catch (error) {
      throw new StreamForgeError(
        `Health check failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      )
    }
  }

  private async makeRequest(
    path: string,
    options: RequestInit = {}
  ): Promise<Response> {
    const url = `${this.config.apiUrl}${path}`
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers as Record<string, string>
    }

    if (this.config.apiKey) {
      headers['Authorization'] = `Bearer ${this.config.apiKey}`
    }

    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout)

    try {
      const response = await fetch(url, {
        ...options,
        headers,
        signal: controller.signal
      })

      clearTimeout(timeoutId)
      return response
    } catch (error) {
      clearTimeout(timeoutId)
      throw error
    }
  }

  private attemptReconnect(callbacks: any): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        this.connectWebSocket(callbacks)
      }, this.reconnectDelay * this.reconnectAttempts)
    }
  }
}

// 便利な関数
export function createMetric(
  name: string,
  value: number,
  unit: string,
  labels?: Record<string, string>
): Metric {
  return {
    name,
    value,
    unit,
    labels,
    timestamp: new Date()
  }
}

export function createLogEntry(
  level: LogEntry['level'],
  message: string,
  fields?: Record<string, any>
): LogEntry {
  return {
    level,
    message,
    fields,
    timestamp: new Date()
  }
}

// デフォルトエクスポート
export default StreamForgeClient 