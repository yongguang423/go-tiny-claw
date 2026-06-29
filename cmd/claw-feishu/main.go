// cmd/claw/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yongguang423/go-tiny-claw/internal/engine"
	"github.com/yongguang423/go-tiny-claw/internal/feishu"
	"github.com/yongguang423/go-tiny-claw/internal/provider"
	"github.com/yongguang423/go-tiny-claw/internal/tools"
)

func main() {
	// 1. 初始化引擎依赖
	workDir, _ := os.Getwd()

	// 默认使用智谱 GLM-4
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
	}
	llmProvider := provider.NewZhipuOpenAIProvider("glm-4.5-air")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 开启慢思考
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, true)

	// 2. 初始化飞书 Bot（内部会读取 FEISHU_APP_ID / FEISHU_APP_SECRET）
	bot := feishu.NewFeishuBot(eng)

	// 3. 启动 WebSocket 长连接（本地无需公网/域名，程序主动连飞书）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 优雅退出：捕获 Ctrl+C / kill 信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("🛑 收到退出信号，正在关闭...")
		cancel()
	}()

	log.Printf("🚀 go-tiny-claw 飞书 WebSocket 长连接模式启动中...")
	if err := bot.StartWebSocket(ctx); err != nil {
		log.Fatalf("WebSocket 启动失败: %v", err)
	}
}
