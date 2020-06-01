package microcore

import (
	"context"
	"errors"
	"github.com/juju/ratelimit"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"microcore/security"
	"time"
)

///////////////////////////////////////////////////////////////////////
//
// 一些通用的选项，例如限流，服务熔断，鉴权等等
//
///////////////////////////////////////////////////////////////////////



/**
 * 带上限流功能，针对整个服务的所有接口，用一个令牌桶控制整个服务的访问量
 * 参数：
 * fillIntervalMs 向令牌桶添加令牌的周期，以毫秒为单位
 * bucketCapicty  令牌桶中的容量
 * quantumAdd     每次添加多少令牌到桶里
 * wait           当令牌耗尽时是否等待
 */
func WithRateLimit(fillIntervalMs int, bucketCapicty, quantumAdd int64, wait bool) micro.Option {
	return micro.WrapClient(newRateLimitWrapper(fillIntervalMs, bucketCapicty, quantumAdd, wait))
}

/**
 * 带上限流功能，针对服务的某个接口，每个接口用一个令牌桶来控制访问量
 * 参数：
 * bucket 控制接口访问的令牌桶
 * wait   当令牌耗尽时是否等待
 */
func WithRateLimitCall(bucket *ratelimit.Bucket, wait bool) client.CallOption {
	wrapper := func(f client.CallFunc) client.CallFunc {
		if wait {
			// 一直等待到令牌桶中有令牌为止
			time.Sleep(bucket.Take(1))
		} else if bucket.TakeAvailable(1) == 0 {
			// 没有拿到令牌，又不等待，直接返回错误给客户端
			return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
				return errors.New("too many requests")
			}
		}
		return f
	}
	return client.WithCallWrapper(wrapper)
}

/**
 * 对接口调用进行鉴权，客户端传入的token校验无误再去调用服务端
 */
func CheckSign(sign string, checkSign security.CheckSign) client.CallOption {
	wrapper := func(f client.CallFunc) client.CallFunc {
		if !checkSign(sign) {
			return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
				return errors.New("invalid sign")
			}
		} else {
			// 没有拿到令牌，又不等待，直接返回错误给客户端
			return f
		}
	}
	return client.WithCallWrapper(wrapper)
}


