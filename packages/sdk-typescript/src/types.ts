/**
 * StreamForge client configuration
 */
export interface StreamForgeConfig {
  /** Base URL for the StreamForge API */
  baseURL?: string;
  /** Request timeout in milliseconds */
  timeout?: number;
  /** Number of retries for failed requests */
  retries?: number;
  /** API key for authentication */
  apiKey?: string;
}

/**
 * Metric data structure
 */
export interface MetricData {
  /** Metric name */
  name: string;
  /** Metric value */
  value: number;
  /** Metric timestamp (Unix timestamp) */
  timestamp: number;
  /** Metric type */
  type: 'counter' | 'gauge' | 'histogram' | 'summary';
  /** Metric labels/tags */
  labels?: Record<string, string>;
  /** Additional metadata */
  metadata?: Record<string, any>;
}

/**
 * Query options for metrics
 */
export interface QueryOptions {
  /** Start time for the query (Unix timestamp) */
  startTime?: number;
  /** End time for the query (Unix timestamp) */
  endTime?: number;
  /** Maximum number of results to return */
  limit?: number;
  /** Number of results to skip */
  offset?: number;
  /** Filters to apply to the query */
  filters?: Record<string, string>;
  /** Group by fields */
  groupBy?: string[];
  /** Aggregation function */
  aggregation?: 'sum' | 'avg' | 'min' | 'max' | 'count';
  /** Time interval for aggregation */
  interval?: string;
}

/**
 * Query result structure
 */
export interface QueryResult {
  /** Query results */
  data: MetricData[];
  /** Total number of results */
  total: number;
  /** Query execution time in milliseconds */
  executionTime: number;
  /** Query metadata */
  metadata?: Record<string, any>;
}

/**
 * Alert rule configuration
 */
export interface AlertRule {
  /** Unique identifier for the alert rule */
  id: string;
  /** Alert rule name */
  name: string;
  /** Alert rule description */
  description?: string;
  /** Query to evaluate */
  query: string;
  /** Condition for triggering the alert */
  condition: AlertCondition;
  /** Alert severity */
  severity: 'low' | 'medium' | 'high' | 'critical';
  /** Alert notification channels */
  notifications: NotificationChannel[];
  /** Whether the alert rule is enabled */
  enabled: boolean;
  /** Alert rule creation timestamp */
  createdAt: number;
  /** Alert rule last update timestamp */
  updatedAt: number;
}

/**
 * Alert condition
 */
export interface AlertCondition {
  /** Condition operator */
  operator: '>' | '<' | '>=' | '<=' | '==' | '!=';
  /** Threshold value */
  threshold: number;
  /** Duration for which the condition must be true */
  duration: string;
}

/**
 * Notification channel
 */
export interface NotificationChannel {
  /** Channel type */
  type: 'email' | 'slack' | 'webhook' | 'pagerduty';
  /** Channel configuration */
  config: Record<string, any>;
}

/**
 * System health status
 */
export interface HealthStatus {
  /** Overall health status */
  status: 'healthy' | 'degraded' | 'unhealthy';
  /** Health check timestamp */
  timestamp: number;
  /** Component health status */
  components: Record<string, ComponentHealth>;
}

/**
 * Component health status
 */
export interface ComponentHealth {
  /** Component status */
  status: 'healthy' | 'degraded' | 'unhealthy';
  /** Component message */
  message?: string;
  /** Component last check timestamp */
  lastCheck: number;
}

/**
 * System metrics
 */
export interface SystemMetrics {
  /** CPU usage percentage */
  cpuUsage: number;
  /** Memory usage percentage */
  memoryUsage: number;
  /** Disk usage percentage */
  diskUsage: number;
  /** Network bytes sent */
  networkBytesSent: number;
  /** Network bytes received */
  networkBytesReceived: number;
  /** Active connections */
  activeConnections: number;
  /** Uptime in seconds */
  uptime: number;
  /** Timestamp */
  timestamp: number;
} 