package microcore

import (
	reg "framework/registry"
	"github.com/micro/go-micro/v2"
	"reflect"
)

/**
 * 服务的路径
 */
type Service struct {

	Name string

	HandlerRegisterFunc interface{}

	NewServiceFunc interface{}
}

/**
 * 创建并运行客户端（代理）
 * 参数说明
 * gatewayNameSpace 与api网关对应的namespace，启动网关时通过--namespace设置
 * hdlr             接口实现
 * ops              选项参数
 */
func (p Service) Client(gatewayNamespace string, hdlr interface{}, opts ...micro.Option) error {
	// 创建服务并初始化
	s := micro.NewService(
		micro.Name(gatewayNamespace + "." + p.Name), //注意这里服务的名称：namespace + service，这样网关才能找的到
		micro.Registry(reg.NewRegistryFromEnv()),
		)
	s.Init(opts...)
	
	// 注册处理器
	backEndService := "svr." + p.Name
	client := reflect.ValueOf(p.NewServiceFunc).Call([]reflect.Value{reflect.ValueOf(backEndService), reflect.ValueOf(s.Client())})
	h := reflect.ValueOf(hdlr).Elem()
	h.FieldByName("Client").Set(client[0])
	reflect.ValueOf(p.HandlerRegisterFunc).Call([]reflect.Value{
		reflect.ValueOf(s.Server()),
		reflect.ValueOf(hdlr),
	})

	// 启动服务运行
	return s.Run()
}

/**
 * 创建并运行服务端
 */
func (p Service) Server(hdlr interface{}, opts ...micro.Option) error {
	// 创建一个服务
	s := micro.NewService(
		micro.Name("svr." + p.Name),
		micro.Registry(reg.NewRegistryFromEnv()))
	// 初始化
	s.Init(opts...)
	// 注册处理器
	reflect.ValueOf(p.HandlerRegisterFunc).Call([]reflect.Value{
		reflect.ValueOf(s.Server()),
		reflect.ValueOf(hdlr),
	})
	// 启动服务运行
	return s.Run()
}
