# 消息通知示例

白虎面板提供了一个名为 `baihu` 的内建包，让您可以在脚本（Python 或 Node.js）中通过一行代码实现零配置推送。

---

## 准备工作

在运行消息通知脚本之前，请确保已经完成了必要的初始化工作。

### 1. 安装内建包
在白虎面板的在线终端或通过任务执行以下命令，为您当前的所有语言环境安装 `baihu` 包：

```bash
baihu builtininstall
```

### 2. 配置环境变量
前往 **「定时任务」** -> **「编辑任务」** -> **「环境变量」**，添加以下两个键值对：

- `BHPKG_NOTIFY_TOKEN`: 进入「消息推送」->「脚本调用说明」标签即可找到。
- `BHPKG_NOTIFY_CHANNEL`: 进入「消息推送」->「渠道列表」标签即可查看对应的 **ID**。

> [!TIP]
> 如果您更改了容器内部的服务端口（默认 8052），还需要额外添加 `BHPKG_NOTIFY_URL` 变量。详情请参考 [消息中心说明](../notify.md)。

---

## 代码示例

### Python (同步)

```python
import baihu

# 内建通知测试示例 (Python)
def main():
    print("正在尝试发送 Python 内建通知...")
    try:
        # 调用内建 notify 函数
        # 内部会自动使用环境变量进行鉴权和投递
        response = baihu.notify(
            title="Python 任务提醒",
            text="这是一条来自 Python 示例脚本的通知消息。调用非常简单！"
        )
        print(f"服务器响应: {response}")
            
    except Exception as e:
        print(f"发送过程发生异常: {e}")

if __name__ == "__main__":
    main()
```

### Node.js (异步)

```javascript
const baihu = require('baihu');

/**
 * 内建通知测试示例 (Node.js)
 */
console.log("正在尝试发送 Node.js 内建通知...");

try {
    // 简单的一行代码即可完成推送
    baihu.notify(
        "Node.js 任务提醒", 
        "这是一条来自 Node.js 示例脚本的通知消息。无需配置 API 地址或 Token。"
    );
    
    console.log("发送请求已提交。");
    console.log("提示：内建包采用异步非阻塞发送，不会干扰主逻辑执行。");
    
} catch (e) {
    console.error(`通知失败: ${e.message}`);
}
```

---

## 运行与验证

1. **保存脚本**：将上述代码保存为 `.py` 或 `.js` 文件。
2. **创建任务**：在面板中创建新任务并关联该文件。
3. **注入配置**：在任务配置中填入 `BHPKG_NOTIFY_TOKEN` 等变量。
4. **立即运行**：点击「运行」按钮，检查对应的消息渠道是否收到了推送消息。
