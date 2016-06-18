package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

type EndPoint struct {
	Destination string
}

type Node struct {
	Key    string
	NodeId string

	EndPoints []EndPoint
}

type EntryPoint struct {
	Domain string
	Path   string
	Node   *Node
}

type Manager struct {
	Nodes       []Node
	EntryPoints []EntryPoint
}

func GetStartPointForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<form method='post'/>")
	fmt.Fprintf(w, "Key: <input name='key' type='password'/><br/>")
	fmt.Fprintf(w, "Domain: <input name='domain'/><br/>")
	fmt.Fprintf(w, "EndPoint: <input name='endpoint'/><br/>")
	fmt.Fprintf(w, "<input type='submit'/>")
	fmt.Fprintf(w, "</form>")
}

func StartPointFData(w http.ResponseWriter, r *http.Request) *Node {
	mgmtPoint := Node{r.FormValue("key"), r.FormValue("domain"), make([]EndPoint, 0, 10)}
	mgmtPoint.EndPoints = append(mgmtPoint.EndPoints, EndPoint{r.FormValue("endpoint")})

	return &mgmtPoint
}

func GetEntryPointForm(w http.ResponseWriter, r *http.Request, manager *Manager) {
	fmt.Fprintf(w, "<form method='post'/>")
	fmt.Fprintf(w, "Domain: <input name='domain'/><br/>")
	fmt.Fprintf(w, "Path: <input name='path' value='/'><br/>")
	fmt.Fprintf(w, "Node: <select name='node'>")

	for _, node := range manager.Nodes {
		fmt.Fprintf(w, "<option value='%s'>%s</option>", node.NodeId, node.NodeId)
	}

	fmt.Fprintf(w, "</select>")
	fmt.Fprintf(w, "<input type='submit'/>")
	fmt.Fprintf(w, "</form>")
}

func EntryPointFData(w http.ResponseWriter, r *http.Request, manager *Manager) *EntryPoint {
	entryPoint := EntryPoint{r.FormValue("domain"), r.FormValue("path"), nil}
	for _, node := range manager.Nodes {
		if node.NodeId == r.FormValue("node") {
			entryPoint.Node = &node
			break
		}
	}

	return &entryPoint
}

func startNginx() {
	cmd := exec.Command("nginx", "-g", "daemon off;")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		log.Print(text)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal("Nginx died")
}

func reloadNginx() {
	cmd := exec.Command("nginx", "-s", "reload")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(stdout)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("RELOADDATA:")
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Print(err)
			break
		}
		log.Print(text)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Nginx reloaded")
}

