package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type writer struct {
	Str []string
}

func (w *writer) Write(p []byte) (n int, err error) {
	str := string(p)
	if len(str) > 0 {
		w.Str = append(w.Str, str)
	}
	return len(str), nil
}

// RemoteExec will execute a shell command on a given pod in a namespace.
// The shell command does not support chained commands.
// It returns either the stdout or the stderr and error message from the
// command.
func RemoteExec(namespace, podName, command string) (output string, err error) {
	// Initialize
	// TODO: is there a better way to do this?
	// operator/kobjects/driver.go#82 does something similar
	// So in the operator, the config object should already exist
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	kubeClient := kubernetes.NewForConfigOrDie(config)

	// Get the Container name
	pod, err := kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	// TODO: what happens when the first container in the Pods list is not an ES container?  Will this ever be an issue?
	// The Container option in PodExecOptions defaults to the only container in the pod if there is only one container, so may not be necessary if there will only ever be one container in the pod
	containerName := pod.Spec.Containers[0].Name

	// Create the request
	execRequest := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	execRequest.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   strings.Split(command, " "),
		Stdout:    true,
		Stderr:    true,
		Stdin:     true,
	}, scheme.ParameterCodec)

	// DEBUG
	fmt.Printf("URL:\t%s\n", execRequest.URL())
	fmt.Printf("Executing command \"%s\" on container [%s] in pod [%s] \n", command, containerName, podName)

	// Create the executor
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
	if err != nil {
		panic(err.Error())
	}

	// Create the stream objects
	streamIn := strings.NewReader("")
	streamOut := new(writer)
	streamErr := new(writer)

	// Execute the command
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  streamIn,
		Stdout: streamOut,
		Stderr: streamErr,
	})

	// Convert stream objects to strings
	stringOut := strings.Join(streamOut.Str, "")
	stringErr := strings.Join(streamErr.Str, "")

	if err != nil {
		return stringErr, err
	}

	return stringOut, nil
}

func main() {
	fmt.Println("/////////////// Remote Command PoC ///////////////")

	// Parameters
	// namespace := "default"
	// podName := "bb"
	// command := "ls -al"
	namespace := "es1"
	podName := "elasticsearch-0"
	command := "touch hello.txt"

	output, err := RemoteExec(namespace, podName, command)

	if err != nil {
		fmt.Printf("[stderr]\n%s\n", output)
		fmt.Printf("[err]\n%s\n", err)
		return
	}
	if output != "" {
		fmt.Printf("[stdout]\n%s\n", output)
	} else {
		fmt.Println("[Command successful with no stdout]")
	}

	fmt.Println("//////////////////////////////////////////////////")
}
