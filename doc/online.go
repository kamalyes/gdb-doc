package doc

import (
	"errors"
	"fmt"
	"gdb-doc/model"
	"gdb-doc/util"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
)

// createOnlineDoc create _siderbar.md
func createOnlineDoc(docPath string, dbInfo model.DbInfo, tables []model.Table) {
	var sidebar []string
	var readme []string
	// sidebar = append(sidebar, "* [数据库文档](README.md)")
	readme = append(readme, fmt.Sprintf("# %s 数据库文档", dbInfo.DbName))
	// 生成基础信息
	readme = append(readme, "### 基础信息")
	readme = append(readme, "| 数据库名称 | 版本 | 字符集 | 排序规则 | Table数量 |")
	readme = append(readme, "| ---- | ---- | ---- | ---- | ---- |")
	readme = append(readme, fmt.Sprintf("| %s | %s | %s | %s | %s |", dbInfo.DbName, dbInfo.Version, dbInfo.Charset, dbInfo.Collation, dbInfo.TableNumber))
	for i := range tables {
		sidebar = append(sidebar, fmt.Sprintf("* [%s(%s)](%s.md)", tables[i].TableName, tables[i].TableComment, tables[i].TableName))
		var tableMd []string
		tableMd = append(tableMd, fmt.Sprintf("# %s(%s)", tables[i].TableName, tables[i].TableComment))
		tableMd = append(tableMd, "| 列名 | 类型 | KEY | 可否为空 | 默认值 | 注释 |")
		tableMd = append(tableMd, "| ---- | ---- | ---- | ---- | ---- | ----  |")
		// create table.md
		cols := tables[i].ColList
		for j := range cols {
			tableMd = append(tableMd, fmt.Sprintf("| %s | %s | %s | %s | %s | %s |",
				cols[j].ColName, cols[j].ColType, cols[j].ColKey, cols[j].IsNullable, cols[j].ColDefault, cols[j].ColComment))
		}
		tableStr := strings.Join(tableMd, "\r\n")
		util.WriteToFile(path.Join(docPath, tables[i].TableName+".md"), tableStr)
	}
	// create readme.md
	readmeStr := strings.Join(readme, "\r\n")
	util.WriteToFile(path.Join(docPath, "README.md"), readmeStr)
	// create _sidebar.md
	sidebarStr := strings.Join(sidebar, "\r\n")
	util.WriteToFile(path.Join(docPath, "_sidebar.md"), sidebarStr)
	// create index.html
	util.WriteToFile(path.Join(docPath, "index.html"), docsifyHTML)
	// create .nojekyll
	util.WriteToFile(path.Join(docPath, ".nojekyll"), "")
	fmt.Println("doc generate successfully!")
	// run server
	runServer(docPath)
}

// 获取ip
func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

// runServer run http static server
func runServer(dir string) {
	var docsPort = "3000"
	ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	http.Handle("/", http.FileServer(http.Dir(dir)))
	s1 := "[i] doc server is started\n"
	s2 := "[i] use your device to visit the following URL list, gets the IP of the URL you can access:\n"
	s3 := "\t" + "http://127.0.0.1:" + docsPort + "\n"
	s4 := "\t" + "http://" + ip.String() + ":" + docsPort + "\n"
	var build strings.Builder
	build.WriteString(s1)
	build.WriteString(s2)
	build.WriteString(s3)
	build.WriteString(s4)
	s5 := build.String()
	fmt.Println(s5)
	log.Fatal(http.ListenAndServe(":"+docsPort, nil))
}
