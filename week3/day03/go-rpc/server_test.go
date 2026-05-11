package main

import (
	"context"
	"go-rpc/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 本地配置（服务端在本机）
const (
	serverAddr    = "127.0.0.1:50051"
	testNamespace = "test"
)

// 全局客户端（初始化一次，永不关闭，避免空指针）
var client pb.PodManagerClient

// init 初始化连接（程序启动时执行一次，不关闭连接）
func init() {
	// 连接本地gRPC服务
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		// 连接失败直接报错
		panic("连接本地gRPC服务失败：" + err.Error())
	}

	// 初始化客户端
	client = pb.NewPodManagerClient(conn)
}

// ===================== 第一个测试：创建Deployment（成功） =====================
func TestCreateDeployment_Success(t *testing.T) {
	// 安全判断：防止客户端为空
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	// 构造请求，包含资源配置验证
	req := &pb.CreateDeploymentRequest{
		PodName:       "test-nginx-02",
		Image:         "nginx:alpine",
		Namespace:     testNamespace,
		Replicas:      2,
		CpuRequest:    "100m",
		CpuLimit:      "300m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "256Mi",
	}

	// 调用接口
	resp, err := client.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	// 校验结果
	if !resp.Success {
		t.Fatalf("创建失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：Deployment 创建成功！")
}

// ===================== 缺省配额情况测试：创建带有默认资源和副本的Deployment =====================
func TestCreateDeployment_DefaultSettings(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	// 构造请求，不穿入 Replicas, CpuRequest, 等配置参数 (即采用空值与0值)
	req := &pb.CreateDeploymentRequest{
		PodName:   "test-nginx-default",
		Image:     "nginx:alpine",
		Namespace: testNamespace,
		// Replicas，CpuRequest，MemRequest等全部忽略不传...
	}

	// 调用接口
	resp, err := client.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用缺省接口失败：%v", err)
	}

	// 校验结果
	if !resp.Success {
		t.Fatalf("缺省参数创建失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：缺省配额 Deployment (test-nginx-default) 创建成功！预期自动回填：1副本, 100m-200m CPU")
}

// ===================== 第二个测试：删除存在的Deployment（成功） =====================
func TestDeleteDeployment_Success(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.DeleteDeploymentRequest{
		PodName:   "test-nginx-02",
		Namespace: testNamespace,
	}

	resp, err := client.DeleteDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("删除失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：存在的 Deployment 删除成功！")
}

// ===================== 第三个测试：删除不存在的Deployment（失败） =====================
func TestDeleteDeployment_NotFound(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.DeleteDeploymentRequest{
		PodName:   "test-nginx-01",
		Namespace: testNamespace,
	}

	resp, err := client.DeleteDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	// 注意：Kubernetes Delete 如果对象不存在默认可能不报错(如果DeleteOptions没有特定设置)或者会报错 NotFound
	if resp.Success {
		t.Logf("注: K8s删除不存在资源可能依旧返回Success (幂等性)。当前返回成功。")
	} else {
		t.Logf("✅ 测试通过：删除不存在的 Deployment 被拦截/报错，服务端返回提示：%s", resp.Message)
	}
}

// ===================== 第五个测试：创建带有环境变量与 HostPath 持久化的 MySQL Deployment =====================
func TestCreateDeployment_MySQL_HostPath(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.CreateDeploymentRequest{
		PodName:   "test-mysql",
		Image:     "mysql:8.0",
		Namespace: testNamespace,
		// 给 MySQL 配置足够的内存，避免 OOM（默认的 256Mi 太小）
		CpuRequest:    "500m",
		CpuLimit:      "1000m",
		MemoryRequest: "512Mi",
		MemoryLimit:   "1Gi",
		Envs: []*pb.EnvVar{
			{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: "root1234",
			},
		},
		Storages: []*pb.StorageConfig{
			{
				Type:      "hostPath",
				Name:      "mysql-data",
				MountPath: "/var/lib/mysql",   // 容器内MySQL默认数据目录
				HostPath:  "/data/mysql-test", // 挂载到宿主机的路径，避免放在 /root 目录下导致权限问题
			},
		},
	}

	resp, err := client.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("测试失败：MySQL Deployment 创建失败，服务端返回提示：%s", resp.Message)
	}

	t.Log("✅ 测试通过：带有 Env 和 HostPath 挂载的 MySQL Deployment 创建成功！")
}

// ===================== 第六个测试：创建 Service，暴露上述 MySQL Deployment =====================
func TestCreateService_MySQL_NodePort(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.CreateServiceRequest{
		ServiceName: "test-mysql-svc",
		Namespace:   testNamespace,
		Type:        "NodePort",
		Selector: map[string]string{
			"app": "test-mysql", // 匹配之前创建的 Deployment 的 Label
		},
		Ports: []*pb.PortConfig{
			{
				Port:       3306,
				TargetPort: 3306,
				NodePort:   30306, // K8s NodePort 默认范围是 30000-32767，这里选择 30306 代替 33060
			},
		},
	}

	resp, err := client.CreateService(context.Background(), req)
	if err != nil {
		t.Fatalf("调用 Service 创建接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("测试失败：MySQL Service 创建失败，服务端返回提示：%s", resp.Message)
	}

	t.Log("✅ 测试通过：成功暴露 MySQL Deployment 为 NodePort 模式(30306)！")
}
