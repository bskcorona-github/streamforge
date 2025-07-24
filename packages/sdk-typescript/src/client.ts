import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import { StreamForgeConfig, MetricData, QueryOptions, QueryResult } from './types';
import { StreamForgeError } from './errors';
import { validateMetricData, validateQueryOptions } from './utils';

/**
 * StreamForge Client for interacting with the StreamForge API
 */
export class StreamForgeClient {
  private client: AxiosInstance;
  private config: StreamForgeConfig;

  constructor(config: StreamForgeConfig) {
    this.config = {
      baseURL: 'http://localhost:8080',
      timeout: 30000,
      retries: 3,
      ...config,
    };

    this.client = axios.create({
      baseURL: this.config.baseURL,
      timeout: this.config.timeout,
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': `StreamForge-SDK-TypeScript/${VERSION}`,
      },
    });

    // Add request interceptor for retries
    this.client.interceptors.request.use(this.handleRequest.bind(this));
    this.client.interceptors.response.use(
      this.handleResponse.bind(this),
      this.handleError.bind(this)
    );
  }

  /**
   * Send metrics data to StreamForge
   */
  async sendMetrics(data: MetricData[]): Promise<void> {
    try {
      // Validate input data
      for (const metric of data) {
        validateMetricData(metric);
      }

      const payload = {
        timestamp: Date.now(),
        batch_id: this.generateBatchId(),
        data: data,
      };

      await this.client.post('/api/v1/metrics', payload);
    } catch (error) {
      throw new StreamForgeError('Failed to send metrics', error);
    }
  }

  /**
   * Query metrics from StreamForge
   */
  async queryMetrics(options: QueryOptions): Promise<QueryResult> {
    try {
      validateQueryOptions(options);

      const params = new URLSearchParams();
      if (options.startTime) params.append('start_time', options.startTime.toString());
      if (options.endTime) params.append('end_time', options.endTime.toString());
      if (options.limit) params.append('limit', options.limit.toString());
      if (options.offset) params.append('offset', options.offset.toString());
      if (options.filters) {
        for (const [key, value] of Object.entries(options.filters)) {
          params.append(`filter_${key}`, value);
        }
      }

      const response = await this.client.get(`/api/v1/metrics?${params.toString()}`);
      return response.data;
    } catch (error) {
      throw new StreamForgeError('Failed to query metrics', error);
    }
  }

  /**
   * Get system health status
   */
  async getHealth(): Promise<{ status: string; timestamp: number }> {
    try {
      const response = await this.client.get('/health');
      return response.data;
    } catch (error) {
      throw new StreamForgeError('Failed to get health status', error);
    }
  }

  /**
   * Get system metrics
   */
  async getSystemMetrics(): Promise<any> {
    try {
      const response = await this.client.get('/api/v1/system/metrics');
      return response.data;
    } catch (error) {
      throw new StreamForgeError('Failed to get system metrics', error);
    }
  }

  /**
   * Create a new alert rule
   */
  async createAlertRule(rule: any): Promise<any> {
    try {
      const response = await this.client.post('/api/v1/alerts/rules', rule);
      return response.data;
    } catch (error) {
      throw new StreamForgeError('Failed to create alert rule', error);
    }
  }

  /**
   * Get alert rules
   */
  async getAlertRules(): Promise<any[]> {
    try {
      const response = await this.client.get('/api/v1/alerts/rules');
      return response.data;
    } catch (error) {
      throw new StreamForgeError('Failed to get alert rules', error);
    }
  }

  /**
   * Delete an alert rule
   */
  async deleteAlertRule(ruleId: string): Promise<void> {
    try {
      await this.client.delete(`/api/v1/alerts/rules/${ruleId}`);
    } catch (error) {
      throw new StreamForgeError('Failed to delete alert rule', error);
    }
  }

  /**
   * Handle request interceptor
   */
  private handleRequest(config: AxiosRequestConfig): AxiosRequestConfig {
    // Add authentication if configured
    if (this.config.apiKey) {
      config.headers = {
        ...config.headers,
        'Authorization': `Bearer ${this.config.apiKey}`,
      };
    }
    return config;
  }

  /**
   * Handle response interceptor
   */
  private handleResponse(response: any): any {
    return response;
  }

  /**
   * Handle error interceptor with retry logic
   */
  private async handleError(error: any): Promise<any> {
    const config = error.config;
    
    if (!config || !config.retry) {
      config.retry = 0;
    }

    if (config.retry >= this.config.retries) {
      return Promise.reject(error);
    }

    config.retry += 1;
    
    // Exponential backoff
    const delay = Math.pow(2, config.retry) * 1000;
    await new Promise(resolve => setTimeout(resolve, delay));

    return this.client(config);
  }

  /**
   * Generate a unique batch ID
   */
  private generateBatchId(): string {
    return `batch_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}

// Import VERSION from index
import { VERSION } from './index'; 