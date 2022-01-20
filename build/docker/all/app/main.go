package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/apoorvam/goterminal"
	"github.com/olekukonko/tablewriter"
	"github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"
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

	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(30*time.Second),
		wait.WithDebug(true),
	).Do([]string{"127.0.0.1:6443"}) {
		log.Fatalf("k3s is not available")
		return
	}

	log.Println("unzip images")
	cmd := exec.Command("/bin/gunzip", "-v", "/images.tar.gz")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	log.Println("untar images")
	cmd = exec.Command("/bin/tar", "-xvf", "/images.tar")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	// delete because we import all tars as images
	os.Remove("/images.tar")

	ff, err := os.ReadDir("/")
	if err != nil {
		log.Fatalf("error reading dir")
	}

	log.Println("importing images")
	for i := range ff {
		f := ff[i]
		if strings.HasSuffix(f.Name(), ".tar") {
			log.Printf("importing %v", f.Name())
			importImage(f.Name())
		}
	}

	installKnative(kc)

	runRegistry(kc)

	runDB(kc)

	runHelm(kc)

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
			time.Sleep(2 * time.Second)
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
		time.Sleep(1 * time.Second)
		writer.Clear()
	}

	writer.Reset()

	fmt.Println("direktiv connecting services, please wait")
	for {
		res, err := http.Get("http://localhost/api/namespaces")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// if api key set it is unauthorized but that is fine too
		if res.StatusCode == 200 || res.StatusCode == 401 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("direktiv ready at http://<HOST-IP>:8080")

	select {}

}

func importImage(img string) {

	cmd := exec.Command("/k3s", "ctr", "images", "import", img)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	os.Remove(img)

}

func runDB(kc string) {
	log.Println("deploying db")

	cmd := exec.Command(kc, "apply", "-f", "/pg")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

}

type byName []v1.Pod

func (a byName) Len() int           { return len(a) }
func (a byName) Less(i, j int) bool { return a[i].GetName() < a[j].GetName() }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func runHelm(kc string) {

	f, err := os.OpenFile("/debug.yaml", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	/* #nosec */
	defer f.Close()

	log.Printf("adding apikey if configured\n")
	if os.Getenv("APIKEY") != "" {
		if _, err = f.Write([]byte(fmt.Sprintf("\napikey: \"%v\"\n", os.Getenv("APIKEY")))); err != nil {
			fmt.Printf("could not add api key: %v", err)
		}
	}

	addProxy(f)

	log.Printf("creating service namespace\n")
	/* #nosec */
	cmd := exec.Command(kc, "create", "namespace", "direktiv-services-direktiv")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	log.Printf("running direktiv helm\n")
	cmd = exec.Command("/helm", "install", "-f", "/debug.yaml", "direktiv", ".")
	cmd.Dir = "/direktiv-charts/charts/direktiv"
	cmd.Env = []string{"KUBECONFIG=/etc/rancher/k3s/k3s.yaml"}
	cmd.Run()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

}

func runRegistry(kc string) {

	log.Println("installing docker repository")
	log.Printf("applying registry.yaml\n")
	/* #nosec */
	cmd := exec.Command(kc, "apply", "-f", "/registry.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

}

func addProxy(f *os.File) {

	log.Printf("adding proxy settings to file: %v\n", f.Name())
	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("HTTP_PROXY") != "" {
		log.Println("http proxy set")
		f.Write([]byte(fmt.Sprintf("https_proxy: \"%v\"\n", os.Getenv("HTTPS_PROXY"))))
		f.Write([]byte(fmt.Sprintf("http_proxy: \"%v\"\n", os.Getenv("HTTP_PROXY"))))
		f.Write([]byte(fmt.Sprintf("no_proxy: \"%v\"\n", os.Getenv("NO_PROXY"))))
	} else {
		log.Println("http proxy not set")
		f.Write([]byte("http_proxy: \"\"\n"))
		f.Write([]byte("https_proxy: \"\"\n"))
		f.Write([]byte("no_proxy: \"\"\n"))
	}
}

func installKnative(kc string) {

	log.Printf("running knative helm\n")

	/* #nosec */
	f, err := os.Create("/tmp/knative.yaml")

	/* #nosec */
	defer f.Close()

	if err != nil {
		panic(err)
	}

	addProxy(f)

	cmd := exec.Command("/helm", "install", "-n", "knative-serving", "--create-namespace", "-f", "/tmp/knative.yaml", "knative", ".")
	cmd.Dir = "/direktiv-charts/charts/knative"
	cmd.Env = []string{"KUBECONFIG=/etc/rancher/k3s/k3s.yaml"}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	isgcp := isGCP()
	log.Printf("running on GCP: %v\n", isgcp)
	if isgcp {
		/* #nosec */
		cmd := exec.Command(kc, "apply", "-f", "/google-dns.yaml")
		cmd.Dir = "/"
		cmd.Run()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if os.Getenv("EVENTING") != "" {
		log.Printf("installing knative eventing")

		yamls := []string{
			"https://github.com/knative/eventing/releases/download/v0.26.1/eventing-crds.yaml",
			"https://github.com/knative/eventing/releases/download/v0.26.1/eventing-core.yaml",
			"https://github.com/knative/eventing/releases/download/v0.26.1/mt-channel-broker.yaml",
			"https://github.com/knative/eventing/releases/download/v0.26.1/in-memory-channel.yaml",
		}

		for i := range yamls {
			cmd = exec.Command(kc, "apply", "-f", yamls[i])
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

		// waiting for controller to be Ready
		cmd = exec.Command(kc, "wait", "--for=condition=ready", "pod", "-l", "app=mt-broker-controller", "-n", "knative-eventing")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		cmd = exec.Command(kc, "apply", "-f", "/broker.yaml")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	}

}

func startingK3s() error {

	log.Println("starting k3s now")
	cmd := exec.Command("/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0",
		"--disable", "traefik", "--write-kubeconfig-mode=644", "--kube-apiserver-arg",
		"feature-gates=TTLAfterFinished=true")

	log.Println("evacuate cgroup2")
	if err := cgrouputil.EvacuateCgroup2("init"); err != nil {
		log.Println("could not evacuate cgroup2")
		return err
	}

	// passing env in for http_prox values
	cmd.Env = os.Environ()

	if len(os.Getenv("DEBUG")) > 0 {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()

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
