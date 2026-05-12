package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-rpc/pb"
	"log"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	trpc "trpc.group/trpc-go/trpc-go"
)

var k8sClient *kubernetes.Clientset

type PodManagerServer struct {
	pb.UnimplementedPodManager
}

// 辅助函数：构造卷和环境变量
func buildVolumesAndEnvs(storages []*pb.StorageConfig, envs []*pb.EnvVar) ([]corev1.Volume, []corev1.VolumeMount, []corev1.EnvVar) {
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var coreEnvs []corev1.EnvVar

	for _, e := range envs {
		coreEnvs = append(coreEnvs, corev1.EnvVar{
			Name:  e.Name,
			Value: e.Value,
		})
	}

	for _, s := range storages {
		if s == nil {
			continue
		}
		vol := corev1.Volume{Name: s.Name}
		vm := corev1.VolumeMount{
			Name:      s.Name,
			MountPath: s.MountPath,
		}
		switch s.Type {
		case "hostPath":
			vol.VolumeSource = corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: s.HostPath,
				},
			}
		case "emptyDir":
			vol.VolumeSource = corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}
		case "pvc":
			vol.VolumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: s.PvcName,
				},
			}
		}
		volumes = append(volumes, vol)
		volumeMounts = append(volumeMounts, vm)
	}
	return volumes, volumeMounts, coreEnvs
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

func (s *PodManagerServer) CreateDeployment(ctx context.Context, req *pb.CreateDeploymentRequest) (*pb.CreateDeploymentReply, error) {
	log.Printf("创建Deployment请求：NS=%s, Name=%s", req.Namespace, req.PodName)

	// 1. 副本数默认值处理
	replicas := int32(1)
	if req.Replicas > 0 {
		replicas = req.Replicas
	}

	// 2. 资源默认值处理
	cpuReq := req.CpuRequest
	if cpuReq == "" {
		cpuReq = "100m"
	}
	cpuLim := req.CpuLimit
	if cpuLim == "" {
		cpuLim = "200m"
	}
	memReq := req.MemoryRequest
	if memReq == "" {
		memReq = "128Mi"
	}
	memLim := req.MemoryLimit
	if memLim == "" {
		memLim = "256Mi"
	}

	// 解析卷与环境变量
	volumes, volumeMounts, coreEnvs := buildVolumesAndEnvs(req.Storages, req.Envs)

	// 3. 构建 K8s 资源结构体
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuReq),
			corev1.ResourceMemory: resource.MustParse(memReq),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuLim),
			corev1.ResourceMemory: resource.MustParse(memLim),
		},
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: req.PodName, Namespace: req.Namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": req.PodName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": req.PodName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:         req.PodName,
						Image:        req.Image,
						Resources:    resources,
						VolumeMounts: volumeMounts,
						Env:          coreEnvs,
					}},
					Volumes: volumes,
				},
			},
		},
	}
	_, err := k8sClient.AppsV1().Deployments(req.Namespace).Create(ctx, deploy, metav1.CreateOptions{})
	if err != nil {
		return &pb.CreateDeploymentReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.CreateDeploymentReply{Success: true, Message: "创建Deployment成功"}, nil
}

