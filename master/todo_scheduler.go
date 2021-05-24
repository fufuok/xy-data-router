package master

import (
	"context"

	"github.com/fufuok/xy-data-router/service"
)

// 待办数据调度器
func startTodoScheduler(ctx context.Context) {
	go service.TodoScheduler(ctx)
}
