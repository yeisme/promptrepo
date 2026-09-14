# 实现决策与验收

## 范围

新增 BrowseRequest/BrowseResult、多选元数据筛选、跨语言关键词、分组排序、双重计数和共用 conformance 样例。

## 查询与统计

同维度 OR、跨维度 AND；关键词只查元数据，模板语言与展示语言分开。分组最多两级，先筛选统计排序再按方案分页。原始条目数与有效 digest 去重内容数分开；多值分组明确标记不相加。去重只折叠当前页显示，保留原始引用。

## 输出与报告

表格为默认，可选 tree/detail；JSON 保持旧字段，新增 YAML 1.1。终端使用工作区既有字符宽度及 TTY 依赖版本，窄屏逐项排列。报告由 application service 原子写入单一 HTML，快照包含全部筛选结果而不是终端当前页；没有远程脚本、字体或图片依赖。元数据经过 HTML 转义及本机路径清除，不输出正文。

## 测试工具决策

复用 Go/testing、testscript 和既有集成证据 runner。Go 无法验证浏览器 DOM 与剪贴板，因此复用工作区已安装 Playwright 1.60.0；Node 原生断言将相同 fixture 交给 Go 和 HTML 模型核对，未引入第二套前端工程。

## 兼容与回滚

没有删除或改名旧命令、字段或内容引用；多选使用新增数组，旧单值仍可传入。HTML/YAML 不改变源目录或内容哈希。回滚本地 CLI/SDK 二进制即可，报告是独立生成物。

## 验证

- temp/integration-test-runs/integration-20260914T043350.939710Z/summary.json

未发布或部署远端服务。
