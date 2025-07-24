package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bskcorona-github/streamforge/apps/collector/internal/config"
	"go.uber.org/zap"
)

// Sender represents the data sender
type Sender struct {
	config   *config.Config
	logger   *zap.Logger
	client   *http.Client
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// New creates a new sender instance
func New(cfg *config.Config, logger *zap.Logger) (*Sender, error) {
	client := &http.Client{
		Timeout: cfg.API.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &Sender{
		config:   cfg,
		logger:   logger,
		client:   client,
		stopChan: make(chan struct{}),
	}, nil
}

// Start starts the sender
func (s *Sender) Start(ctx context.Context) error {
	s.logger.Info("Starting sender",
		zap.String("endpoint", s.config.API.Endpoint),
		zap.Duration("timeout", s.config.API.Timeout),
	)

	s.logger.Info("Sender started successfully")
	return nil
}

// Shutdown gracefully shuts down the sender
func (s *Sender) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down sender")

	// Signal stop
	close(s.stopChan)

	// Wait for any pending operations
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Sender stopped")
	case <-ctx.Done():
		s.logger.Warn("Shutdown timeout, forcing stop")
	}

	s.logger.Info("Sender shutdown complete")
	return nil
}

// Send sends the given data
func (s *Sender) Send(ctx context.Context, data []interface{}) error {
	start := time.Now()
	s.logger.Debug("Sending data", zap.Int("count", len(data)))

	// Split data into batches if needed
	batches := s.splitIntoBatches(data)

	// Send batches in parallel
	var wg sync.WaitGroup
	errors := make(chan error, len(batches))

	for i, batch := range batches {
		wg.Add(1)
		go func(index int, batchData []interface{}) {
			defer wg.Done()
			if err := s.sendBatch(ctx, batchData); err != nil {
				errors <- fmt.Errorf("batch %d failed: %w", index, err)
			}
		}(i, batch)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		s.logger.Error("Some batches failed to send",
			zap.Int("total_batches", len(batches)),
			zap.Int("failed_batches", len(errs)),
		)
		return fmt.Errorf("failed to send %d batches", len(errs))
	}

	s.logger.Debug("Data sending completed",
		zap.Duration("duration", time.Since(start)),
		zap.Int("total_count", len(data)),
		zap.Int("batches", len(batches)),
	)

	return nil
}

// splitIntoBatches splits data into batches
func (s *Sender) splitIntoBatches(data []interface{}) [][]interface{} {
	batchSize := s.config.Collection.BatchSize
	var batches [][]interface{}

	for i := 0; i < len(data); i += batchSize {
		end := i + batchSize
		if end > len(data) {
			end = len(data)
		}
		batches = append(batches, data[i:end])
	}

	return batches
}

// sendBatch sends a single batch of data
func (s *Sender) sendBatch(ctx context.Context, batch []interface{}) error {
	// Prepare payload
	payload := map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"batch_id":  generateBatchID(),
		"data":      batch,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.API.Endpoint+"/api/v1/metrics", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "StreamForge-Collector/0.1.0")

	// Send with retries
	var lastErr error
	for attempt := 0; attempt <= s.config.API.Retries; attempt++ {
		if attempt > 0 {
			s.logger.Debug("Retrying request",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", s.config.API.Retries),
			)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			s.logger.Debug("Batch sent successfully",
				zap.Int("status_code", resp.StatusCode),
				zap.Int("batch_size", len(batch)),
			)
			return nil
		}

		lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return fmt.Errorf("failed after %d retries: %w", s.config.API.Retries, lastErr)
}

// generateBatchID generates a unique batch ID
func generateBatchID() string {
	return fmt.Sprintf("batch_%d", time.Now().UnixNano())
} 