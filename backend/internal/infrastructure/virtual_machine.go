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
			Name:      "challenge",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: kubevirt.VirtualMachineSpec{
			RunStrategy: ptr(kubevirt.RunStrategyRerunOnFailure),
			Template: &kubevirt.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: kubevirt.VirtualMachineInstanceSpec{
					Domain: kubevirt.DomainSpec{
						Devices: kubevirt.Devices{
							AutoattachGraphicsDevice: ptr(false),
							Disks: []kubevirt.Disk{
								{
									Name: containerDiskName,
									DiskDevice: kubevirt.DiskDevice{
										Disk: &kubevirt.DiskTarget{
											Bus: kubevirt.DiskBusVirtio,
										},
									},
								},
								{
									Name: cloudInitDiskName,
									DiskDevice: kubevirt.DiskDevice{
										Disk: &kubevirt.DiskTarget{
											Bus: kubevirt.DiskBusVirtio,
										},
									},
								},
							},
							Interfaces: []kubevirt.Interface{
								{
									Name: "default",
									InterfaceBindingMethod: kubevirt.InterfaceBindingMethod{
										Masquerade: &kubevirt.InterfaceMasquerade{},
									},
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
							Name: containerDiskName,
							VolumeSource: kubevirt.VolumeSource{
								ContainerDisk: &kubevirt.ContainerDiskSource{
									Image: config.Values.VMImageUrl,
								},
							},
						}, {
							Name: cloudInitDiskName,
							VolumeSource: kubevirt.VolumeSource{
								CloudInitNoCloud: &kubevirt.CloudInitNoCloudSource{
									UserData: userData,
								},
							},
						},
					},
					Networks: []kubevirt.Network{
						{
							Name: "default",
							NetworkSource: kubevirt.NetworkSource{
								Pod: &kubevirt.PodNetwork{},
							},
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
- DOMAIN="%s" docker compose -f /tmp/challenge/compose.yaml up -d
`, config.Values.BackendUrl, challengeId, token, challengeUrl)
	return userData
}
