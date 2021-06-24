package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/apoorvam/goterminal"
	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	log.Println("all-in-one version of direktiv")

	kc, err := exec.LookPath("kubectl")
	if err != nil {
		panic(err.Error())
	}

	changeContour()

	go func() {
		err := startingK3s()
		if err != nil {
			panic(err.Error())
		}
	}()

	log.Println("waiting for k3s kubeconfig")
	for {
		if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); !os.IsNotExist(err) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	log.Println("installing direktiv")

	applyYaml(kc)
	patch(kc)
	runHelm()

	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	writer := goterminal.New(os.Stdout)
	n := time.Now()

	log.Println("waiting for pods (can take several minutes)")

	var lo metav1.ListOptions
	for {

		table := tablewriter.NewWriter(writer)
		table.SetHeader([]string{"Pod", "Status", "Time"})

		pods, err := clientset.CoreV1().Pods("").List(context.Background(), lo)
		if err != nil {
			panic(err.Error())
		}

		if len(pods.Items) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		allRun := true

		sort.Sort(byName(pods.Items))
		for _, pod := range pods.Items {

			t := "ready"
			if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodSucceeded {
				allRun = false
				t = fmt.Sprintf("%vs", int(time.Since(n).Seconds()))
			}
			table.Append([]string{pod.GetName(), string(pod.Status.Phase), t})
		}

		table.Render()
		writer.Print()
		if allRun && len(pods.Items) > 0 {
			break
		}
		time.Sleep(5 * time.Second)
		writer.Clear()
	}

	writer.Reset()
	fmt.Println("direktiv ready at http://localhost:8080")

	select {}

}

type byName []v1.Pod

func (a byName) Len() int           { return len(a) }
func (a byName) Less(i, j int) bool { return a[i].GetName() < a[j].GetName() }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func runHelm() {

	if os.Getenv("PERSIST") != "" {

		f, err := os.OpenFile("/debug.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if _, err := f.WriteString("supportPersist: true\n"); err != nil {
			panic(err)
		}

	}

	log.Printf("running direktiv helm\n")
	cmd := exec.Command("/helm", "install", "-f", "/debug.yaml", "direktiv", ".")
	cmd.Dir = "/direktiv/kubernetes/charts/direktiv"
	cmd.Env = []string{"KUBECONFIG=/etc/rancher/k3s/k3s.yaml"}
	cmd.Run()

}

func applyYaml(kc string) {

	fs := []string{"serving-crds.yaml", "serving-core.yaml", "contour.yaml", "net-contour.yaml"}

	for _, f := range fs {
		log.Printf("applying %s\n", f)
		cmd := exec.Command(kc, "apply", "-f", f)
		cmd.Dir = "/direktiv/scripts/knative"
		cmd.Run()
		time.Sleep(2 * time.Second)
	}

	isgcp := isGCP()
	log.Printf("running on GCP: %v\n", isgcp)
	if isgcp {
		cmd := exec.Command(kc, "apply", "-f", "/google-dns.yaml")
		cmd.Dir = "/"
		cmd.Run()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

}

func patch(kc string) {

	log.Printf("patching configmap\n")
	cmd := exec.Command(kc, "patch", "configmap/config-network",
		"--namespace", "knative-serving", "--type", "merge", "--patch",
		"{\"data\":{\"ingress.class\":\"contour.ingress.networking.knative.dev\"}}")
	cmd.Run()

}

func startingK3s() error {

	log.Println("starting k3s now")
	cmd := exec.Command("k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0", "--disable", "traefik", "--write-kubeconfig-mode=644")

	if len(os.Getenv("DEBUG")) > 0 {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()

}

func changeContour() {
	iyaml, err := ioutil.ReadFile("/direktiv/scripts/knative/contour.yaml")
	if err != nil {
		panic(err.Error())
	}

	output := bytes.Replace(iyaml, []byte("replicas: 2"), []byte("replicas: 1"), -1)

	if err = ioutil.WriteFile("/direktiv/scripts/knative/contour.yaml", output, 0666); err != nil {
		panic(err.Error())
	}
}

func isGCP() bool {

	data, err := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return false
	}
	name := strings.TrimSpace(string(data))

	if name == "Google" || name == "Google Compute Engine" {
		return true
	}

	return false

}
