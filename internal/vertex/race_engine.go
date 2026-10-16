package vertex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"AnonymousVertex/internal/config"
	"AnonymousVertex/internal/nodes"
)

// raceConfig 是 RunRace 的可配置策略。
type raceConfig[T any] struct {
	noCancelOnSuccess bool
	// isWinningResult 决定某个成功结果是否可"立即胜出"。
	// nil 表示首个无错误结果立即胜出（流式默认）。
	isWinningResult func(val T) bool
	// finalizeCollected 在收集多个非胜出结果后调用，选出最佳结果。
	finalizeCollected func([]raceResult[T]) (T, error)
}

type RaceOption[T any] func(*raceConfig[T])

// WithNoCancelOnSuccess 使胜出路径上不 cancel 胜出节点的 ctx（流式胜出后数据流需继续传输）。
func WithNoCancelOnSuccess[T any]() RaceOption[T] {
	return func(cfg *raceConfig[T]) {
		cfg.noCancelOnSuccess = true
	}
}

// WithWinningCheck 注入成功判定：fn 返回 true 时该结果立即胜出；
// fn 返回 false 时结果被收集，所有候选结束后通过 WithCollectedFinalizer 选最佳结果。
func WithWinningCheck[T any](fn func(T) bool) RaceOption[T] {
	return func(cfg *raceConfig[T]) {
		cfg.isWinningResult = fn
	}
}

// WithCollectedFinalizer 注入最终结果选择函数，在收集多个非胜出结果后调用。
func WithCollectedFinalizer[T any](fn func([]raceResult[T]) (T, error)) RaceOption[T] {
	return func(cfg *raceConfig[T]) {
		cfg.finalizeCollected = fn
	}
}

type raceResult[T any] struct {
	uri string
	val T
	err error
}

// errorPriority 返回错误的优先级数值（越小优先级越高）。
// 可重试错误优先于不可重试错误，避免临时网络波动掩盖节点故障。
func errorPriority(err error) int {
	var ve *VertexError
	if errors.As(err, &ve) {
		if ve.IsRetryable() {
			if ve.Kind == "ratelimit" || ve.Code == 429 {
				return 1
			}
			return 2
		}
		if ve.IsGlobalHardError() {
			return 3
		}
		return 4
	}
	return 5
}

// pickBestError 从多个错误中挑选优先级最高（数值最小）的一个返回。
func pickBestError(errs []error) error {
	if len(errs) == 0 {
		return fmt.Errorf("all nodes failed")
	}
	best := errs[0]
	bestPrio := errorPriority(best)
	for _, e := range errs[1:] {
		if p := errorPriority(e); p < bestPrio {
			best = e
			bestPrio = p
		}
	}
	return best
}

