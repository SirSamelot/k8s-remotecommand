package main

import (
	"flag"
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

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func main() {
	fmt.Println("/////////////// Remote Command PoC ///////////////")

	// get the kubeconfig
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	kubeClient := kubernetes.NewForConfigOrDie(config)

	namespace := "default"
	podName := "bb"

	// Testing that we are talking to kubernetes
	pods, _ := kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	fmt.Printf("There are %d pods in the namespace '%s'\n", len(pods.Items), namespace)

	pod, _ := kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	containerName := pod.Spec.Containers[0].Name
	fmt.Printf("Container Name: %s\n", containerName)

	// create the request
	execRequest := kubeClient.CoreV1().RESTClient().Post()
	execRequest = execRequest.
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	execRequest.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   []string{"ls", "-al"},
		Stdout:    true,
		Stderr:    true,
		Stdin:     true,
	}, scheme.ParameterCodec)

	fmt.Printf("URL:\t%s\n", execRequest.URL())
	// Create the executor
	exec, _ := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())

	stdIn := newStringReader([]string{"-c", "-al"})
	stdOut := new(Writer)
	stdErr := new(Writer)

	// Stream the command
	_ = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdIn,
		Stdout: stdOut,
		Stderr: stdErr,
		Tty:    false,
	})
	fmt.Printf("stdIn:\t%s\nstdOut:\t%s\nstdErr:\t%s\n", stdIn, stdOut, stdErr)

	fmt.Println("----------- FINISHED -------------")
}
