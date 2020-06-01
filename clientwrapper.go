package microcore

import (
	"context"
	"errors"
	"github.com/juju/ratelimit"
	"github.com/micro/go-micro/v2/client"
	"time"
)

type ClientExFunc func() error

/**
 * 对原始客户端对象的包装
 */
type clientWrapper struct {
	client.Client
	before ClientExFunc
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if err := c.before(); err != nil {
		return err
	}
	return c.Client.Call(ctx, req, rsp, opts...)
}

/**
 * 对原始客户端进行包装，使其具备限流能力
 * 参数：
 * fillIntervalMs 向令牌桶添加令牌的周期，以毫秒为单位
 * bucketCapicty  令牌桶中的容量
 * quantumAdd     每次添加多少令牌到桶里
 * wait 当令牌耗尽时是否等待
 */
func newRateLimitWrapper(fillIntervalMs int, bucketCapicty, quantumAdd int64, wait bool) client.Wrapper {
	bucket := ratelimit.NewBucketWithQuantum(time.Millisecond * time.Duration(fillIntervalMs), bucketCapicty, quantumAdd)
	fn := func() error {
		if wait {
			time.Sleep(bucket.Take(1))
		} else if bucket.TakeAvailable(1) == 0 {
			return errors.New("too many requests")
		}
		return nil
	}
	return func(c client.Client) client.Client {
		return &clientWrapper{c, fn}
	}
}
