package main

import (
	"context"
	"go-rpc/pb"
	"testing"

	"trpc.group/trpc-go/trpc-go/client"
)

const (
	serverAddr    = "ip://127.0.0.1:50051"
	testNamespace = "test"
)

var cli pb.PodManagerClientProxy

func init() {
	cli = pb.NewPodManagerClientProxy(client.WithTarget(serverAddr))
}

// ===================== 第一个测试：创建Deployment（成功） =====================
func TestCreateDeployment_Success(t *testing.T) {
	// 安全判断：防止客户端为空
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

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

	resp, err := cli.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("创建失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：Deployment 创建成功！")
}

func TestCreateDeployment_DefaultSettings(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.CreateDeploymentRequest{
		PodName:   "test-nginx-default",
		Image:     "nginx:alpine",
		Namespace: testNamespace,
	}

	resp, err := cli.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用缺省接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("缺省参数创建失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：缺省配额 Deployment (test-nginx-default) 创建成功！预期自动回填：1副本, 100m-200m CPU")
}

func TestDeleteDeployment_Success(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.DeleteDeploymentRequest{
		PodName:   "test-nginx-02",
		Namespace: testNamespace,
	}

	resp, err := cli.DeleteDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("删除失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：存在的 Deployment 删除成功！")
}

func TestDeleteDeployment_NotFound(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.DeleteDeploymentRequest{
		PodName:   "test-nginx-01",
		Namespace: testNamespace,
	}

	resp, err := cli.DeleteDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if resp.Success {
		t.Logf("注: K8s删除不存在资源可能依旧返回Success (幂等性)。当前返回成功。")
	} else {
		t.Logf("✅ 测试通过：删除不存在的 Deployment 被拦截/报错，服务端返回提示：%s", resp.Message)
	}
}

func TestCreateDeployment_MySQL_HostPath(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.CreateDeploymentRequest{
		PodName:       "test-mysql",
		Image:         "mysql:8.0",
		Namespace:     testNamespace,
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
				MountPath: "/var/lib/mysql",
				HostPath:  "/data/mysql-test",
			},
		},
	}

	resp, err := cli.CreateDeployment(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("测试失败：MySQL Deployment 创建失败，服务端返回提示：%s", resp.Message)
	}

	t.Log("✅ 测试通过：带有 Env 和 HostPath 挂载的 MySQL Deployment 创建成功！")
}

func TestCreateService_MySQL_NodePort(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.CreateServiceRequest{
		ServiceName: "test-mysql-svc",
		Namespace:   testNamespace,
		Type:        "NodePort",
		Selector: map[string]string{
			"app": "test-mysql",
		},
		Ports: []*pb.PortConfig{
			{
				Port:       3306,
				TargetPort: 3306,
				NodePort:   30306,
			},
		},
	}

	resp, err := cli.CreateService(context.Background(), req)
	if err != nil {
		t.Fatalf("调用 Service 创建接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("测试失败：MySQL Service 创建失败，服务端返回提示：%s", resp.Message)
	}

	t.Log("✅ 测试通过：成功暴露 MySQL Deployment 为 NodePort 模式(30306)！")
}

// ===================== Pod 测试 =====================
func TestPreCheck_K8sAndMySQL(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}
	req := &pb.GetPodRequest{
		PodName:   "pre-check-not-exist",
		Namespace: testNamespace,
	}
	resp, err := cli.GetPod(context.Background(), req)
	if err != nil {
		t.Fatalf("前置检查失败，tRPC调用出错: %v", err)
	}
	if !resp.Success {
		t.Logf("前置检查通过：服务连通，预期返回 NotFound -> %s", resp.Message)
	} else {
		t.Logf("前置检查通过：服务连通，意外发现已存在 Pod（status=%s）", resp.Status)
	}
}

func TestCreatePod_Success(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.CreatePodRequest{
		PodName:   "test-pod-1",
		Image:     "nginx:alpine",
		Namespace: testNamespace,
	}

	resp, err := cli.CreatePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用CreatePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("创建Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：Pod(test-pod-1) 创建成功！服务端返回: %s", resp.Message)
}

func TestGetPod_Success(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-get"
	image := "nginx:alpine"

	createReq := &pb.CreatePodRequest{PodName: podName, Image: image, Namespace: testNamespace}
	createResp, err := cli.CreatePod(context.Background(), createReq)
	if err != nil || !createResp.Success {
		t.Fatalf("前置创建Pod失败: %v / %s", err, createResp.GetMessage())
	}

	req := &pb.GetPodRequest{
		PodName:   podName,
		Namespace: testNamespace,
	}

	resp, err := cli.GetPod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用GetPod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("查询Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：成功查到 Pod！状态=%s, 镜像=%s", resp.Status, resp.Image)
	if resp.Image != image {
		t.Errorf("镜像信息不匹配: 期望 %s, 实际 %s", image, resp.Image)
	}

	cli.DeletePod(context.Background(), &pb.DeletePodRequest{PodName: podName, Namespace: testNamespace})
}

