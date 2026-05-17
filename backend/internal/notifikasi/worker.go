package notifikasi

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type WorkerConfig struct {
	Interval    time.Duration
	BatchSize   int
	Throttle    time.Duration
	StaleAfter  time.Duration
	RetryDelay  time.Duration
}

func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Interval:   30 * time.Second,
		BatchSize:  10,
		Throttle:   500 * time.Millisecond,
		StaleAfter: 5 * time.Minute,
		RetryDelay: 30 * time.Second,
	}
}

type Worker struct {
	repo     *Repository
	registry *Registry
	cfg      WorkerConfig
	logger   *slog.Logger
}

func NewWorker(repo *Repository, registry *Registry, cfg WorkerConfig) *Worker {
	return &Worker{
		repo:     repo,
		registry: registry,
		cfg:      cfg,
		logger:   slog.Default().With("component", "notification-worker"),
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("worker started",
		"interval", w.cfg.Interval,
		"batch_size", w.cfg.BatchSize,
		"stale_after", w.cfg.StaleAfter,
		"providers", w.registry.Types(),
	)

	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker stopped")
			return
		case <-ticker.C:
			w.releaseStale()
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) releaseStale() {
	released, err := w.repo.ReleaseStale(w.cfg.StaleAfter)
	if err != nil {
		w.logger.Error("gagal release stale items", "error", err)
		return
	}
	if released > 0 {
		w.logger.Warn("released stale processing items", "count", released)
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	items, err := w.repo.ClaimPending(w.cfg.BatchSize)
	if err != nil {
		w.logger.Error("gagal claim pending items", "error", err)
		return
	}

	if len(items) == 0 {
		return
	}

	w.logger.Info("claimed batch", "count", len(items))

	for i := range items {
		select {
		case <-ctx.Done():
			w.logger.Warn("context cancelled, releasing remaining items")
			for j := i; j < len(items); j++ {
				w.repo.UpdateStatus(items[j].ID, "pending", "worker shutdown", items[j].RetryCount)
			}
			return
		default:
		}

		w.processItem(&items[i])

		if w.cfg.Throttle > 0 && i < len(items)-1 {
			time.Sleep(w.cfg.Throttle)
		}
	}
}

func (w *Worker) processItem(n *Notifikasi) {
	provider, err := w.registry.Get(n.Tipe)
	if err != nil {
		w.markFailed(n, fmt.Sprintf("provider error: %v", err))
		return
	}

	result := provider.Send(n)

	if result.Success {
		if err := w.repo.UpdateStatus(n.ID, "sent", "", n.RetryCount); err != nil {
			w.logger.Error("gagal update status sent", "id", n.ID, "error", err)
		} else {
			w.logger.Info("notifikasi terkirim", "id", n.ID, "tipe", n.Tipe, "penerima", n.Penerima)
		}
		return
	}

	errMsg := ""
	if result.Error != nil {
		errMsg = result.Error.Error()
	}

	newRetryCount := n.RetryCount + 1
	if newRetryCount >= n.MaxRetries {
		w.markFailed(n, fmt.Sprintf("max retries reached: %s", errMsg))
		return
	}

	if err := w.repo.UpdateStatus(n.ID, "pending", errMsg, newRetryCount); err != nil {
		w.logger.Error("gagal update retry", "id", n.ID, "error", err)
	} else {
		w.logger.Warn("notifikasi gagal, akan di-retry",
			"id", n.ID, "tipe", n.Tipe, "retry", newRetryCount, "error", errMsg)
	}
}

func (w *Worker) markFailed(n *Notifikasi, errMsg string) {
	if err := w.repo.UpdateStatus(n.ID, "failed", errMsg, n.RetryCount); err != nil {
		w.logger.Error("gagal update status failed", "id", n.ID, "error", err)
	} else {
		w.logger.Warn("notifikasi gagal final", "id", n.ID, "tipe", n.Tipe, "error", errMsg)
	}
}
