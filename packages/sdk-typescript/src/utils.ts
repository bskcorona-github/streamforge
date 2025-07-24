import { MetricData, QueryOptions } from './types';
import { ValidationError } from './errors';

/**
 * Validate metric data
 */
export function validateMetricData(data: MetricData): void {
  if (!data.name || typeof data.name !== 'string') {
    throw new ValidationError('Metric name is required and must be a string', 'name');
  }

  if (typeof data.value !== 'number' || isNaN(data.value)) {
    throw new ValidationError('Metric value is required and must be a number', 'value');
  }

  if (!data.timestamp || typeof data.timestamp !== 'number') {
    throw new ValidationError('Metric timestamp is required and must be a number', 'timestamp');
  }

  if (!data.type || !['counter', 'gauge', 'histogram', 'summary'].includes(data.type)) {
    throw new ValidationError('Metric type must be one of: counter, gauge, histogram, summary', 'type');
  }

  if (data.labels && typeof data.labels !== 'object') {
    throw new ValidationError('Metric labels must be an object', 'labels');
  }

  if (data.metadata && typeof data.metadata !== 'object') {
    throw new ValidationError('Metric metadata must be an object', 'metadata');
  }
}

/**
 * Validate query options
 */
export function validateQueryOptions(options: QueryOptions): void {
  if (options.startTime && (typeof options.startTime !== 'number' || options.startTime < 0)) {
    throw new ValidationError('Start time must be a positive number', 'startTime');
  }

  if (options.endTime && (typeof options.endTime !== 'number' || options.endTime < 0)) {
    throw new ValidationError('End time must be a positive number', 'endTime');
  }

  if (options.startTime && options.endTime && options.startTime >= options.endTime) {
    throw new ValidationError('Start time must be before end time');
  }

  if (options.limit && (typeof options.limit !== 'number' || options.limit <= 0)) {
    throw new ValidationError('Limit must be a positive number', 'limit');
  }

  if (options.offset && (typeof options.offset !== 'number' || options.offset < 0)) {
    throw new ValidationError('Offset must be a non-negative number', 'offset');
  }

  if (options.filters && typeof options.filters !== 'object') {
    throw new ValidationError('Filters must be an object', 'filters');
  }

  if (options.groupBy && !Array.isArray(options.groupBy)) {
    throw new ValidationError('Group by must be an array', 'groupBy');
  }

  if (options.aggregation && !['sum', 'avg', 'min', 'max', 'count'].includes(options.aggregation)) {
    throw new ValidationError('Aggregation must be one of: sum, avg, min, max, count', 'aggregation');
  }

  if (options.interval && typeof options.interval !== 'string') {
    throw new ValidationError('Interval must be a string', 'interval');
  }
}

/**
 * Convert timestamp to ISO string
 */
export function timestampToISO(timestamp: number): string {
  return new Date(timestamp).toISOString();
}

/**
 * Convert ISO string to timestamp
 */
export function isoToTimestamp(isoString: string): number {
  return new Date(isoString).getTime();
}

/**
 * Format bytes to human readable string
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Format duration to human readable string
 */
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  return `${Math.floor(seconds / 86400)}d ${Math.floor((seconds % 86400) / 3600)}h`;
}

/**
 * Generate a unique ID
 */
export function generateId(): string {
  return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Deep clone an object
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj;
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as unknown as T;
  }

  if (obj instanceof Array) {
    return obj.map(item => deepClone(item)) as unknown as T;
  }

  if (typeof obj === 'object') {
    const cloned = {} as T;
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        cloned[key] = deepClone(obj[key]);
      }
    }
    return cloned;
  }

  return obj;
}

/**
 * Debounce function
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: NodeJS.Timeout;
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
}

/**
 * Throttle function
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean;
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  };
} 