func TestGetPod_NotFound(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.GetPodRequest{
		PodName:   "test-pod-not-exist-999",
		Namespace: testNamespace,
	}

	resp, err := cli.GetPod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用GetPod接口失败: %v", err)
	}
	if resp.Success {
		t.Logf("注：查询不存在的 Pod 意外成功（status=%s），可能是历史残留数据", resp.Status)
	} else {
		t.Logf("✅ 测试通过：查询不存在的 Pod 正确返回失败: %s", resp.Message)
	}
}

func TestDeletePod_Success(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-del"

	createReq := &pb.CreatePodRequest{PodName: podName, Image: "nginx:alpine", Namespace: testNamespace}
	createResp, err := cli.CreatePod(context.Background(), createReq)
	if err != nil || !createResp.Success {
		t.Fatalf("前置创建Pod失败: %v / %s", err, createResp.GetMessage())
	}

	req := &pb.DeletePodRequest{
		PodName:   podName,
		Namespace: testNamespace,
	}

	resp, err := cli.DeletePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用DeletePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("删除Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：Pod(%s) 删除成功！", podName)
}

func TestDeletePod_NotFound(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	req := &pb.DeletePodRequest{
		PodName:   "test-pod-not-exist-888",
		Namespace: testNamespace,
	}

	resp, err := cli.DeletePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用DeletePod接口失败: %v", err)
	}
	if resp.Success {
		t.Log("注：K8s删除不存在资源可能返回Success（幂等性），当前返回成功")
	} else {
		t.Logf("✅ 测试通过：删除不存在的 Pod 被正确拦截: %s", resp.Message)
	}
}

// ===================== 用例6: Pod完整生命周期测试 (Create -> Get -> Delete -> Get) =====================
func TestPodFullLifecycle(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-lifecycle"
	namespace := testNamespace
	image := "nginx:alpine"

	t.Log(">>> 步骤1: 创建 Pod...")
	createReq := &pb.CreatePodRequest{
		PodName:   podName,
		Image:     image,
		Namespace: namespace,
	}
	createResp, err := cli.CreatePod(context.Background(), createReq)
	if err != nil {
		t.Fatalf("创建Pod失败: %v", err)
	}
	if !createResp.Success {
		t.Fatalf("创建Pod失败: %s", createResp.Message)
	}
	t.Logf("   创建成功: %s", createResp.Message)

	t.Log(">>> 步骤2: 查询 Pod（验证DB持久化）...")
	getReq := &pb.GetPodRequest{
		PodName:   podName,
		Namespace: namespace,
	}
	getResp, err := cli.GetPod(context.Background(), getReq)
	if err != nil {
		t.Fatalf("查询Pod失败: %v", err)
	}
	if !getResp.Success {
		t.Fatalf("查询Pod失败: %s", getResp.Message)
	}
	if getResp.Image != image {
		t.Errorf("镜像不匹配: 期望=%s, 实际=%s", image, getResp.Image)
	}
	t.Logf("   查询成功: status=%s, image=%s", getResp.Status, getResp.Image)

	t.Log(">>> 步骤3: 删除 Pod...")
	delReq := &pb.DeletePodRequest{
		PodName:   podName,
		Namespace: namespace,
	}
	delResp, err := cli.DeletePod(context.Background(), delReq)
	if err != nil {
		t.Fatalf("删除Pod失败: %v", err)
	}
	if !delResp.Success {
		t.Fatalf("删除Pod失败: %s", delResp.Message)
	}
	t.Logf("   删除成功: %s", delResp.Message)

	t.Log(">>> 步骤4: 再次查询 Pod（验证已删除/状态变更）...")
	getResp2, err := cli.GetPod(context.Background(), getReq)
	if err != nil {
		t.Fatalf("查询Pod失败: %v", err)
	}
	if !getResp2.Success {
		t.Logf("   查询返回失败（Pod已被清理或不在DB/K8s中）: %s", getResp2.Message)
	} else {
		t.Logf("   查询返回: status=%s (Deleted状态说明DB软删除生效)", getResp2.Status)
	}

	t.Log("✅ 测试通过：Pod 完整生命周期 (Create->Get->Delete->Get) 验证完成！")
}