func (s *PodManagerServer) DeleteDeployment(ctx context.Context, req *pb.DeleteDeploymentRequest) (*pb.DeleteDeploymentReply, error) {
	log.Printf("删除Deployment请求：NS=%s, Name=%s", req.Namespace, req.PodName)
	err := k8sClient.AppsV1().Deployments(req.Namespace).Delete(ctx, req.PodName, metav1.DeleteOptions{})
	if err != nil {
		return &pb.DeleteDeploymentReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.DeleteDeploymentReply{Success: true, Message: "删除Deployment成功"}, nil
}

func (s *PodManagerServer) CreatePod(ctx context.Context, req *pb.CreatePodRequest) (*pb.CreatePodReply, error) {
	log.Printf("创建Pod请求：NS=%s, Name=%s", req.Namespace, req.PodName)
	cpuReq := req.CpuRequest
	if cpuReq == "" {
		cpuReq = "100m"
	}
	cpuLim := req.CpuLimit
	if cpuLim == "" {
		cpuLim = "200m"
	}
	memReq := req.MemoryRequest
	if memReq == "" {
		memReq = "128Mi"
	}
	memLim := req.MemoryLimit
	if memLim == "" {
		memLim = "256Mi"
	}

	volumes, volumeMounts, coreEnvs := buildVolumesAndEnvs(req.Storages, req.Envs)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: req.PodName, Namespace: req.Namespace, Labels: map[string]string{"app": req.PodName}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:         req.PodName,
				Image:        req.Image,
				Env:          coreEnvs,
				VolumeMounts: volumeMounts,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(cpuReq),
						corev1.ResourceMemory: resource.MustParse(memReq),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(cpuLim),
						corev1.ResourceMemory: resource.MustParse(memLim),
					},
				},
			}},
			Volumes: volumes,
		},
	}
	_, err := k8sClient.CoreV1().Pods(req.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return &pb.CreatePodReply{Success: false, Message: err.Error()}, nil
	}

	// === 开始数据库持久化 ===
	labelsJson, _ := json.Marshal(map[string]string{"app": req.PodName})

	podRecord := &PodInfo{
		PodName:       req.PodName,
		Namespace:     req.Namespace,
		Image:         req.Image,
		Status:        string(corev1.PodPending), // 刚创建默认为 Pending
		Labels:        string(labelsJson),
		CPURequest:    cpuReq,
		CPULimit:      cpuLim,
		MemoryRequest: memReq,
		MemoryLimit:   memLim,
	}

	if err := DB.Create(podRecord).Error; err != nil {
		// K8s 创建成功，但是写入数据库失败，可以考虑在这里做补充处理或是仅打印日志
		log.Printf("⚠️ Pod写入数据库失败: %v", err)
		return &pb.CreatePodReply{Success: true, Message: fmt.Sprintf("创建Pod成功，但数据库持久化失败: %v", err)}, nil
	}

	return &pb.CreatePodReply{Success: true, Message: "创建Pod成功并在数据库持久化"}, nil
}

func (s *PodManagerServer) DeletePod(ctx context.Context, req *pb.DeletePodRequest) (*pb.DeletePodReply, error) {
	log.Printf("删除Pod请求：NS=%s, Name=%s", req.Namespace, req.PodName)
	err := k8sClient.CoreV1().Pods(req.Namespace).Delete(ctx, req.PodName, metav1.DeleteOptions{})
	if err != nil {
		return &pb.DeletePodReply{Success: false, Message: err.Error()}, nil
	}
	// 更新数据库中该 Pod 的状态为 Deleted（如果存在）
	if DB != nil {
		if err := DB.Model(&PodInfo{}).
			Where("pod_name = ? AND namespace = ?", req.PodName, req.Namespace).
			Update("status", "Deleted").Error; err != nil {
			log.Printf("⚠️ 更新数据库 Pod 状态为 Deleted 失败: %v", err)
		}
	}
	return &pb.DeletePodReply{Success: true, Message: "删除Pod成功"}, nil
}

func (s *PodManagerServer) GetPod(ctx context.Context, req *pb.GetPodRequest) (*pb.GetPodReply, error) {
	// 优先从数据库读取持久化的 Pod 信息
	if DB != nil {
		var pi PodInfo
		if err := DB.Where("pod_name = ? AND namespace = ?", req.PodName, req.Namespace).First(&pi).Error; err == nil {
			// 成功从 DB 获取到记录，尝试从 K8s 查询实时状态并同步
			liveStatus := pi.Status
			pod, err := k8sClient.CoreV1().Pods(req.Namespace).Get(ctx, req.PodName, metav1.GetOptions{})
			if err == nil {
				liveStatus = string(pod.Status.Phase)
				// 如果和 DB 中不同，则更新数据库
				if liveStatus != pi.Status {
					if err := DB.Model(&PodInfo{}).
						Where("pod_name = ? AND namespace = ?", req.PodName, req.Namespace).
						Update("status", liveStatus).Error; err != nil {
						log.Printf("⚠️ 同步 Pod 实时状态到 DB 失败: %v", err)
					}
				}
				// 取镜像信息
				image := ""
				if len(pod.Spec.Containers) > 0 {
					image = pod.Spec.Containers[0].Image
				}
				return &pb.GetPodReply{Success: true, Message: "获取Pod成功(来自DB并同步K8s)", Status: liveStatus, Image: image}, nil
			}
			// 如果无法读取 K8s（例如已被删除），仍然返回 DB 中的信息
			return &pb.GetPodReply{Success: true, Message: "获取Pod成功(来自DB)", Status: pi.Status, Image: pi.Image}, nil
		}
	}

	// 如果数据库不存在记录或 DB 未初始化，直接从 K8s 获取
	pod, err := k8sClient.CoreV1().Pods(req.Namespace).Get(ctx, req.PodName, metav1.GetOptions{})
	if err != nil {
		return &pb.GetPodReply{Success: false, Message: err.Error()}, nil
	}
	image := ""
	if len(pod.Spec.Containers) > 0 {
		image = pod.Spec.Containers[0].Image
	}
	return &pb.GetPodReply{Success: true, Message: "获取Pod成功(来自K8s)", Status: string(pod.Status.Phase), Image: image}, nil
}

