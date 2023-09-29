package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apoorvam/goterminal"
	"github.com/olekukonko/tablewriter"
	"github.com/rootless-containers/rootlesskit/pkg/parent/cgrouputil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/scale/scheme"
	"k8s.io/client-go/tools/clientcmd"

	"knative.dev/operator/pkg/apis/operator/base"
	ks "knative.dev/operator/pkg/apis/operator/v1beta1"
)

func waitForUp() {
	// first wait for the kubeconfig file
	log.Println("waiting for k3s kubeconfig")
	for {
		if _, err := os.Stat("/etc/rancher/k3s/k3s.yaml"); !os.IsNotExist(err) {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// run kubectl command till it is successful, api seems to work
	log.Println("waiting for kubernetes API")
	for {
		cmd := exec.Command("k3s", "kubectl", "get", "nodes")
		err := cmd.Run()
		if err == nil {
			break
		}
	}

	// wait for kube-system pods to be up
	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	log.Println("waiting for systems pods")
	for {
		var lo metav1.ListOptions
		pods, err := clientset.CoreV1().Pods("kube-system").List(context.Background(), lo)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if len(pods.Items) == 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		break

	}

	log.Println("k3s up")
}

func main() {
	log.Println("all-in-one version of direktiv")

	b, err := os.ReadFile("/proc/sys/fs/inotify/max_user_instances")
	if err != nil {
		fmt.Print(err)
	}

	c := strings.ReplaceAll(string(b), "\n", "")

	files, err := strconv.Atoi(c)
	if err != nil {
		panic(err.Error())
	}

	if files < 4096 {
		panic("not enough inotifies. Set on host with 'sudo sysctl fs.inotify.max_user_instances=4096'")
	}

	go func() {
		err := startingK3s()
		if err != nil {
			panic(err.Error())
		}
	}()

	waitForUp()

	// if that file exists, the installation has already happened
	_, err = os.Stat("/.service")
	if err != nil {

		installKnative()

		runRegistry()

		runDB()

		runHelm()

		f, _ := os.Create("/.service")
		f.Close()

	}

	config, err := clientcmd.BuildConfigFromFlags("", "/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	writer := goterminal.New(os.Stdout)
	n := time.Now().UTC()

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

	err = addNamespace()
	if err != nil {
		fmt.Printf("could not add namespace from git: %v\n", err)
	}

	fmt.Println("direktiv ready at http://localhost:8080 or http://<HOST-IP>:8080")

	select {}
}

func addNamespace() error {
	fmt.Println("adding namespace from git")

	ns := make(map[string]string)
	ns["url"] = "https://github.com/direktiv/direktiv-examples.git"
	ns["ref"] = "main"

	client := &http.Client{}
	json, err := json.Marshal(ns)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, "http://localhost/api/namespaces/examples", bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	if os.Getenv("APIKEY") != "" {
		req.Header.Set("direktiv-token", os.Getenv("APIKEY"))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	time.Sleep(2 * time.Second)
	req, err = http.NewRequest(http.MethodPost, "http://localhost/api/namespaces/examples/tree?op=sync-mirror&force=true", nil)
	if err != nil {
		return err
	}

	if os.Getenv("APIKEY") != "" {
		req.Header.Set("direktiv-token", os.Getenv("APIKEY"))
	}

	respSync, err := client.Do(req)
	if err != nil {
		fmt.Printf("could not sync git repo: %v\n", err)
	}
	defer respSync.Body.Close()

	return nil
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

func runHelm() {
	fmt.Println("deploying direktiv")

	b, err := os.ReadFile("/debug.yaml")
	if err != nil {
		panic(err)
	}

	y := string(b)

	log.Printf("adding apikey if configured\n")
	if os.Getenv("APIKEY") != "" {
		y = y + "\n\n" + fmt.Sprintf("\napikey: \"%v\"\n", os.Getenv("APIKEY"))
	}

	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("HTTP_PROXY") != "" {

		add := fmt.Sprintf("https_proxy: \"%v\"\n", os.Getenv("HTTPS_PROXY"))
		add = add + "  " + fmt.Sprintf("http_proxy: \"%v\"\n", os.Getenv("HTTP_PROXY"))
		add = add + "  " + fmt.Sprintf("no_proxy: \"%v\"\n", os.Getenv("NO_PROXY"))

		y = strings.Replace(y, "PROXY", add, 1)

		add = fmt.Sprintf("https_proxy: \"%v\"\n", os.Getenv("HTTPS_PROXY"))
		add = add + fmt.Sprintf("http_proxy: \"%v\"\n", os.Getenv("HTTP_PROXY"))
		add = add + fmt.Sprintf("no_proxy: \"%v\"\n", os.Getenv("NO_PROXY"))
		y = y + "\n\n" + add

	} else {
		y = strings.Replace(y, "PROXY", "", 1)
	}

	err = os.WriteFile("/direktiv.yaml", []byte(y), 0o755)
	if err != nil {
		panic(err)
	}

	log.Printf("running direktiv helm\n")
	cmd := exec.Command("helm", "install", "-f", "/direktiv.yaml", "direktiv", ".")
	cmd.Dir = "/direktiv-charts/charts/direktiv"
	cmd.Env = []string{"KUBECONFIG=/etc/rancher/k3s/k3s.yaml"}
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func runRegistry() {
	log.Println("installing docker repository")
	log.Printf("applying registry.yaml\n")

	cmd := exec.Command("k3s", "kubectl", "apply", "-f", "/registry.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func waitForPod(pod, ns string) {
	log.Printf("waiting for %s in %s", pod, ns)
	cmd := exec.Command("k3s", "kubectl", "wait", "--for=condition=ready", "pod", "-l", pod, "-n", ns)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func downloadFile(url string) ([]byte, error) {
	log.Printf("downloading %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func createDecoder() runtime.Decoder {
	sch := runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = ks.AddToScheme(sch)
	return serializer.NewCodecFactory(sch).UniversalDeserializer()
}

func installKnative() {
	log.Printf("installing knative\n")

	var buf bytes.Buffer

	cmd := exec.Command("k3s", "kubectl", "apply", "-f", "https://github.com/knative/operator/releases/download/knative-v1.11.6/operator.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	waitForPod("name=knative-operator", "default")

	b, err := downloadFile("https://raw.githubusercontent.com/direktiv/direktiv/main/kubernetes/install/knative/basic.yaml")
	if err != nil {
		panic(fmt.Sprintf("can not download knative yaml: %s", err.Error()))
	}

	if os.Getenv("HTTPS_PROXY") != "" || os.Getenv("HTTP_PROXY") != "" {

		decoder := createDecoder()
		y := printers.YAMLPrinter{}

		obj, _, err := decoder.Decode([]byte(b), nil, nil)
		if err != nil {
			log.Print(err)
		}

		depl := obj.(*ks.KnativeServing)

		depl.Spec.DeploymentOverride = depl.Spec.DeploymentOverride[:0]

		activator := base.WorkloadOverride{
			Name:        "activator",
			Annotations: make(map[string]string, 0),
		}
		activator.Annotations["linkerd.io/inject"] = "enabled"
		depl.Spec.DeploymentOverride = append(depl.Spec.DeploymentOverride, activator)

		controller := base.WorkloadOverride{
			Name:        "controller",
			Annotations: make(map[string]string, 0),
			Env:         make([]base.EnvRequirementsOverride, 0),
		}
		controller.Annotations["linkerd.io/inject"] = "enabled"

		envOver := base.EnvRequirementsOverride{
			Container: "controller",
			EnvVars:   make([]v1.EnvVar, 0),
		}

		envOver.EnvVars = append(envOver.EnvVars, v1.EnvVar{
			Name:  "HTTP_PROXY",
			Value: os.Getenv("HTTP_PROXY"),
		})
		envOver.EnvVars = append(envOver.EnvVars, v1.EnvVar{
			Name:  "HTTPS_PROXY",
			Value: os.Getenv("HTTPS_PROXY"),
		})
		envOver.EnvVars = append(envOver.EnvVars, v1.EnvVar{
			Name:  "NO_PROXY",
			Value: os.Getenv("NO_PROXY"),
		})

		controller.Env = append(controller.Env, envOver)

		depl.Spec.DeploymentOverride = append(depl.Spec.DeploymentOverride, controller)

		y.PrintObj(depl, &buf)

		os.WriteFile("/knative.yaml", buf.Bytes(), 0o755)

	} else {
		os.WriteFile("/knative.yaml", b, 0o755)
	}

	// create namespace
	cmd = exec.Command("k3s", "kubectl", "create", "ns", "knative-serving")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("can not create namespace knative-serving: %s", err.Error()))
	}

	// create knative instance
	cmd = exec.Command("k3s", "kubectl", "apply", "-f", "/knative.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("can not deploy knative-serving: %s", err.Error()))
	}

	isgcp := isGCP()
	log.Printf("running on GCP: %v\n", isgcp)
	if isgcp {
		cmd := exec.Command("k3s", "kubectl", "apply", "-f", "/google-dns.yaml")
		cmd.Dir = "/"
		cmd.Run()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// contour
	cmd = exec.Command("k3s", "kubectl", "apply", "-f", "https://github.com/knative/net-contour/releases/download/knative-v1.11.0/contour.yaml")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("can not deploy contour: %s", err.Error()))
	}

	// delete namespace contour-external in background
	cmd = exec.Command("k3s", "kubectl", "delete", "ns", "contour-external")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		panic(fmt.Sprintf("can not deploy contour: %s", err.Error()))
	}

	if os.Getenv("EVENTING") != "" {
		log.Printf("installing knative eventing")

		cmd = exec.Command("k3s", "kubectl", "create", "ns", "knative-eventing")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			panic(fmt.Sprintf("can not create namespace knative-serving: %s", err.Error()))
		}

		cmd = exec.Command("k3s", "kubectl", "apply", "-f", "/eventing.yaml")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		// waiting for controller to be Ready
		cmd = exec.Command("k3s", "kubectl", "wait", "--for=condition=ready", "pod", "-l", "app=mt-broker-controller", "-n", "knative-eventing")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

		cmd = exec.Command("k3s", "kubectl", "apply", "-f", "/broker.yaml")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()

	}
}

func startingK3s() error {
	log.Println("starting k3s now")
	cmd := exec.Command("/usr/local/bin/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0",
		"--disable", "traefik", "--write-kubeconfig-mode=644", "--snapshotter", "native")

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
