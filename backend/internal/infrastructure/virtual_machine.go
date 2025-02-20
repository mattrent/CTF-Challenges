package infrastructure

import (
	"deployer/config"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	kubevirt "kubevirt.io/api/core/v1"
)

const labelName = "custom-challenge-selector"

func BuildContainer(challengeId, token, namespace, challengeUrl string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "challenge",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelName: namespace,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						labelName: namespace,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "challenge-container",
							Image: "hello-world",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
								{
									ContainerPort: 8022,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HTTP_PORT",
									Value: "8080",
								},
								{
									Name:  "SSH_PORT",
									Value: "8022",
								},
								{
									Name:  "DOMAIN",
									Value: challengeUrl,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: config.Values.ChallengeLivenessProbe.InitialDelaySeconds,
								PeriodSeconds:       config.Values.ChallengeLivenessProbe.PeriodSeconds,
								TimeoutSeconds:      config.Values.ChallengeLivenessProbe.TimeoutSeconds,
								FailureThreshold:    config.Values.ChallengeLivenessProbe.FailureThreshold,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: config.Values.ChallengeReadinessProbe.InitialDelaySeconds,
								PeriodSeconds:       config.Values.ChallengeReadinessProbe.PeriodSeconds,
								TimeoutSeconds:      config.Values.ChallengeReadinessProbe.TimeoutSeconds,
								FailureThreshold:    config.Values.ChallengeReadinessProbe.FailureThreshold,
							},
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: config.Values.ChallengeStartupProbe.InitialDelaySeconds,
								PeriodSeconds:       config.Values.ChallengeStartupProbe.PeriodSeconds,
								TimeoutSeconds:      config.Values.ChallengeStartupProbe.TimeoutSeconds,
								FailureThreshold:    config.Values.ChallengeStartupProbe.FailureThreshold,
							},
						},
					},
				},
			},
		},
	}
}

func BuildVm(challengeId, token, namespace, challengeUrl string) *kubevirt.VirtualMachine {
	containerDiskName := "containerdisk"
	cloudInitDiskName := "cloudinitdisk"
	userData := buildCloudInit(challengeId, token, challengeUrl)

	containerDisk := kubevirt.ContainerDiskSource{
		Image: config.Values.VMImageUrl,
	}
	if strings.TrimSpace(config.Values.ImagePullSecret) != "" {
		containerDisk.ImagePullSecret = config.Values.ImagePullSecret
	}

	return &kubevirt.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kubevirt.SchemeGroupVersion.String(),
			Kind:       "VirtualMachine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Labels: map[string]string{
				labelName: namespace,
			},
			Name: "challenge",
		},
		Spec: kubevirt.VirtualMachineSpec{
			RunStrategy: ptr.To(kubevirt.RunStrategyAlways),
			Template: &kubevirt.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						labelName: namespace,
					},
				},
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
								ContainerDisk: &containerDisk,
							},
							Name: containerDiskName,
						},
						{
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
						InitialDelaySeconds: config.Values.ChallengeLivenessProbe.InitialDelaySeconds,
						PeriodSeconds:       config.Values.ChallengeLivenessProbe.PeriodSeconds,
						TimeoutSeconds:      config.Values.ChallengeLivenessProbe.TimeoutSeconds,
						FailureThreshold:    config.Values.ChallengeLivenessProbe.FailureThreshold,
					},
					ReadinessProbe: &kubevirt.Probe{
						Handler: kubevirt.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/",
								Port: intstr.FromInt(8080),
							},
						},
						InitialDelaySeconds: config.Values.ChallengeReadinessProbe.InitialDelaySeconds,
						PeriodSeconds:       config.Values.ChallengeReadinessProbe.PeriodSeconds,
						TimeoutSeconds:      config.Values.ChallengeReadinessProbe.TimeoutSeconds,
						FailureThreshold:    config.Values.ChallengeReadinessProbe.FailureThreshold,
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
- HTTP_PORT="8080" SSH_PORT="8022" DOMAIN="%s" docker compose -f "/run/challenge/challenge/compose.yaml" up -d
`, config.Values.BackendUrl, challengeId, token, challengeUrl)

	if len(config.Values.VMSSHPUBLICKEY) > 0 {
		userData += fmt.Sprintf(`
ssh_authorized_keys:
- %s
`, config.Values.VMSSHPUBLICKEY)
	}

	return userData
}
