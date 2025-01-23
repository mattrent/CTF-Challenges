package infrastructure

import (
	"deployer/config"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	kubevirt "kubevirt.io/api/core/v1"
)

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
			RunStrategy: ptr.To(kubevirt.RunStrategyAlways),
			Template: &kubevirt.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirt.VirtualMachineInstanceSpec{
					Domain: kubevirt.DomainSpec{
						CPU: &kubevirt.CPU{
							Cores: config.Values.VMCPUs,
						},
						Devices: kubevirt.Devices{
							AutoattachGraphicsDevice: ptr.To(false),
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
							Guest: ptr.To(resource.MustParse(config.Values.MaxVMMemory)),
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
					TerminationGracePeriodSeconds: ptr.To(int64(0)),
					Networks: []kubevirt.Network{
						{
							NetworkSource: kubevirt.NetworkSource{
								Pod: &kubevirt.PodNetwork{},
							},
							Name: "default",
						},
					},
					LivenessProbe: &kubevirt.Probe{
						Handler: kubevirt.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/",
								Port: intstr.FromInt(8080),
							},
						},
						InitialDelaySeconds: 600,
						PeriodSeconds:       30,
						TimeoutSeconds:      10,
						FailureThreshold:    5,
					},
				},
			},
		},
	}
}

func buildCloudInit(challengeId, token, challengeUrl string) string {
	userData := fmt.Sprintf(`#cloud-config
runcmd:
- mkdir /run/challenge
- wget --no-check-certificate -O "/run/challenge/challenge.zip" "%s/challenges/%s/download?token=%s"
- unzip -d "/run/challenge/challenge/" "/run/challenge/challenge.zip"
- HTTP_PORT="8080" HTTPS_PORT="8443" SSH_PORT="8022" DOMAIN="%s" docker compose -f "/run/challenge/challenge/compose.yaml" up -d
`, config.Values.BackendUrl, challengeId, token, challengeUrl)

	if len(config.Values.VMSSHPUBLICKEY) > 0 {
		userData += fmt.Sprintf(`
ssh_authorized_keys:
- %s
`, config.Values.VMSSHPUBLICKEY)
	}

	return userData
}
