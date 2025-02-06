package lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var shell = []string{"bash", "-c"}

/*
🙂'😀'
slightly smiling face
Unicode: U+1F642, UTF-8: F0 9F 99 82
*/
const (
	nok         = "💣"
	ok          = "😀"
	getContexts = "kubectl config get-contexts --output=name"
	getNS       = "kubectl %s get ns --no-headers -o custom-columns=:metadata.name"
	getSVC      = "kubectl %s get svc --no-headers -o custom-columns=:metadata.name,:spec.ports[*].port"
)

func OkIcon(eval bool) string {
	if eval {
		return ok
	}

	return nok
}

type K8SService struct {
	Context   string
	Namespace string
	Name      string
	Ports     []string
}

func GetKubectlContexts() ([]string, error) {
	return runKubectlCommand(getContexts)
}

func GetKubectlNamespaces(ctx string) ([]string, error) {
	ctxArg := ""
	if ctx != "" {
		ctxArg = fmt.Sprintf("--context %s", ctx) // Removed extra space
	}
	return runKubectlCommand(fmt.Sprintf(getNS, ctxArg))
}

func GetKubectlServices(context, namespace string) ([]K8SService, error) {
	extraArgs := ""

	if context != "" {
		extraArgs = fmt.Sprintf("--context %s", context) // Removed extra space
	}

	if namespace != "" {
		extraArgs = fmt.Sprintf("%s -n %s", extraArgs, namespace) // Removed extra space
	}

	output, err := runKubectlCommand(fmt.Sprintf(getSVC, extraArgs))
	if err != nil {
		return nil, err
	}

	var services []K8SService
	lines := output

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			service := K8SService{
				Context:   context,
				Namespace: namespace,
				Name:      fields[0],
				Ports:     strings.Split(fields[1], ","),
			}
			services = append(services, service)
		}
	}

	return services, nil
}

func runKubectlCommand(command string) ([]string, error) {
	cmdArgs := append(shell, command)
	fmt.Println("cmdArgs:", cmdArgs) // Keep for debugging

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stderr = os.Stderr // Redirect stderr for better debugging

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return splitLines(string(output)), nil
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func GeneratePortForwardCommand(svc K8SService) string {
	return fmt.Sprintf("kubectl --context %s -n %s port-forward service/%s %s", svc.Context, svc.Namespace, svc.Name, strings.Join(svc.Ports, " "))
}

func NewCommand(svc K8SService) *exec.Cmd {
	c := append(shell, GeneratePortForwardCommand(svc))
	cmd := exec.Command(c[0], c[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Run in separate process group
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (svc K8SService) NewCommand2() *exec.Cmd {
	c := append(shell, GeneratePortForwardCommand(svc))
	cmd := exec.Command(c[0], c[1:]...)

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // Run in separate process group
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
