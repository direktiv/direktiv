package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"

	// "github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"
	v1 "k8s.io/api/core/v1"
)

func waitForUp() {

	log.Println("waiting for k3s kubeconfig")
	for {
		if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); !os.IsNotExist(err) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// run kubectl command tiull it is successful
	for {
		cmd := exec.Command("k3s", "kubectl", "get", "pods")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err == nil {
			break
		}
	}

	log.Println("k3s up")

}

func main() {

	log.Println("all-in-one version of direktiv")

	go func() {
		err := startingK3s()
		if err != nil {
			panic(err.Error())
		}
	}()

	waitForUp()

	installKnative()

	runRegistry()

	runDB()

	// runHelm(kc)

	// config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	// if err != nil {
	// 	panic(err.Error())
	// }

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// writer := goterminal.New(os.Stdout)
	// n := time.Now()

	// log.Println("waiting for pods (can take several minutes)")

	// var lo metav1.ListOptions
	// for {

	// 	table := tablewriter.NewWriter(writer)
	// 	table.SetHeader([]string{"Pod", "Status", "Time"})

	// 	pods, err := clientset.CoreV1().Pods("").List(context.Background(), lo)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}

	// 	if len(pods.Items) == 0 {
	// 		time.Sleep(2 * time.Second)
	// 		continue
	// 	}

	// 	allRun := true

	// 	sort.Sort(byName(pods.Items))
	// 	for _, pod := range pods.Items {

	// 		t := "ready"
	// 		if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodSucceeded {
	// 			allRun = false
	// 			t = fmt.Sprintf("%vs", int(time.Since(n).Seconds()))
	// 		}
	// 		table.Append([]string{pod.GetName(), string(pod.Status.Phase), t})
	// 	}

	// 	table.Render()
	// 	writer.Print()
	// 	if allRun && len(pods.Items) > 0 {
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// 	writer.Clear()
	// }

	// writer.Reset()

	// fmt.Println("direktiv connecting services, please wait")
	// for {
	// 	res, err := http.Get("http://localhost/api/namespaces")
	// 	if err != nil {
	// 		time.Sleep(1 * time.Second)
	// 		continue
	// 	}

	// 	// if api key set it is unauthorized but that is fine too
	// 	if res.StatusCode == 200 || res.StatusCode == 401 {
	// 		break
	// 	}
	// 	time.Sleep(1 * time.Second)
	// }

	// fmt.Println("direktiv ready at http://localhost:8080 or http://<HOST-IP>:8080")

	select {}

}

func runDB() {
	log.Println("deploying db")

	cmd := exec.Command("k3s", "kubectl", "apply", "-f", "/pg")
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
	defer f.Close()

	log.Printf("adding apikey if configured\n")
	if os.Getenv("APIKEY") != "" {
		if _, err = f.Write([]byte(fmt.Sprintf("\napikey: \"%v\"\n", os.Getenv("APIKEY")))); err != nil {
			fmt.Printf("could not add api key: %v", err)
		}
	}

	addProxy(f)

	log.Printf("creating service namespace\n")

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

func runRegistry() {

	log.Println("installing docker repository")
	log.Printf("applying registry.yaml\n")

	cmd := exec.Command("k3s", "kubectl", "apply", "-f", "/registry.yaml")
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

func installKnative() {

	log.Printf("1!!!!!!!!!!!!!!!!!!!!running knative helm\n")

	time.Sleep(60 * time.Second)

	log.Printf("2!!!!!!!!!!!!!!!!!!!!running knative helm\n")

	f, err := os.Create("/tmp/knative.yaml")
	defer f.Close()

	if err != nil {
		panic(err)
	}

	addProxy(f)

	cmd := exec.Command("/helm", "install", "-n", "knative-serving", "--create-namespace", "-f", "/tmp/knative.yaml", "knative", "/direktiv-charts/charts/knative")
	cmd.Dir = "/direktiv-charts/charts/knative"
	cmd.Env = []string{"KUBECONFIG=/etc/rancher/k3s/k3s.yaml"}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Run()

	isgcp := isGCP()
	log.Printf("running on GCP: %v\n", isgcp)
	if isgcp {
		cmd := exec.Command("k3s", "kubectl", "apply", "-f", "/google-dns.yaml")
		cmd.Dir = "/"
		cmd.Run()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if os.Getenv("EVENTING") != "" {
		log.Printf("installing knative eventing")

		// yamls := []string{
		// 	"https://github.com/knative/eventing/releases/download/v0.26.1/eventing-crds.yaml",
		// 	"https://github.com/knative/eventing/releases/download/v0.26.1/eventing-core.yaml",
		// 	"https://github.com/knative/eventing/releases/download/v0.26.1/mt-channel-broker.yaml",
		// 	"https://github.com/knative/eventing/releases/download/v0.26.1/in-memory-channel.yaml",
		// }

		// for i := range yamls {
		// 	cmd = exec.Command(kc, "apply", "-f", yamls[i])
		// 	cmd.Stdout = os.Stdout
		// 	cmd.Stderr = os.Stderr
		// 	cmd.Run()
		// }

		// // waiting for controller to be Ready
		// cmd = exec.Command(kc, "wait", "--for=condition=ready", "pod", "-l", "app=mt-broker-controller", "-n", "knative-eventing")
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		// cmd.Run()

		// cmd = exec.Command(kc, "apply", "-f", "/broker.yaml")
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		// cmd.Run()

	}

}

func startingK3s() error {

	log.Println("starting k3s now")
	cmd := exec.Command("/usr/local/bin/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0",
		"--disable", "traefik", "--write-kubeconfig-mode=644", "--kube-apiserver-arg",
		"feature-gates=TTLAfterFinished=true")

	log.Println("evacuate cgroup2")
	if err := cgrouputil.EvacuateCgroup2("init"); err != nil {
		log.Println("could not evacuate cgroup2")
		return err
	}

	// passing env in for http_prox values
	cmd.Env = os.Environ()

	// if len(os.Getenv("DEBUG")) > 0 {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// }

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
