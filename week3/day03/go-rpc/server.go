package main

import (
	"context"
	"go-rpc/pb"
	"log"
	"net"
	"path/filepath"

	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var k8sClient *kubernetes.Clientset

type PodManagerServer struct {
	pb.UnimplementedPodManagerServer
}

func initK8s() {
	// 自动读取本地config，原生正常连接，无任何hack
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("加载K8s配置失败: %v", err)
	}

	// 跳过 TLS 证书校验（解决本地直连外部服务器由于证书 IP 不匹配导致的报错）
	config.Insecure = true
	config.CAFile = ""
	config.CAData = nil

	k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("创建K8s客户端失败: %v", err)
	}
	log.Println("✅ K8s 连接成功！")
}

func (s *PodManagerServer) CreatePod(ctx context.Context, req *pb.CreatePodRequest) (*pb.CreatePodReply, error) {
	log.Printf("创建请求：NS=%s, Name=%s", req.Namespace, req.PodName)
	replicas := int32(1)
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: req.PodName, Namespace: req.Namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": req.PodName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": req.PodName}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: req.PodName, Image: req.Image}}},
			},
		},
	}
	_, err := k8sClient.AppsV1().Deployments(req.Namespace).Create(ctx, deploy, metav1.CreateOptions{})
	if err != nil {
		return &pb.CreatePodReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.CreatePodReply{Success: true, Message: "创建成功"}, nil
}

func (s *PodManagerServer) DeletePod(ctx context.Context, req *pb.DeletePodRequest) (*pb.DeletePodReply, error) {
	log.Printf("删除请求：NS=%s, Name=%s", req.Namespace, req.PodName)
	err := k8sClient.AppsV1().Deployments(req.Namespace).Delete(ctx, req.PodName, metav1.DeleteOptions{})
	if err != nil {
		return &pb.DeletePodReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.DeletePodReply{Success: true, Message: "删除成功"}, nil
}

// TODO:添加单个pod的资源限制

// TODO:添加副本数量的参数:不然一个deployment只有一个副本pod

// TODO:添加这个滚动更新

// TODO:如何指定这个pod中的容器启动顺序

// TODO:服务暴露,拉起的pod如何在外网能够成功访问这个服务

// 数据的持久化和共享问题

// 权限问题

// 环境变量

func main() {
	initK8s()
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPodManagerServer(s, &PodManagerServer{})
	log.Println("🚀 gRPC服务已启动 :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