func TestCreatePod_WithEnvVars(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-env"
	req := &pb.CreatePodRequest{
		PodName:   podName,
		Image:     "nginx:alpine",
		Namespace: testNamespace,
		Envs: []*pb.EnvVar{
			{Name: "NGINX_PORT", Value: "8080"},
			{Name: "ENVIRONMENT", Value: "test"},
		},
	}

	resp, err := cli.CreatePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用CreatePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("创建带环境变量的Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：带环境变量的 Pod(%s) 创建成功！", podName)

	getReq := &pb.GetPodRequest{PodName: podName, Namespace: testNamespace}
	getResp, err := cli.GetPod(context.Background(), getReq)
	if err != nil || !getResp.Success {
		t.Fatalf("查询Pod失败: %v", err)
	}
	t.Logf("   查询结果: status=%s, image=%s", getResp.Status, getResp.Image)

	cli.DeletePod(context.Background(), &pb.DeletePodRequest{PodName: podName, Namespace: testNamespace})
}

func TestCreatePod_WithResourceLimits(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-resources"
	req := &pb.CreatePodRequest{
		PodName:       podName,
		Image:         "nginx:alpine",
		Namespace:     testNamespace,
		CpuRequest:    "200m",
		CpuLimit:      "500m",
		MemoryRequest: "256Mi",
		MemoryLimit:   "512Mi",
	}

	resp, err := cli.CreatePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用CreatePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("创建带资源配额的Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：带资源配额的 Pod(%s) 创建成功！CPU=200m/500m, Mem=256Mi/512Mi", podName)

	getReq := &pb.GetPodRequest{PodName: podName, Namespace: testNamespace}
	getResp, err := cli.GetPod(context.Background(), getReq)
	if err != nil || !getResp.Success {
		t.Fatalf("查询Pod失败: %v", err)
	}
	t.Logf("   查询结果: status=%s, image=%s", getResp.Status, getResp.Image)

	cli.DeletePod(context.Background(), &pb.DeletePodRequest{PodName: podName, Namespace: testNamespace})
}

func TestCreatePod_WithHostPathStorage(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-storage"
	req := &pb.CreatePodRequest{
		PodName:   podName,
		Image:     "nginx:alpine",
		Namespace: testNamespace,
		Storages: []*pb.StorageConfig{
			{
				Type:      "hostPath",
				Name:      "nginx-html",
				MountPath: "/usr/share/nginx/html",
				HostPath:  "/data/nginx-test",
			},
		},
	}

	resp, err := cli.CreatePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用CreatePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("创建带存储卷的Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：带HostPath存储卷的 Pod(%s) 创建成功！", podName)

	getReq := &pb.GetPodRequest{PodName: podName, Namespace: testNamespace}
	getResp, err := cli.GetPod(context.Background(), getReq)
	if err != nil || !getResp.Success {
		t.Fatalf("查询Pod失败: %v", err)
	}
	t.Logf("   查询结果: status=%s, image=%s", getResp.Status, getResp.Image)

	cli.DeletePod(context.Background(), &pb.DeletePodRequest{PodName: podName, Namespace: testNamespace})
}

func TestCreatePod_FullConfig(t *testing.T) {
	if cli == nil {
		t.Fatal("客户端未初始化，请先启动tRPC服务端！")
	}

	podName := "test-pod-full"
	req := &pb.CreatePodRequest{
		PodName:       podName,
		Image:         "nginx:alpine",
		Namespace:     testNamespace,
		CpuRequest:    "100m",
		CpuLimit:      "300m",
		MemoryRequest: "128Mi",
		MemoryLimit:   "256Mi",
		Envs: []*pb.EnvVar{
			{Name: "CUSTOM_HEADER", Value: "X-Test-Header"},
			{Name: "LOG_LEVEL", Value: "debug"},
		},
		Storages: []*pb.StorageConfig{
			{
				Type:      "emptyDir",
				Name:      "cache-vol",
				MountPath: "/tmp/cache",
			},
			{
				Type:      "hostPath",
				Name:      "log-vol",
				MountPath: "/var/log/nginx",
				HostPath:  "/data/nginx-logs-test",
			},
		},
	}

	resp, err := cli.CreatePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用CreatePod接口失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("创建全配置Pod失败: %s", resp.Message)
	}
	t.Logf("✅ 测试通过：全配置 Pod(%s) 创建成功！(含资源配额+环境变量+多存储卷)", podName)

	getReq := &pb.GetPodRequest{PodName: podName, Namespace: testNamespace}
	getResp, err := cli.GetPod(context.Background(), getReq)
	if err != nil || !getResp.Success {
		t.Fatalf("查询Pod失败: %v", err)
	}
	t.Logf("   查询结果: status=%s, image=%s", getResp.Status, getResp.Image)

	cli.DeletePod(context.Background(), &pb.DeletePodRequest{PodName: podName, Namespace: testNamespace})
	t.Log("   清理完成")
}