func writeStartPoint(startPoint Node) {
	file, err := os.Create("/etc/nginx/conf.d/" + startPoint.NodeId + ".conf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "server {\n")
	fmt.Fprintf(file, "\tlisten 80;\n")
	fmt.Fprintf(file, "\tserver_name %s;\n", startPoint.NodeId)
	fmt.Fprintf(file, "\tlocation / {\n")
	for _, end := range startPoint.EndPoints {
		fmt.Fprintf(file, "\t\tproxy_pass %s;\n", end.Destination)
	}
	fmt.Fprintf(file, "\t}\n")
	fmt.Fprintf(file, "}\n")
}

func writeDefaultConfig() {

	b64data := `
dXNlciBuZ2lueDsKd29ya2VyX3Byb2Nlc3NlcyBhdXRvOwplcnJvcl9sb2cgL3Zhci9sb2cvbmdp
bngvZXJyb3IubG9nOwpwaWQgL3J1bi9uZ2lueC5waWQ7CgpldmVudHMgewogICAgd29ya2VyX2Nv
bm5lY3Rpb25zIDEwMjQ7Cn0KCmh0dHAgewogICAgbG9nX2Zvcm1hdCAgbWFpbiAgJyRyZW1vdGVf
YWRkciAtICRyZW1vdGVfdXNlciBbJHRpbWVfbG9jYWxdICIkcmVxdWVzdCIgJwogICAgICAgICAg
ICAgICAgICAgICAgJyRzdGF0dXMgJGJvZHlfYnl0ZXNfc2VudCAiJGh0dHBfcmVmZXJlciIgJwog
ICAgICAgICAgICAgICAgICAgICAgJyIkaHR0cF91c2VyX2FnZW50IiAiJGh0dHBfeF9mb3J3YXJk
ZWRfZm9yIic7CgogICAgYWNjZXNzX2xvZyAgL3Zhci9sb2cvbmdpbngvYWNjZXNzLmxvZyAgbWFp
bjsKCiAgICBzZW5kZmlsZSAgICAgICAgICAgIG9uOwogICAgdGNwX25vcHVzaCAgICAgICAgICBv
bjsKICAgIHRjcF9ub2RlbGF5ICAgICAgICAgb247CiAgICBrZWVwYWxpdmVfdGltZW91dCAgIDY1
OwogICAgdHlwZXNfaGFzaF9tYXhfc2l6ZSAyMDQ4OwoKICAgIGluY2x1ZGUgICAgICAgICAgICAg
L2V0Yy9uZ2lueC9taW1lLnR5cGVzOwogICAgZGVmYXVsdF90eXBlICAgICAgICBhcHBsaWNhdGlv
bi9vY3RldC1zdHJlYW07CgogICAgaW5jbHVkZSAvZXRjL25naW54L2NvbmYuZC8qLmNvbmY7Cgog
ICAgc2VydmVyIHsKICAgICAgICBsaXN0ZW4gICAgICAgODAgZGVmYXVsdF9zZXJ2ZXI7CiAgICAg
ICAgbGlzdGVuICAgICAgIFs6Ol06ODAgZGVmYXVsdF9zZXJ2ZXI7CiAgICAgICAgc2VydmVyX25h
bWUgIF87CgogICAgICAgICMgTG9hZCBjb25maWd1cmF0aW9uIGZpbGVzIGZvciB0aGUgZGVmYXVs
dCBzZXJ2ZXIgYmxvY2suCiAgICAgICAgaW5jbHVkZSAvZXRjL25naW54L2RlZmF1bHQuZC8qLmNv
bmY7CgogICAgfQp9Cgo=
`
	data, err := base64.StdEncoding.DecodeString(strings.Trim(b64data, " \n\r\t"))
	if err != nil {
		log.Fatal("Can't decode init config", err)
		return
	}

	ioutil.WriteFile("/etc/nginx/nginx.conf", data, 0644)

}

func writeMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "<a href='/add'>Add</a>\n")
	fmt.Fprint(w, "<a href='/entrypoint/add'>Add Entrypoint</a>\n")
	fmt.Fprint(w, "<a href='/listnodes'>List</a>\n")
	fmt.Fprint(w, "<a href='/valid'>Validate</a>\n")
	fmt.Fprint(w, "<a href='/update'>Update</a>\n")
	fmt.Fprint(w, "<br/>\n\n")
}

func handler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
}

func (manager *Manager) addHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method != "POST" {
		GetStartPointForm(w, r)
	} else {
		fmt.Fprintf(w, "You managed to post something... %s", r.FormValue("domain"))
		mgmtPoint := StartPointFData(w, r)

		manager.AddNode(mgmtPoint)
	}
}

func (manager *Manager) updateHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method == "POST" {
		mgmtPoint := StartPointFData(w, r)
		manager.UpdateNode(mgmtPoint)
	} else {
		GetStartPointForm(w, r)
	}
}

func updateConfiguration(manager *Manager) {
	file, err := os.Create("/etc/nginx/conf.d/risoxy.conf")
	if err != nil {
		log.Print(err)
		return
	}
	defer file.Close()

	for _, entryPoint := range manager.EntryPoints {

		fmt.Fprintf(file, "server {\n")
		fmt.Fprintf(file, "\tlisten 80;\n")
		fmt.Fprintf(file, "\tserver_name %s;\n", entryPoint.Domain)
		fmt.Fprintf(file, "\tlocation %s {\n", entryPoint.Path)

		fmt.Fprintf(file, "\t\tproxy_pass %s;\n", entryPoint.Node.EndPoints[0].Destination)

		fmt.Fprintf(file, "\t}\n")
		fmt.Fprintf(file, "}\n")
	}
	saveState(manager)
}

func (manager *Manager) addNode(node *Node, reload bool) {
	manager.Nodes = append(manager.Nodes, *node)

	if reload {
		updateConfiguration(manager)
		reloadNginx()
	}
}

func (manager *Manager) AddNode(node *Node) {
	manager.addNode(node, true)

}

func (manager *Manager) UpdateNode(newNode *Node) {
	for _, node := range manager.Nodes {
		if node.NodeId == newNode.NodeId {
			node.EndPoints[0] = newNode.EndPoints[0]
			writeStartPoint(node)
			reloadNginx()
			return
		}
	}
}

