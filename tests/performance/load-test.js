const autocannon = require('autocannon');
const { expect } = require('chai');

describe('Performance Tests', () => {
  const baseURL = process.env.API_GATEWAY_URL || 'http://localhost:8085';

  describe('API Gateway Load Test', () => {
    it('should handle high load on health endpoint', async () => {
      const result = await autocannon({
        url: `${baseURL}/health`,
        connections: 10,
        duration: 10,
        pipelining: 1,
        timeout: 10
      });

      console.log('Health endpoint performance:', {
        requests: result.requests,
        throughput: result.throughput,
        latency: result.latency,
        errors: result.errors
      });

      expect(result.errors).to.equal(0);
      expect(result.requests.total).to.be.greaterThan(100);
      expect(result.latency.p99).to.be.lessThan(1000); // 99th percentile < 1s
    });

    it('should handle concurrent metric submissions', async () => {
      const metricData = {
        service: 'load-test-service',
        metric: 'test_metric',
        value: Math.random() * 100,
        timestamp: new Date().toISOString(),
        labels: {
          test: 'load-test',
          iteration: '1'
        }
      };

      const result = await autocannon({
        url: `${baseURL}/api/v1/metrics`,
        method: 'POST',
        headers: {
          'content-type': 'application/json'
        },
        body: JSON.stringify(metricData),
        connections: 5,
        duration: 10,
        pipelining: 1,
        timeout: 10
      });

      console.log('Metrics endpoint performance:', {
        requests: result.requests,
        throughput: result.throughput,
        latency: result.latency,
        errors: result.errors
      });

      expect(result.errors).to.be.lessThan(5); // Allow some errors under load
      expect(result.requests.total).to.be.greaterThan(50);
      expect(result.latency.p95).to.be.lessThan(2000); // 95th percentile < 2s
    });
  });

  describe('Database Performance', () => {
    it('should handle concurrent queries', async () => {
      const queryData = {
        service: 'load-test-service',
        metric: 'test_metric',
        start_time: new Date(Date.now() - 3600000).toISOString(),
        end_time: new Date().toISOString()
      };

      const result = await autocannon({
        url: `${baseURL}/api/v1/query/metrics`,
        method: 'POST',
        headers: {
          'content-type': 'application/json'
        },
        body: JSON.stringify(queryData),
        connections: 3,
        duration: 10,
        pipelining: 1,
        timeout: 15
      });

      console.log('Query endpoint performance:', {
        requests: result.requests,
        throughput: result.throughput,
        latency: result.latency,
        errors: result.errors
      });

      expect(result.errors).to.be.lessThan(3);
      expect(result.latency.p99).to.be.lessThan(5000); // 99th percentile < 5s
    });
  });

  describe('Memory Usage', () => {
    it('should not have memory leaks under load', async () => {
      const initialMemory = process.memoryUsage();
      
      // Run load test
      await autocannon({
        url: `${baseURL}/health`,
        connections: 20,
        duration: 30,
        pipelining: 1,
        timeout: 10
      });

      const finalMemory = process.memoryUsage();
      const memoryIncrease = finalMemory.heapUsed - initialMemory.heapUsed;
      
      console.log('Memory usage:', {
        initial: initialMemory.heapUsed,
        final: finalMemory.heapUsed,
        increase: memoryIncrease
      });

      // Memory increase should be reasonable (less than 50MB)
      expect(memoryIncrease).to.be.lessThan(50 * 1024 * 1024);
    });
  });

  describe('Concurrent Service Tests', () => {
    it('should handle multiple service types simultaneously', async () => {
      const endpoints = [
        { url: `${baseURL}/health`, method: 'GET' },
        { url: `${baseURL}/api/v1/metrics`, method: 'POST', body: JSON.stringify({
          service: 'concurrent-test',
          metric: 'test_metric',
          value: 100,
          timestamp: new Date().toISOString()
        })},
        { url: `${baseURL}/api/v1/traces`, method: 'POST', body: JSON.stringify({
          trace_id: 'concurrent-test-trace',
          span_id: 'concurrent-test-span',
          operation_name: 'test_operation',
          start_time: new Date().toISOString(),
          duration: 100,
          service_name: 'concurrent-test'
        })}
      ];

      const results = await Promise.all(
        endpoints.map(endpoint => 
          autocannon({
            url: endpoint.url,
            method: endpoint.method,
            headers: endpoint.method === 'POST' ? { 'content-type': 'application/json' } : {},
            body: endpoint.body,
            connections: 3,
            duration: 10,
            pipelining: 1,
            timeout: 10
          })
        )
      );

      results.forEach((result, index) => {
        console.log(`Endpoint ${index + 1} performance:`, {
          requests: result.requests,
          throughput: result.throughput,
          latency: result.latency,
          errors: result.errors
        });

        expect(result.errors).to.be.lessThan(5);
        expect(result.latency.p95).to.be.lessThan(3000);
      });
    });
  });
}); 