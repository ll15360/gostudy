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
func TestCreatePod_Success(t *testing.T) {
	// 安全判断：防止客户端为空
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	// 构造请求，包含资源配置验证
	req := &pb.CreatePodRequest{
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
	resp, err := client.CreatePod(context.Background(), req)
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
func TestCreatePod_DefaultSettings(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	// 构造请求，不穿入 Replicas, CpuRequest, 等配置参数 (即采用空值与0值)
	req := &pb.CreatePodRequest{
		PodName:   "test-nginx-default",
		Image:     "nginx:alpine",
		Namespace: testNamespace,
		// Replicas，CpuRequest，MemRequest等全部忽略不传...
	}

	// 调用接口
	resp, err := client.CreatePod(context.Background(), req)
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
func TestDeletePod_Success(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.DeletePodRequest{
		PodName:   "test-nginx-02",
		Namespace: testNamespace,
	}

	resp, err := client.DeletePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if !resp.Success {
		t.Fatalf("删除失败：%s", resp.Message)
	}

	t.Log("✅ 测试通过：存在的 Deployment 删除成功！")
}

// ===================== 第三个测试：删除不存在的Deployment（失败） =====================
func TestDeletePod_NotFound(t *testing.T) {
	if client == nil {
		t.Fatal("客户端未初始化，请先启动gRPC服务端！")
	}

	req := &pb.DeletePodRequest{
		PodName:   "test-nginx-01",
		Namespace: testNamespace,
	}

	resp, err := client.DeletePod(context.Background(), req)
	if err != nil {
		t.Fatalf("调用接口失败：%v", err)
	}

	if resp.Success {
		t.Fatalf("测试失败：删除不存在的资源预期应该返回失败，但返回了成功！")
	}

	t.Logf("✅ 测试通过：删除不存在的 Deployment 成功拦截，服务端返回提示：%s", resp.Message)
}
