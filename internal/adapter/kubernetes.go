package adapter

import (
	"context"
	"errors"
	"fmt"
	"github.com/sakuraapp/shared/pkg/model"
	"github.com/sakuraapp/supervisor/internal/config"
	"github.com/sakuraapp/supervisor/internal/supervisor"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

var configErr = errors.New("couldn't find kubeconfig")

const (
	namespace = "rooms"
	podName = "room-%v"
)

type KubernetesAdapter struct {
	conf *config.Config
	clientset *kubernetes.Clientset
}

func (a *KubernetesAdapter) createPod(roomId model.RoomId, region supervisor.Region) *corev1.Pod {
	conf := a.conf
	strRoomId := strconv.FormatInt(int64(roomId), 10)

	nodeSelector := map[string]string{}

	if region != supervisor.RegionANY {
		nodeSelector["region"] = string(region)
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(podName, roomId),
			Namespace: namespace,
			Labels: map[string]string{
				"room": strRoomId,
				"kind": "room",
			},
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "dshm",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Image: conf.RoomImage,
					Resources: corev1.ResourceRequirements{
						Limits: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    conf.RoomCPULimit,
							corev1.ResourceMemory: conf.RoomMemoryLimit,
						},
						Requests: map[corev1.ResourceName]resource.Quantity{
							corev1.ResourceCPU:    conf.RoomCPURequests,
							corev1.ResourceMemory: conf.RoomMemoryRequests,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name: "dshm",
							MountPath: "/dev/shm",
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "ROOM_ID",
							Value: strRoomId,
						},
						{
							Name: "CHAKRA_ADDR",
							Value: conf.ChakraAddr,
						},
					},
				},
			},
			NodeSelector: nodeSelector,
		},
	}
}

func (a *KubernetesAdapter) Deploy(ctx context.Context, roomId model.RoomId, region supervisor.Region) error {
	var err error

	pod := a.createPod(roomId, region)
	pod, err = a.clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})

	return err
}

func (a *KubernetesAdapter) Destroy(ctx context.Context, roomId model.RoomId) error {
	name := fmt.Sprintf(podName, roomId)

	return a.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (a *KubernetesAdapter) GetAvailableRooms(ctx context.Context, region supervisor.Region) (int64, error) {
	matchLabels := map[string]string{}

	if region != supervisor.RegionANY {
		matchLabels["region"] = string(region)
	}

	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: matchLabels,
	})

	if err != nil {
		return 0, err
	}

	nodeList, err := a.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		return 0, err
	}

	memoryPerRoom := a.conf.RoomMemoryLimit
	cpuPerRoom := a.conf.RoomCPULimit

	var availRooms int64

	for _, node := range nodeList.Items {
		allocatable := node.Status.Allocatable

		mem := allocatable.Memory()
		cpu := allocatable.Cpu()

		availRoomsMem := math.Floor(float64(mem.Value() / memoryPerRoom.Value()))
		availRoomsCpu := math.Floor(float64(cpu.Value() / cpuPerRoom.Value()))

		availRooms = availRooms + int64(math.Min(availRoomsMem, availRoomsCpu))
	}

	return availRooms, nil
}

func NewKubernetesAdapter(conf *config.Config) (*KubernetesAdapter, error) {
	var myConfig *rest.Config
	var err error

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		// in-cluster
		myConfig, err = rest.InClusterConfig()
	} else {
		// out-of-cluster
		var kubeconfig string

		if conf.K8SConfigPath != "" {
			kubeconfig = conf.K8SConfigPath
		} else if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return nil, configErr
		}

		myConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(myConfig)

	if err != nil {
		return nil, err
	}

	return &KubernetesAdapter{
		conf: conf,
		clientset: clientset,
	}, nil
}