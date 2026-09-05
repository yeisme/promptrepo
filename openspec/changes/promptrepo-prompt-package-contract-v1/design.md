## 设计

通过独立 Go 类型增加方案、步骤绑定和提示包 manifest。方案只描述静态依赖，不含脚本或可执行插件。输入绑定仅允许 field、source 和 step_output。

拓扑排序稳定；缺失依赖与循环直接拒绝。manifest 的 digest 不包括自身 digest 字段，使用 JCS 与 SHA-256；文件路径在 Windows/macOS/Linux 上均须可搬运，禁止逃逸、保留设备名和大小写碰撞。

Registry 保存用户值、模板快照和会话，SDK 只在内存中校验显式传入的内容。不增加存储、服务端或 Provider 依赖。

兼容与回滚：所有类型和函数为新增；旧消费者无需升级，Registry 可回退到旧版本但不得读取新包并声称已验证。正文、模板与私有输入不进入普通日志或证据。
