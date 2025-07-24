package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bskcorona-github/streamforge/apps/collector/internal/config"
	"github.com/bskcorona-github/streamforge/apps/collector/internal/processor"
	"github.com/bskcorona-github/streamforge/apps/collector/internal/sender"
	"go.uber.org/zap"
)

// Collector represents the main data collector
type Collector struct {
	config    *config.Config
	logger    *zap.Logger
	processor *processor.Processor
	sender    *sender.Sender
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// New creates a new collector instance
func New(cfg *config.Config, logger *zap.Logger) (*Collector, error) {
	proc, err := processor.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	snd, err := sender.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	return &Collector{
		config:    cfg,
		logger:    logger,
		processor: proc,
		sender:    snd,
		stopChan:  make(chan struct{}),
	}, nil
}

// Start starts the collector
func (c *Collector) Start(ctx context.Context) error {
	c.logger.Info("Starting StreamForge Collector")

	// Start processor
	if err := c.processor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start processor: %w", err)
	}

	// Start sender
	if err := c.sender.Start(ctx); err != nil {
		return fmt.Errorf("failed to start sender: %w", err)
	}

	// Start collection loop
	c.wg.Add(1)
	go c.collectionLoop(ctx)

	c.logger.Info("StreamForge Collector started successfully")
	return nil
}

// Shutdown gracefully shuts down the collector
func (c *Collector) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down StreamForge Collector")

	// Signal stop
	close(c.stopChan)

	// Wait for collection loop to finish
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.logger.Info("Collection loop stopped")
	case <-ctx.Done():
		c.logger.Warn("Shutdown timeout, forcing stop")
	}

	// Shutdown processor
	if err := c.processor.Shutdown(ctx); err != nil {
		c.logger.Error("Error shutting down processor", zap.Error(err))
	}

	// Shutdown sender
	if err := c.sender.Shutdown(ctx); err != nil {
		c.logger.Error("Error shutting down sender", zap.Error(err))
	}

	c.logger.Info("StreamForge Collector shutdown complete")
	return nil
}

// collectionLoop runs the main collection loop
func (c *Collector) collectionLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.Collection.Interval)
	defer ticker.Stop()

	c.logger.Info("Starting collection loop",
		zap.Duration("interval", c.config.Collection.Interval),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Collection loop context cancelled")
			return
		case <-c.stopChan:
			c.logger.Info("Collection loop stop signal received")
			return
		case <-ticker.C:
			if err := c.collectData(ctx); err != nil {
				c.logger.Error("Error collecting data", zap.Error(err))
			}
		}
	}
}

// collectData performs a single data collection cycle
func (c *Collector) collectData(ctx context.Context) error {
	start := time.Now()
	c.logger.Debug("Starting data collection cycle")

	// Collect system metrics
	metrics, err := c.collectSystemMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect system metrics: %w", err)
	}

	// Process metrics
	processed, err := c.processor.Process(ctx, metrics)
	if err != nil {
		return fmt.Errorf("failed to process metrics: %w", err)
	}

	// Send processed data
	if err := c.sender.Send(ctx, processed); err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	c.logger.Debug("Data collection cycle completed",
		zap.Duration("duration", time.Since(start)),
		zap.Int("metrics_count", len(metrics)),
		zap.Int("processed_count", len(processed)),
	)

	return nil
}

// collectSystemMetrics collects system metrics
func (c *Collector) collectSystemMetrics(ctx context.Context) ([]interface{}, error) {
	// TODO: Implement actual system metrics collection
	// For now, return mock data
	metrics := []interface{}{
		map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"type":      "system",
			"cpu_usage": 45.2,
			"memory_usage": 67.8,
			"disk_usage": 23.1,
		},
		map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"type":      "network",
			"bytes_sent": 1024000,
			"bytes_received": 2048000,
			"connections": 150,
		},
	}

	return metrics, nil
} 