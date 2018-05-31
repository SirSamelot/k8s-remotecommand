package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	// Testing access to API
	pods, _ := kubeClient.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	fmt.Printf("There are %d pods in the namespace '%s'\n", len(pods.Items), namespace)

	pod, _ := kubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	fmt.Println(pod)
	// HERE up to line 44 of searchlight/lib.go

	// execRequest := kubeClient.CoreV1().RESTClient().Post().
	// 	Resource("pods").
	// 	Name("bb").
	// 	Namespace("default").
	// 	SubResource("exec").
	// 	Param("container", "bb").
	// 	Param("command", "ls").
	// 	Param("stdin", "true").
	// 	Param("stdout", "true").
	// 	Param("stderr", "false").
	// 	Param("tty", "false")

	// // Here we go
	// exec, _ := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())

	// stdIn := newStringReader([]string{"-c", ""})
	// stdOut := new(Writer)
	// stdErr := new(Writer)

	// _ = exec.Stream(remotecommand.StreamOptions{
	// 	Stdin:  stdIn,
	// 	Stdout: stdOut,
	// 	Stderr: stdErr,
	// 	Tty:    false,
	// })

	// fmt.Println(stdOut)
}
