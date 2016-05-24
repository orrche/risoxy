package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type EndPoint struct {
	Destination string
	Start       *StartPoint
}

type StartPoint struct {
	Key    string
	Domain string

	EndPoint []EndPoint
}

type Manager struct {
	Nodes []StartPoint
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

	log.Print("DATA:")
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

func writeStartPoint(startPoint StartPoint) {
	file, err := os.Create("/etc/nginx/conf.d/" + startPoint.Domain + ".conf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintf(file, "server {\n")
	fmt.Fprintf(file, "\tlisten 80;\n")
	fmt.Fprintf(file, "\tserver_name %s;\n", startPoint.Domain)
	fmt.Fprintf(file, "\tlocation /{\n")
	for _, end := range startPoint.EndPoint {
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
		fmt.Fprintf(w, "<form method='post'/>")
		fmt.Fprintf(w, "Key: <input name='key' type='password'/><br/>")
		fmt.Fprintf(w, "Domain: <input name='domain'/><br/>")
		fmt.Fprintf(w, "EndPoint: <input name='endpoint'/><br/>")
		fmt.Fprintf(w, "<input type='submit'/>")
		fmt.Fprintf(w, "</form>")
	} else {
		fmt.Fprintf(w, "You managed to post something... %s", r.FormValue("domain"))
		mgmtPoint := StartPoint{r.FormValue("key"), r.FormValue("domain"), make([]EndPoint, 0, 10)}
		mgmtPoint.EndPoint = append(mgmtPoint.EndPoint, EndPoint{r.FormValue("endpoint"), &mgmtPoint})

		manager.AddNode(&mgmtPoint)
	}
}

func (manager *Manager) updateHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method == "POST" {
	} else {
		fmt.Fprintf(w, "<form method='post'/>")
		fmt.Fprintf(w, "Key: <input name='key' type='password'/><br/>")
		fmt.Fprintf(w, "Domain: <input name='domain'/><br/>")
		fmt.Fprintf(w, "Endpoint: <input name='endpoint'/><br/>")
		fmt.Fprintf(w, "<input type='submit'/>")
		fmt.Fprintf(w, "</form>")
	}
}

func (manager *Manager) addNode(node *StartPoint, reload bool) {
	manager.Nodes = append(manager.Nodes, *node)

	writeStartPoint(*node)
	if reload {
		reloadNginx()
	}
}

func (manager *Manager) AddNode(node *StartPoint) {
	manager.addNode(node, true)

}

func (manager *Manager) UpdateNode(newNode *StartPoint) {
	for _, node := range manager.Nodes {
		if node.Domain == newNode.Domain {
			node.EndPoint[0] = newNode.EndPoint[0]
		}
	}
}

func (manager *Manager) listNodeHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	w.Header().Set("Content-Type", "text/html")

	for _, node := range manager.Nodes {
		fmt.Fprintf(w, "%s <a href='http://%s'>[w]</a><br/>\n", node.Domain, node.Domain)
		for _, endpoint := range node.EndPoint {
			fmt.Fprintf(w, " %s<br/>\n", endpoint.Destination)
		}
	}
}

func (manager *Manager) activeNodeHandler(w http.ResponseWriter, r *http.Request) {
	writeMenu(w, r)
	if r.Method == "POST" {
	} else {
		fmt.Fprintf(w, "<form method='post'/>")
		fmt.Fprintf(w, "Key: <input name='key' type='password'/><br/>")
		fmt.Fprintf(w, "Domain: <input name='domain'/><br/>")
		fmt.Fprintf(w, "Endpoint: <input name='endpoint'/><br/>")
		fmt.Fprintf(w, "<input type='submit'/>")
		fmt.Fprintf(w, "</form>")
	}
}

func LoadManager(filename string) Manager {
	var manager Manager
	configFile, err := os.Open(filename)
	if err != nil {
		log.Print("Config file read error: ", err)
		return manager
	}
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&manager); err != nil {
		log.Print("Parse error", err)
	}
	return manager
}

func main() {
	writeDefaultConfig()
	manager := LoadManager("/etc/loadmanager/loadmanager.conf")

	for _, node := range manager.Nodes {
		writeStartPoint(node)
	}
	go startNginx()

	http.HandleFunc("/", handler)
	http.HandleFunc("/add", manager.addHandler)
	http.HandleFunc("/update", manager.updateHandler)
	http.HandleFunc("/listnodes", manager.listNodeHandler)
	http.HandleFunc("/valid", manager.activeNodeHandler)
	http.ListenAndServe("localhost:8080", nil)

}
