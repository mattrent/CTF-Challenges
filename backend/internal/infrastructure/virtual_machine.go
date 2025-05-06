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

// TODO add volume for docker-compose or fix authentication for wget /download endpoint
// ! Liveness probes basically don't work

func BuildContainer(challengeId, userId, token, namespace, challengeUrl string, testMode bool) *appsv1.Deployment {
	const emptydirDocker = "emptydir-docker"
	const emptydirFlag = "emptydir-flag"
	const emptydirCode = "emptydir-code"

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", config.Values.VMCPUs)),
			corev1.ResourceMemory: resource.MustParse(config.Values.MaxVMMemory),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", config.Values.VMCPUs)),
			corev1.ResourceMemory: resource.MustParse(config.Values.MinVMMemory),
		},
	}

	var runCommand []string
	var codeFolder string
	if testMode {
		codeFolder = "/run/test"
		runCommand = []string{
			"mkdir /run/test",
			fmt.Sprintf(
				`wget --no-check-certificate -O "/run/test/solution.zip" "%s/solutions/%s/download?token=%s"`,
				config.Values.BackendUrl,
				challengeId,
				token,
			),
			`unzip -d "/run/test/solution/" "/run/test/solution.zip"`,
			"docker build /run/test/solution/ -f /run/test/solution/Dockerfile -t test",
			"docker run -e HTTP_PORT -e SSH_PORT -e DOMAIN -e SSH_SERVICE_INTERNAL_URL -v /run/solution:/run/solution test",
			"sleep 30",
			"FLAG=$(cat /run/solution/flag.txt | tr -d '[:space:]')",
			fmt.Sprintf(
				`curl -k -X POST -d "{\"flag\":\"$FLAG\"}" %s/solutions/%s/verify`,
				config.Values.BackendUrl,
				challengeId,
			),
			// The same as halting the VM (without restart)
			"sleep 36000",
		}
	} else {
		codeFolder = "/run/challenge"
		runCommand = []string{
			"mkdir /run/challenge",
			fmt.Sprintf(
				`wget --no-check-certificate -O "/run/challenge/challenge.zip" "%s/challenges/%s/download?token=%s"`,
				config.Values.BackendUrl,
				challengeId,
				token,
			),
			`unzip -d "/run/challenge/challenge/" "/run/challenge/challenge.zip"`,
			// Needs to be attached. Container will stop once the command finishes
			`docker compose -f "/run/challenge/challenge/compose.yaml" up`,
		}
	}

	userData := buildContainerInit(runCommand)

	container := corev1.Container{
		// https://www.jenkins.io/doc/book/installing/docker/
		Name:    "challenge-container",
		Image:   config.Values.ContainerImageUrl,
		Command: []string{"/bin/bash"},

		// run in foreground
		Args: []string{"-c", userData},
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
			{
				Name:  "SSH_SERVICE_INTERNAL_URL",
				Value: sshUrl(challengeUrl),
			},
			// 2376 for TLS; otherwise 2375
			{
				Name:  "DOCKER_HOST",
				Value: "tcp://localhost:2376",
			},
			{
				Name:  "DOCKER_CERT_PATH",
				Value: "/var/run/docker-certs",
			},
			{
				Name:  "DOCKER_TLS_VERIFY",
				Value: "1",
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      emptydirDocker,
				MountPath: "/var/run/docker-certs",
				ReadOnly:  true,
			},
			// For the flag
			{
				Name:      emptydirFlag,
				MountPath: "/run/solution",
				ReadOnly:  true,
			},
			{
				Name:      emptydirCode,
				MountPath: codeFolder,
				ReadOnly:  false,
			},
		},
		Resources: resourceRequirements,
	}

	if !testMode {
		container.LivenessProbe = &corev1.Probe{
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
		}
		container.ReadinessProbe = &corev1.Probe{
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
		}
		container.StartupProbe = &corev1.Probe{
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
		}
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{container,
			{
				// https://hub.docker.com/_/docker
				Name:  "docker",
				Image: "docker:28.0.0-dind",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 2376,
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "DOCKER_TLS_CERTDIR",
						Value: "/certs",
					},
				},
				SecurityContext: &corev1.SecurityContext{
					Privileged: ptr.To(true),
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      emptydirDocker,
						MountPath: "/certs/client",
						ReadOnly:  false,
					},
					{
						Name:      emptydirFlag,
						MountPath: "/run/solution",
						ReadOnly:  false,
					},
					{
						Name:      emptydirCode,
						MountPath: codeFolder,
						ReadOnly:  false,
					},
				},
				Resources: resourceRequirements,
				LivenessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						Exec: &corev1.ExecAction{
							Command: []string{"docker", "ps"},
						},
					},
					// ? Maybe make adjustable via config
					InitialDelaySeconds: 30,
					PeriodSeconds:       30,
					TimeoutSeconds:      10,
					FailureThreshold:    3,
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: emptydirDocker,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: emptydirFlag,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: emptydirCode,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		},
	}

	if strings.TrimSpace(config.Values.ImagePullSecret) != "" {
		podSpec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: config.Values.ImagePullSecret,
			},
		}
	}

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
						labelName:    namespace,
						"managed-by": "container",
					},
				},
				Spec: podSpec,
			},
		},
	}
}

