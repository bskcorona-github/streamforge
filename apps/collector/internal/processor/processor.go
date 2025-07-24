package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bskcorona-github/streamforge/apps/collector/internal/config"
	"go.uber.org/zap"
)

// Processor represents the data processor
type Processor struct {
	config   *config.Config
	logger   *zap.Logger
	workers  []*Worker
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// Worker represents a processing worker
type Worker struct {
	id       int
	config   *config.Config
	logger   *zap.Logger
	workChan chan []interface{}
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// New creates a new processor instance
func New(cfg *config.Config, logger *zap.Logger) (*Processor, error) {
	workers := make([]*Worker, cfg.Collection.MaxWorkers)
	for i := 0; i < cfg.Collection.MaxWorkers; i++ {
		workers[i] = &Worker{
			id:       i,
			config:   cfg,
			logger:   logger.With(zap.Int("worker_id", i)),
			workChan: make(chan []interface{}, cfg.Collection.BufferSize),
			stopChan: make(chan struct{}),
		}
	}

	return &Processor{
		config:   cfg,
		logger:   logger,
		workers:  workers,
		stopChan: make(chan struct{}),
	}, nil
}

// Start starts the processor
func (p *Processor) Start(ctx context.Context) error {
	p.logger.Info("Starting processor",
		zap.Int("workers", len(p.workers)),
		zap.Int("buffer_size", p.config.Collection.BufferSize),
	)

	// Start workers
	for _, worker := range p.workers {
		if err := worker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start worker %d: %w", worker.id, err)
		}
	}

	p.logger.Info("Processor started successfully")
	return nil
}

// Shutdown gracefully shuts down the processor
func (p *Processor) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down processor")

	// Signal stop
	close(p.stopChan)

	// Shutdown workers
	var wg sync.WaitGroup
	for _, worker := range p.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			if err := w.Shutdown(ctx); err != nil {
				p.logger.Error("Error shutting down worker", zap.Int("worker_id", w.id), zap.Error(err))
			}
		}(worker)
	}

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("All workers stopped")
	case <-ctx.Done():
		p.logger.Warn("Shutdown timeout, forcing stop")
	}

	p.logger.Info("Processor shutdown complete")
	return nil
}

// Process processes the given data
func (p *Processor) Process(ctx context.Context, data []interface{}) ([]interface{}, error) {
	start := time.Now()
	p.logger.Debug("Processing data", zap.Int("count", len(data)))

	// Split data into batches
	batches := p.splitIntoBatches(data)
	results := make([][]interface{}, len(batches))

	// Process batches in parallel
	var wg sync.WaitGroup
	for i, batch := range batches {
		wg.Add(1)
		go func(index int, batchData []interface{}) {
			defer wg.Done()
			processed, err := p.processBatch(ctx, batchData)
			if err != nil {
				p.logger.Error("Error processing batch", zap.Int("batch_index", index), zap.Error(err))
				return
			}
			results[index] = processed
		}(i, batch)
	}

	wg.Wait()

	// Combine results
	var processed []interface{}
	for _, result := range results {
		processed = append(processed, result...)
	}

	p.logger.Debug("Data processing completed",
		zap.Duration("duration", time.Since(start)),
		zap.Int("input_count", len(data)),
		zap.Int("output_count", len(processed)),
	)

	return processed, nil
}

// splitIntoBatches splits data into batches
func (p *Processor) splitIntoBatches(data []interface{}) [][]interface{} {
	batchSize := p.config.Collection.BatchSize
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

// processBatch processes a single batch of data
func (p *Processor) processBatch(ctx context.Context, batch []interface{}) ([]interface{}, error) {
	processed := make([]interface{}, 0, len(batch))

	for _, item := range batch {
		// Add processing metadata
		processedItem := p.addMetadata(item)
		
		// Apply transformations
		transformed, err := p.transform(processedItem)
		if err != nil {
			p.logger.Error("Error transforming item", zap.Error(err))
			continue
		}

		processed = append(processed, transformed)
	}

	return processed, nil
}

// addMetadata adds processing metadata to the item
func (p *Processor) addMetadata(item interface{}) map[string]interface{} {
	metadata := map[string]interface{}{
		"processed_at": time.Now().Unix(),
		"processor_id": "streamforge-collector",
		"version":      "0.1.0",
	}

	if itemMap, ok := item.(map[string]interface{}); ok {
		itemMap["metadata"] = metadata
		return itemMap
	}

	return map[string]interface{}{
		"data":     item,
		"metadata": metadata,
	}
}

// transform applies transformations to the item
func (p *Processor) transform(item interface{}) (interface{}, error) {
	// TODO: Implement actual transformations
	// For now, just return the item as-is
	return item, nil
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info("Starting worker")

	w.wg.Add(1)
	go w.workLoop(ctx)

	w.logger.Info("Worker started successfully")
	return nil
}

// Shutdown gracefully shuts down the worker
func (w *Worker) Shutdown(ctx context.Context) error {
	w.logger.Info("Shutting down worker")

	// Signal stop
	close(w.stopChan)

	// Wait for work loop to finish
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("Worker stopped")
	case <-ctx.Done():
		w.logger.Warn("Shutdown timeout, forcing stop")
	}

	return nil
}

// workLoop runs the worker's main processing loop
func (w *Worker) workLoop(ctx context.Context) {
	defer w.wg.Done()

	w.logger.Info("Worker loop started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker context cancelled")
			return
		case <-w.stopChan:
			w.logger.Info("Worker stop signal received")
			return
		case work := <-w.workChan:
			if err := w.processWork(ctx, work); err != nil {
				w.logger.Error("Error processing work", zap.Error(err))
			}
		}
	}
}

// processWork processes a work item
func (w *Worker) processWork(ctx context.Context, work []interface{}) error {
	start := time.Now()
	w.logger.Debug("Processing work", zap.Int("count", len(work)))

	// TODO: Implement actual work processing
	// For now, just simulate processing time
	time.Sleep(10 * time.Millisecond)

	w.logger.Debug("Work processing completed",
		zap.Duration("duration", time.Since(start)),
		zap.Int("count", len(work)),
	)

	return nil
} 