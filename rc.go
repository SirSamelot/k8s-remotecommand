package main

import (
	"fmt"
	"io"
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

type Writer struct {
	Str []string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	str := string(p)
	if len(str) > 0 {
		w.Str = append(w.Str, str)
	}
	return len(str), nil
}

func newStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, "\n")
	reader := strings.NewReader(formattedString)
	return reader
}

func main() {
	fmt.Println("/////////////// Remote Command PoC ///////////////")

	// Parameters
	namespace := "default"
	podName := "bb"
	// note that command does not handle chaining commands
	command := []string{"ls", "-al"}
	// command := []string{"touch", "hello.txt"}

	// Initialize
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
	// TODO: what happens when the first container in the Pods list is not an ES container?
	containerName := pod.Spec.Containers[0].Name
	fmt.Printf("Container Name: %s\n", containerName)

	// Create the request
	execRequest := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	execRequest.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
		Stdin:     true,
	}, scheme.ParameterCodec)

	fmt.Printf("URL:\t%s\n", execRequest.URL())
	fmt.Printf("Executing command \"%s\" on container [%s] in pod [%s] \n", strings.Join(command, " "), containerName, podName)

	// Create the executor
	exec, _ := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())

	// Create the stream objects
	stdIn := newStringReader([]string{})
	stdOut := new(Writer)
	stdErr := new(Writer)

	// Execute the command
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdIn,
		Stdout: stdOut,
		Stderr: stdErr,
		Tty:    false,
	})

	// Display stdout and stderr
	fmt.Printf("[stdOut]\n%s\n", stdOut)
	if err != nil {
		fmt.Printf("[stdErr]\n%s\n", stdErr)
		fmt.Printf("[ERROR] %v\n", err)
	}
	fmt.Println("----------- FINISHED -------------")
}
