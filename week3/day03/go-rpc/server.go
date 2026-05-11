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
"k8s.io/apimachinery/pkg/api/resource"
metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
"k8s.io/apimachinery/pkg/util/intstr"
"k8s.io/client-go/kubernetes"
"k8s.io/client-go/tools/clientcmd"
"k8s.io/client-go/util/homedir"
)

var k8sClient *kubernetes.Clientset

type PodManagerServer struct {
pb.UnimplementedPodManagerServer
}

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
home := homedir.HomeDir()
kubeconfig := filepath.Join(home, ".kube", "config")

config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
if err != nil {
log.Fatalf("Load K8s config failed: %v", err)
}
config.Insecure = true
config.CAFile = ""
config.CAData = nil

k8sClient, err = kubernetes.NewForConfig(config)
if err != nil {
log.Fatalf("Create K8s client failed: %v", err)
}
log.Println("K8s connected!")
}

func (s *PodManagerServer) CreateDeployment(ctx context.Context, req *pb.CreateDeploymentRequest) (*pb.CreateDeploymentReply, error) {
replicas := int32(1)
if req.Replicas > 0 {
replicas = req.Replicas
}

cpuReq := req.CpuRequest
if cpuReq == "" { cpuReq = "100m" }
cpuLim := req.CpuLimit
if cpuLim == "" { cpuLim = "200m" }
memReq := req.MemoryRequest
if memReq == "" { memReq = "128Mi" }
memLim := req.MemoryLimit
if memLim == "" { memLim = "256Mi" }

volumes, volumeMounts, coreEnvs := buildVolumesAndEnvs(req.Storages, req.Envs)

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
return &pb.CreateDeploymentReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) DeleteDeployment(ctx context.Context, req *pb.DeleteDeploymentRequest) (*pb.DeleteDeploymentReply, error) {
err := k8sClient.AppsV1().Deployments(req.Namespace).Delete(ctx, req.PodName, metav1.DeleteOptions{})
if err != nil {
return &pb.DeleteDeploymentReply{Success: false, Message: err.Error()}, nil
}
return &pb.DeleteDeploymentReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) CreatePod(ctx context.Context, req *pb.CreatePodRequest) (*pb.CreatePodReply, error) {
cpuReq := req.CpuRequest
if cpuReq == "" { cpuReq = "100m" }
cpuLim := req.CpuLimit
if cpuLim == "" { cpuLim = "200m" }
memReq := req.MemoryRequest
if memReq == "" { memReq = "128Mi" }
memLim := req.MemoryLimit
if memLim == "" { memLim = "256Mi" }

volumes, volumeMounts, coreEnvs := buildVolumesAndEnvs(req.Storages, req.Envs)

pod := &corev1.Pod{
ObjectMeta: metav1.ObjectMeta{Name: req.PodName, Namespace: req.Namespace, Labels: map[string]string{"app": req.PodName}},
Spec: corev1.PodSpec{
Containers: []corev1.Container{{
Name:      req.PodName,
Image:     req.Image,
Env:       coreEnvs,
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
return &pb.CreatePodReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) DeletePod(ctx context.Context, req *pb.DeletePodRequest) (*pb.DeletePodReply, error) {
err := k8sClient.CoreV1().Pods(req.Namespace).Delete(ctx, req.PodName, metav1.DeleteOptions{})
if err != nil {
return &pb.DeletePodReply{Success: false, Message: err.Error()}, nil
}
return &pb.DeletePodReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) GetPod(ctx context.Context, req *pb.GetPodRequest) (*pb.GetPodReply, error) {
pod, err := k8sClient.CoreV1().Pods(req.Namespace).Get(ctx, req.PodName, metav1.GetOptions{})
if err != nil {
return &pb.GetPodReply{Success: false, Message: err.Error()}, nil
}
image := ""
if len(pod.Spec.Containers) > 0 {
image = pod.Spec.Containers[0].Image
}
return &pb.GetPodReply{Success: true, Message: "Success", Status: string(pod.Status.Phase), Image: image}, nil
}

func (s *PodManagerServer) CreateService(ctx context.Context, req *pb.CreateServiceRequest) (*pb.CreateServiceReply, error) {
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
return &pb.CreateServiceReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) DeleteService(ctx context.Context, req *pb.DeleteServiceRequest) (*pb.DeleteServiceReply, error) {
err := k8sClient.CoreV1().Services(req.Namespace).Delete(ctx, req.ServiceName, metav1.DeleteOptions{})
if err != nil {
return &pb.DeleteServiceReply{Success: false, Message: err.Error()}, nil
}
return &pb.DeleteServiceReply{Success: true, Message: "Success"}, nil
}

func (s *PodManagerServer) GetService(ctx context.Context, req *pb.GetServiceRequest) (*pb.GetServiceReply, error) {
svc, err := k8sClient.CoreV1().Services(req.Namespace).Get(ctx, req.ServiceName, metav1.GetOptions{})
if err != nil {
return &pb.GetServiceReply{Success: false, Message: err.Error()}, nil
}

var rPorts []*pb.PortConfig
for _, p := range svc.Spec.Ports {
rPorts = append(rPorts, &pb.PortConfig{
Port:      p.Port,
TargetPort: p.TargetPort.IntVal,
NodePort:  p.NodePort,
})
}
return &pb.GetServiceReply{
Success:   true,
Message:   "Success",
Type:      string(svc.Spec.Type),
Selector:  svc.Spec.Selector,
Ports:     rPorts,
ClusterIp: svc.Spec.ClusterIP,
}, nil
}

func main() {
initK8s()
lis, err := net.Listen("tcp", "0.0.0.0:50051")
if err != nil {
log.Fatalf("failed to listen: %v", err)
}
s := grpc.NewServer()
pb.RegisterPodManagerServer(s, &PodManagerServer{})
log.Println("Server started on :50051")
if err := s.Serve(lis); err != nil {
log.Fatalf("failed to serve: %v", err)
}
}