func BuildVm(challengeId, userId, token, namespace, challengeUrl string, testMode bool) *kubevirt.VirtualMachine {
	containerDiskName := "containerdisk"
	cloudInitDiskName := "cloudinitdisk"

	var runCommand []string
	if testMode {
		runCommand = []string{
			"mkdir /run/test",
			fmt.Sprintf(
				`wget --no-check-certificate -O "/run/test/solution.zip" "%s/solutions/%s/download?token=%s"`,
				config.Values.BackendUrl,
				challengeId,
				token,
			),
			`echo "" | tee /dev/ttyS0`,
			`unzip -d "/run/test/solution/" "/run/test/solution.zip"`,
			"docker build /run/test/solution/ -f /run/test/solution/Dockerfile -t test &> /dev/ttyS0",
			fmt.Sprintf(
				`docker run --name test-container -e HTTP_PORT=8080 -e SSH_PORT=8022 -e DOMAIN="%s" -e SSH_SERVICE_INTERNAL_URL="%s" -v /run/solution:/run/solution test`,
				challengeUrl,
				sshUrl(challengeUrl),
			),
			`echo "sleep 3; echo "" | tee /dev/ttyS0; docker logs -f test-container | tee /dev/ttyS0" > /run/compose-logs-monitor`,
			`sh /run/compose-logs-monitor &`,
			"sleep 30",
			fmt.Sprintf(
				`FLAG=$(cat /run/solution/flag.txt | tr -d '[:space:]') && wget --no-check-certificate --post-data="{\"flag\":\"$FLAG\"}" "%s/solutions/%s/verify"`,
				config.Values.BackendUrl,
				challengeId,
			),
		}
	} else {
		runCommand = []string{
			"mkdir /run/challenge",
			fmt.Sprintf(
				`wget --no-check-certificate -O "/run/challenge/challenge.zip" "%s/challenges/%s/download?token=%s"`,
				config.Values.BackendUrl,
				challengeId,
				token,
			),
			`unzip -d "/run/challenge/challenge/" "/run/challenge/challenge.zip"`,
			// Can be detached. The VM will be stopped won't the container stops
			fmt.Sprintf(
				`HTTP_PORT="8080" SSH_PORT="8022" DOMAIN="%s" docker compose -f /run/challenge/challenge/compose.yaml up -d`,
				challengeUrl,
			),
			`echo "sleep 3; echo "" | tee /dev/ttyS0; docker compose -f /var/run/challenge/challenge/compose.yaml logs -f | tee /dev/ttyS0" > /run/compose-logs-monitor`,
			`sh /run/compose-logs-monitor &`,
		}
	}

	userData := buildCloudInit(runCommand)

	containerDisk := kubevirt.ContainerDiskSource{
		Image: config.Values.VMImageUrl,
	}
	if strings.TrimSpace(config.Values.ImagePullSecret) != "" {
		containerDisk.ImagePullSecret = config.Values.ImagePullSecret
	}

	vm := &kubevirt.VirtualMachine{
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
						labelName:    namespace,
						"managed-by": "vm",
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
				},
			},
		},
	}

	if !testMode {
		vm.Spec.Template.Spec.LivenessProbe = &kubevirt.Probe{
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
		}
		vm.Spec.Template.Spec.ReadinessProbe = &kubevirt.Probe{
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
		}
	}

	return vm
}

func buildCloudInit(runCommand []string) string {
	userData := fmt.Sprintf(`#cloud-config
runcmd:
- %s`, strings.Join(runCommand, "\n- "))

	if len(config.Values.VMSSHPUBLICKEY) > 0 {
		userData += fmt.Sprintf(`
ssh_authorized_keys:
- %s
`, config.Values.VMSSHPUBLICKEY)
	}

	return userData
}

func buildContainerInit(runCommand []string) string {
	userData := fmt.Sprintf(`sleep 30
%s`, strings.Join(runCommand, "\n"))
	return userData
}

// ! Important to use fully qualified domain name
// ! Only needed for VMs
func sshUrl(domain string) string {
	parts := strings.Split(domain, ".")
	return "ssh.challenge-" + parts[0] + ".svc.cluster.local"
}
