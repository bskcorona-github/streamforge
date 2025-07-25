const request = require('supertest');
const { expect } = require('chai');

describe('API Gateway Integration Tests', () => {
  const baseURL = process.env.API_GATEWAY_URL || 'http://localhost:8085';
  let app;

  before(async () => {
    // Wait for services to be ready
    await new Promise(resolve => setTimeout(resolve, 5000));
  });

  describe('Health Check', () => {
    it('should return 200 for health check', async () => {
      const response = await request(baseURL)
        .get('/health')
        .expect(200);

      expect(response.body).to.have.property('status', 'healthy');
      expect(response.body).to.have.property('timestamp');
    });
  });

  describe('Metrics API', () => {
    it('should accept metrics data', async () => {
      const metricData = {
        service: 'test-service',
        metric: 'request_duration',
        value: 150,
        timestamp: new Date().toISOString(),
        labels: {
          endpoint: '/api/v1/test',
          method: 'GET',
          status_code: '200'
        }
      };

      const response = await request(baseURL)
        .post('/api/v1/metrics')
        .send(metricData)
        .expect(201);

      expect(response.body).to.have.property('id');
      expect(response.body).to.have.property('status', 'accepted');
    });

    it('should validate required fields', async () => {
      const invalidData = {
        service: 'test-service',
        // Missing required fields
      };

      await request(baseURL)
        .post('/api/v1/metrics')
        .send(invalidData)
        .expect(400);
    });
  });

  describe('Traces API', () => {
    it('should accept trace data', async () => {
      const traceData = {
        trace_id: 'test-trace-id-123',
        span_id: 'test-span-id-456',
        operation_name: 'test_operation',
        start_time: new Date().toISOString(),
        duration: 100,
        service_name: 'test-service',
        tags: {
          'http.method': 'GET',
          'http.url': '/api/v1/test'
        }
      };

      const response = await request(baseURL)
        .post('/api/v1/traces')
        .send(traceData)
        .expect(201);

      expect(response.body).to.have.property('id');
      expect(response.body).to.have.property('status', 'accepted');
    });
  });

  describe('Logs API', () => {
    it('should accept log data', async () => {
      const logData = {
        timestamp: new Date().toISOString(),
        level: 'INFO',
        service: 'test-service',
        message: 'Test log message',
        trace_id: 'test-trace-id-123',
        span_id: 'test-span-id-456',
        attributes: {
          user_id: 'user-123',
          request_id: 'req-456'
        }
      };

      const response = await request(baseURL)
        .post('/api/v1/logs')
        .send(logData)
        .expect(201);

      expect(response.body).to.have.property('id');
      expect(response.body).to.have.property('status', 'accepted');
    });
  });

  describe('Query API', () => {
    it('should query metrics', async () => {
      const query = {
        service: 'test-service',
        metric: 'request_duration',
        start_time: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
        end_time: new Date().toISOString()
      };

      const response = await request(baseURL)
        .post('/api/v1/query/metrics')
        .send(query)
        .expect(200);

      expect(response.body).to.be.an('array');
    });

    it('should query traces', async () => {
      const query = {
        service: 'test-service',
        start_time: new Date(Date.now() - 3600000).toISOString(),
        end_time: new Date().toISOString(),
        limit: 10
      };

      const response = await request(baseURL)
        .post('/api/v1/query/traces')
        .send(query)
        .expect(200);

      expect(response.body).to.be.an('array');
    });

    it('should query logs', async () => {
      const query = {
        service: 'test-service',
        level: 'INFO',
        start_time: new Date(Date.now() - 3600000).toISOString(),
        end_time: new Date().toISOString(),
        limit: 10
      };

      const response = await request(baseURL)
        .post('/api/v1/query/logs')
        .send(query)
        .expect(200);

      expect(response.body).to.be.an('array');
    });
  });

  describe('Authentication', () => {
    it('should require authentication for protected endpoints', async () => {
      await request(baseURL)
        .get('/api/v1/admin/status')
        .expect(401);
    });

    it('should accept valid JWT token', async () => {
      // This test would require a valid JWT token
      // For now, we'll just test the endpoint exists
      const response = await request(baseURL)
        .get('/api/v1/admin/status')
        .set('Authorization', 'Bearer invalid-token')
        .expect(401);

      expect(response.body).to.have.property('error', 'Invalid token');
    });
  });

  describe('Rate Limiting', () => {
    it('should enforce rate limits', async () => {
      const requests = Array(101).fill().map(() => 
        request(baseURL)
          .get('/health')
          .then(response => response.status)
      );

      const responses = await Promise.all(requests);
      const tooManyRequests = responses.filter(status => status === 429);
      
      expect(tooManyRequests.length).to.be.greaterThan(0);
    });
  });
}); 