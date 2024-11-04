package infrastructure

import (
	"deployer/config"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirt "kubevirt.io/api/core/v1"
)

func ptr[V any](v V) *V {
	return &v
}

func BuildVm(challengeId, token, namespace, challengeUrl string) *kubevirt.VirtualMachine {
	containerDiskName := "containerdisk"
	cloudInitDiskName := "cloudinitdisk"
	userData := buildCloudInit(challengeId, token, challengeUrl)

	return &kubevirt.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirt.SchemeGroupVersion.String(),
			Kind:       "VirtualMachine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Labels:    map[string]string{},
			Name:      "challenge",
		},
		Spec: kubevirt.VirtualMachineSpec{
			RunStrategy: ptr(kubevirt.RunStrategyRerunOnFailure),
			Template: &kubevirt.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirt.VirtualMachineInstanceSpec{
					Domain: kubevirt.DomainSpec{
						Devices: kubevirt.Devices{
							AutoattachGraphicsDevice: ptr(false),
							Disks: []kubevirt.Disk{
								{
									DiskDevice: kubevirt.DiskDevice{
										Disk: &kubevirt.DiskTarget{
											Bus: kubevirt.DiskBusVirtio,
										},
									},
									Name: containerDiskName,
								},
								{
									DiskDevice: kubevirt.DiskDevice{
										Disk: &kubevirt.DiskTarget{
											Bus: kubevirt.DiskBusVirtio,
										},
									},
									Name: cloudInitDiskName,
								},
							},
							Interfaces: []kubevirt.Interface{
								{
									InterfaceBindingMethod: kubevirt.InterfaceBindingMethod{
										Masquerade: &kubevirt.InterfaceMasquerade{},
									},
									Name: "default",
								},
							},
						},
						Resources: kubevirt.ResourceRequirements{
							Requests: corev1.ResourceList{
								"memory": resource.MustParse(config.Values.MinVMMemory),
							},
						},
						Memory: &kubevirt.Memory{
							Guest: ptr(resource.MustParse(config.Values.MaxVMMemory)),
						},
					},
					Volumes: []kubevirt.Volume{
						{
							VolumeSource: kubevirt.VolumeSource{
								ContainerDisk: &kubevirt.ContainerDiskSource{
									Image: config.Values.VMImageUrl,
								},
							},
							Name: containerDiskName,
						}, {
							VolumeSource: kubevirt.VolumeSource{
								CloudInitNoCloud: &kubevirt.CloudInitNoCloudSource{
									UserData: userData,
								},
							},
							Name: cloudInitDiskName,
						},
					},
					TerminationGracePeriodSeconds: ptr(int64(30)),
					Networks: []kubevirt.Network{
						{
							NetworkSource: kubevirt.NetworkSource{
								Pod: &kubevirt.PodNetwork{},
							},
							Name: "default",
						},
					},
					// LivenessProbe: &kubevirt.Probe{
					// 	Handler: kubevirt.Handler{
					// 		HTTPGet: &corev1.HTTPGetAction{
					// 			Path: "/",
					// 			Port: intstr.FromInt(80),
					// 		},
					// 	},
					// 	InitialDelaySeconds: 300,
					// 	PeriodSeconds:       30,
					// 	TimeoutSeconds:      10,
					// 	FailureThreshold:    5,
					// },
				},
			},
		},
	}
}

func buildCloudInit(challengeId, token, challengeUrl string) string {
	userData := fmt.Sprintf(`#cloud-config
runcmd:
- [wget, --no-check-certificate, -O, "/tmp/challenge.zip", "%s/challenges/%s/download?token=%s"]
- unzip -d /tmp/challenge/ /tmp/challenge.zip
- HTTP_PORT="80" HTTPS_PORT="443" SSH_PORT="22" DOMAIN="%s" docker compose -f /tmp/challenge/compose.yaml up
`, config.Values.BackendUrl, challengeId, token, challengeUrl)
	return userData
}