// RunRace 统一权威滑动窗口并发竞速引擎。
//
// 保持配置上限的候选并发在途，一旦某节点失败或产出非胜出结果，立即从候选池补充新节点，
// 胜出节点立即返回，其余在途节点立即通过 context 取消，杜绝协程和连接泄露。
func RunRace[T any](
	ctx context.Context,
	cfg config.ConfigProvider,
	run func(ctx context.Context, proxyURI string) (T, error),
	opts ...RaceOption[T],
) (T, error) {
	var rc raceConfig[T]
	for _, o := range opts {
		o(&rc)
	}

	var zero T
	stickyPool := nodes.GetStickyPool()

	if !cfg.ParallelPoolEnabled() {
		proxy := cfg.ActiveNodeURI()
		if proxy == "" {
			proxy = cfg.ProxyURL()
		}
		return run(ctx, proxy)
	}

	w := cfg.ParallelPoolSize()
	if w <= 0 {
		w = 1
	}
	retries := cfg.MaxRetries()
	if retries < 0 {
		retries = 0
	}
	if retries > 100 {
		retries = 100
	}
	budget := (retries + 1) * w

	type result struct {
		uri string
		val T
		err error
	}
	results := make(chan result, w)
	active := make(map[string]context.CancelFunc)
	var mu sync.Mutex
	launched := 0
	var failures []error
	var collected []raceResult[T]
	timeout := cfg.RaceTimeout()

	launch := func(uri string) {
		candCtx, cancel := context.WithCancel(ctx)
		mu.Lock()
		active[uri] = cancel
		mu.Unlock()
		launched++
		go func() {
			defer func() {
				if r := recover(); r != nil {
					results <- result{uri: uri, err: NewInternalError(fmt.Sprintf("candidate %s panic: %v", nodes.GetNodeName(uri), r))}
				}
			}()

			var timer *time.Timer
			if timeout > 0 {
				timer = time.AfterFunc(time.Duration(timeout)*time.Second, func() {
					cancel()
				})
			}

			v, err := run(candCtx, uri)

			if timer != nil {
				timer.Stop()
			}

			if err != nil && candCtx.Err() != nil && ctx.Err() == nil {
				err = NewUnavailableError(fmt.Sprintf("节点 %s 竞速超时（%d 秒），已淘汰", nodes.GetNodeName(uri), timeout))
			}

			results <- result{uri: uri, val: v, err: err}
		}()
	}

	fill := func() {
		for len(active) < w && launched < budget {
			cands := nodes.SelectForParallel(w, cfg.ParallelNodeTopK(), cfg.DebugMode(), cfg.StickyNodePriority())
			picked := ""
			for _, c := range cands {
				if _, ok := active[c.RawURI]; !ok && strings.TrimSpace(c.RawURI) != "" {
					picked = c.RawURI
					break
				}
			}
			if picked == "" && len(active) == 0 {
				for _, c := range nodes.LoadNodes() {
					if !c.Disabled {
						if _, ok := active[c.RawURI]; !ok && strings.TrimSpace(c.RawURI) != "" {
							picked = c.RawURI
							break
						}
					}
				}
			}
			if picked == "" {
				return
			}
			launch(picked)
		}
	}

	fill()
	if len(active) == 0 && launched == 0 {
		proxy := cfg.ActiveNodeURI()
		if proxy == "" {
			proxy = cfg.ProxyURL()
		}
		return run(ctx, proxy)
	}

	for {
		if len(active) == 0 {
			if len(collected) > 0 {
				if rc.finalizeCollected != nil {
					return rc.finalizeCollected(collected)
				}
				return collected[0].val, nil
			}
			if launched >= budget {
				if len(failures) > 0 {
					return zero, pickBestError(failures)
				}
				return zero, fmt.Errorf("all nodes failed")
			}
			fill()
			if len(active) == 0 {
				if len(failures) > 0 {
					return zero, pickBestError(failures)
				}
				return zero, fmt.Errorf("all nodes failed")
			}
		}

		select {
		case <-ctx.Done():
			mu.Lock()
			for _, cancel := range active {
				cancel()
			}
			mu.Unlock()
			return zero, ctx.Err()

		case res := <-results:
			mu.Lock()
			winnerCancel := active[res.uri]
			delete(active, res.uri)
			mu.Unlock()

			if res.err == nil {
				nodes.RecordTest(res.uri, true, 50, "")
				stickyPool.Add(res.uri)
				if rc.isWinningResult == nil || rc.isWinningResult(res.val) {
					name := nodes.GetNodeName(res.uri)
					log.Printf("[Racing] 竞速胜出节点: %s", name)

					// 胜出：取消其他所有正在运行的候选
					mu.Lock()
					for _, cancel := range active {
						cancel()
					}
					active = nil
					mu.Unlock()

					if !rc.noCancelOnSuccess && winnerCancel != nil {
						winnerCancel()
					}
					return res.val, nil
				}
				if winnerCancel != nil {
					winnerCancel()
				}
				collected = append(collected, raceResult[T]{uri: res.uri, val: res.val})
			} else {
				if winnerCancel != nil {
					winnerCancel()
				}
				if !errors.Is(res.err, context.Canceled) {
					failures = append(failures, res.err)
					ve := asVertexError(res.err)
					if ve != nil && ve.IsGlobalHardError() {
						log.Printf("[Racing] 节点 %s 报告全局硬错误: %v, 快速熔断竞速", nodes.GetNodeName(res.uri), ve)
						mu.Lock()
						for _, cancel := range active {
							cancel()
						}
						active = nil
						mu.Unlock()
						return zero, res.err
					}
					if ve != nil && ve.Kind == "ratelimit" {
						log.Printf("[Racing] 节点 %s 触发 429，时间片轮转切入其他节点！", nodes.GetNodeName(res.uri))
						nodes.RecordRateLimit(res.uri, 2)
						stickyPool.Evict(res.uri)
					} else {
						nodes.RecordTest(res.uri, false, 0, res.err.Error())
						stickyPool.Evict(res.uri)
					}
				}
			}
			fill()
		}
	}
}
