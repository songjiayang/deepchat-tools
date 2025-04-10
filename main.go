package main

import (
	"flag"
	"log"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	btools "github.com/songjiayang/deepchat-tools/bailian/tools"
	"github.com/songjiayang/deepchat-tools/pkg"
	vtools "github.com/songjiayang/deepchat-tools/volc/tools"
)

var (
	mode         string
	serverlisten string
)

var (
	serverName    = "DeepChat Tools"
	serverVersion = "v0.1.0"
)

func init() {
	flag.StringVar(&mode, "transport", "sse", "The transport to use, should be \"stdio\" or \"sse\"")
	flag.StringVar(&serverlisten, "server_listen", "127.0.0.1:8082", "The sse server listen address")
	flag.Parse()
}

func getTransport() (t transport.ServerTransport) {
	if mode == "stdio" {
		log.Println("start mcp server with stdio transport")
		t = transport.NewStdioServerTransport()
	} else {
		log.Printf("start mcp server with sse transport, listen %s", serverlisten)
		t, _ = transport.NewSSEServerTransport(serverlisten)
	}

	return t
}

func main() {
	svr, _ := server.NewServer(
		getTransport(),
		server.WithServerInfo(protocol.Implementation{
			Name:    serverName,
			Version: serverVersion,
		}),
	)

	// register poster tool
	svr.RegisterTool(btools.NewPosterTool(), btools.NewPosterToolHandler())
	svr.RegisterTool(vtools.NewImageStyleTool(), vtools.NewImageStyleHandler())
	pkg.RunWithSignalWaiter(svr)
}