func (s *PodManagerServer) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.CreateServiceReply, error) {
	log.Printf("创建Service请求：NS=%s, Name=%s", req.Namespace, req.ServiceName)
	var svcPorts []corev1.ServicePort
	for _, p := range req.Ports {
		portCfg := corev1.ServicePort{
			Port:       p.Port,
			TargetPort: intstr.FromInt32(p.TargetPort),
		}
		if req.Type == "NodePort" && p.NodePort > 0 {
			portCfg.NodePort = p.NodePort
		}
		svcPorts = append(svcPorts, portCfg)
	}

	svcType := corev1.ServiceTypeClusterIP
	if req.Type == "NodePort" {
		svcType = corev1.ServiceTypeNodePort
	} else if req.Type == "LoadBalancer" {
		svcType = corev1.ServiceTypeLoadBalancer
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: req.ServiceName, Namespace: req.Namespace},
		Spec: corev1.ServiceSpec{
			Type:     svcType,
			Selector: req.Selector,
			Ports:    svcPorts,
		},
	}

	_, err := k8sClient.CoreV1().Services(req.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return &pb.CreateServiceReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.CreateServiceReply{Success: true, Message: "创建Service成功"}, nil
}

func (s *PodManagerServer) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest) (*pb.DeleteServiceReply, error) {
	err := k8sClient.CoreV1().Services(req.Namespace).Delete(ctx, req.ServiceName, metav1.DeleteOptions{})
	if err != nil {
		return &pb.DeleteServiceReply{Success: false, Message: err.Error()}, nil
	}
	return &pb.DeleteServiceReply{Success: true, Message: "删除Service成功"}, nil
}

func (s *PodManagerServer) GetService(ctx context.Context, req *pb.GetServiceRequest) (*pb.GetServiceReply, error) {
	svc, err := k8sClient.CoreV1().Services(req.Namespace).Get(ctx, req.ServiceName, metav1.GetOptions{})
	if err != nil {
		return &pb.GetServiceReply{Success: false, Message: err.Error()}, nil
	}

	var rPorts []*pb.PortConfig
	for _, p := range svc.Spec.Ports {
		rPorts = append(rPorts, &pb.PortConfig{
			Port:       p.Port,
			TargetPort: p.TargetPort.IntVal,
			NodePort:   p.NodePort,
		})
	}
	return &pb.GetServiceReply{
		Success:   true,
		Message:   "获取Service成功",
		Type:      string(svc.Spec.Type),
		Selector:  svc.Spec.Selector,
		Ports:     rPorts,
		ClusterIp: svc.Spec.ClusterIP,
	}, nil
}

// TODO:待添加数据库持久化

// TODO:添加这个滚动更新

// TODO:如何指定这个pod中的容器启动顺序

// TODO:服务暴露,拉起的pod如何在外网能够成功访问这个服务

// 数据的持久化和共享问题

// 权限问题

// 环境变量

func main() {
	InitDB()

	initK8s()

	s := trpc.NewServer()
	pb.RegisterPodManagerService(s, &PodManagerServer{})
	log.Println("🚀 tRPC服务已启动 :50051")
	if err := s.Serve(); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