func (manager *Manager) listNodeHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	w.Header().Set("Content-Type", "text/html")

	for _, node := range manager.Nodes {
		fmt.Fprintf(w, "%s <a href='http://%s'>[w]</a><br/>\n", node.NodeId, node.NodeId)
		for _, endpoint := range node.EndPoints {
			fmt.Fprintf(w, " %s<br/>\n", endpoint.Destination)
		}
	}
}

func (manager *Manager) activeNodeHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method == "POST" {
		mgmtPoint := StartPointFData(w, r)
		for _, node := range manager.Nodes {
			if node.NodeId == mgmtPoint.NodeId && node.EndPoints[0].Destination == mgmtPoint.EndPoints[0].Destination {
				fmt.Fprintf(w, "OK")
				return
			} else if node.NodeId == mgmtPoint.NodeId {
				fmt.Fprintf(w, "Domain ok...")
				fmt.Fprintf(w, "%s = %s", node.EndPoints[0].Destination, mgmtPoint.EndPoints[0].Destination)
				return
			}
		}
		fmt.Fprintf(w, "Inactive")
		return
	} else {
		GetStartPointForm(w, r)
	}
}

func (manager *Manager) addEntryPointHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method == "POST" {
		entryPoint := EntryPointFData(w, r, manager)
		manager.EntryPoints = append(manager.EntryPoints, *entryPoint)
		updateConfiguration(manager)
		reloadNginx()
	}
	GetEntryPointForm(w, r, manager)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

type TOML_EndPoint struct {
	Destination string
}

type TOML_Node struct {
	Key      string
	NodeId   string
	EndPoint []TOML_EndPoint
}

type TOML_EntryPoint struct {
	Domain string
	Path   string
	NodeId string
}

type TOML_Manager struct {
	Version    string
	Node       []TOML_Node
	EntryPoint []TOML_EntryPoint
}

func saveState(manager *Manager) {
	var tManager TOML_Manager

	tManager.Version = "0.1"
	for _, node := range manager.Nodes {
		var tNode TOML_Node
		tNode.Key = node.Key
		tNode.NodeId = node.NodeId

		for _, endpoint := range node.EndPoints {
			var tEndPoint TOML_EndPoint
			tEndPoint.Destination = endpoint.Destination

			tNode.EndPoint = append(tNode.EndPoint, tEndPoint)

		}
		tManager.Node = append(tManager.Node, tNode)
	}

	for _, entry := range manager.EntryPoints {
		var tEntry TOML_EntryPoint
		tEntry.Domain = entry.Domain
		tEntry.Path = entry.Path
		tEntry.NodeId = entry.Node.NodeId

		tManager.EntryPoint = append(tManager.EntryPoint, tEntry)
	}

	file, err := os.Create("/etc/loadmanager/state/risoxy.state")
	failOnError(err, "Unable to save state file")
	defer file.Close()
	enc := toml.NewEncoder(file)
	enc.Encode(tManager)
}

func loadState(filename string) Manager {
	_, err := os.Stat(filename)
	failOnError(err, "Config file missing")

	var tManager TOML_Manager
	_, err = toml.DecodeFile(filename, &tManager)
	failOnError(err, "")

	var manager Manager

	for _, tNode := range tManager.Node {
		var node Node
		node.Key = tNode.Key
		node.NodeId = tNode.NodeId

		for _, tEndPoint := range tNode.EndPoint {
			var endpoint EndPoint
			endpoint.Destination = tEndPoint.Destination

			node.EndPoints = append(node.EndPoints, endpoint)
		}

		manager.Nodes = append(manager.Nodes, node)
	}

	for _, tEntry := range tManager.EntryPoint {
		var entry EntryPoint
		entry.Domain = tEntry.Domain
		entry.Path = tEntry.Path

		for _, node := range manager.Nodes {
			if node.NodeId == tEntry.NodeId {
				entry.Node = &node
				break
			}
		}

		manager.EntryPoints = append(manager.EntryPoints, entry)
	}

	return manager
}

func LoadManager(filename string) Manager {
	manager := loadState(filename)
	return manager
}

func main() {
	writeDefaultConfig()
	manager := LoadManager("/etc/loadmanager/state/risoxy.state")

	updateConfiguration(&manager)

	go startNginx()

	http.HandleFunc("/", handler)
	http.HandleFunc("/add", manager.addHandler)
	http.HandleFunc("/update", manager.updateHandler)
	http.HandleFunc("/entrypoint/add", manager.addEntryPointHandler)
	http.HandleFunc("/listnodes", manager.listNodeHandler)
	http.HandleFunc("/valid", manager.activeNodeHandler)
	http.ListenAndServe(":8080", nil)
}
