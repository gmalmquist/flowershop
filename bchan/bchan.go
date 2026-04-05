package bchan

import (
  "context"
  "log"
  "time"
)

const epsilon time.Duration = 10 * time.Millisecond

type BucketedChannel[E comparable] struct {
  ctx context.Context
  ingress chan E
  egress chan []E
  clock chan E
  pending map[E]int
  nextTick time.Time
  period time.Duration
}

func New[E comparable](
  ctx context.Context,
  period time.Duration,
) *BucketedChannel[E] {
  b := &BucketedChannel[E]{
    ctx: ctx,
    egress: make(chan []E, 1),
    ingress: make(chan E, 1),
    clock: make(chan E, 1),
    pending: make(map[E]int),
    period: period,
  }
  go b.listen()
  return b
}

func (b *BucketedChannel[E]) Send(items ...E) {
  for _, item := range items {
    b.ingress <- item
  }
}

func (b *BucketedChannel[E]) Receive() []E {
  return <-b.egress
}

func (b *BucketedChannel[E]) Receiver() chan []E {
  return b.egress
}

func (b *BucketedChannel[E]) listen() {
  log.Print("Bucketed channel listener started.")
  for {
    select {
      case <-b.ctx.Done():
        return
      case item := <-b.ingress:
        b.accept(item)
      case _ = <-b.clock:
        arr := []E{ }
        for k, v := range b.pending {
          if v > 0 {
            arr = append(arr, k)
          }
        }
        if len(arr) == 0 {
          continue
        }
        b.egress <- arr
        b.pending = make(map[E]int)
    }
  }
}

func (b *BucketedChannel[E]) accept(item E) {
  b.increment(item)
  tick := b.nextTick
  now := time.Now()
  if tick.IsZero() || now.Add(epsilon).After(tick) {
    // need to schedule a tick
    b.nextTick = now.Add(b.period)
    go b.tickAfter(item)

  }
}

func (b *BucketedChannel[E]) increment(item E) {
  b.pending[item] = b.pending[item] + 1
}

func (b *BucketedChannel[E]) tickAfter(item E) {
  time.Sleep(b.period)
  b.clock <- item
}

