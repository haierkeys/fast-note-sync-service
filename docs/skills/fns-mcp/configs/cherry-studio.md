# Cherry Studio MCP Configuration Guide / 配置指南

Follow these steps to use Fast Note Sync Service in Cherry Studio:
按照以下步骤在 Cherry Studio 中使用 Fast Note Sync Service：

### 1. Open MCP Settings / 进入 MCP 设置
- Open **Cherry Studio**.
- Click **Settings** (gear icon) / 点击 **设置**（齿轮图标）。
- Select **MCP Servers** / 选择 **MCP 服务器**。

### 2. Add New Server / 添加服务器
- Click **Add Server** / 点击 **添加服务器**。
- Fill out the form / 填写以下内容：
    - **Name / 名称**: `FNS-Service`
    - **Type / 类型**: `SSE (Server-Sent Events)`
    - **URL**: `http://<YOUR_DOMAIN>:9000/api/mcp/sse`
    - **Headers / 请求头**:
        - **Key**: `Authorization`, **Value**: `Bearer <YOUR_AUTH_TOKEN>`
        - (Optional) **Key**: `X-Client`, **Value**: `CherryStudio`
        - (Optional) **Key**: `X-Client-Name`, **Value**: `Cherry Agent`
        - (Optional) **Key**: `X-Client-Version`, **Value**: `1.0.0`
        - (Optional) **Key**: `X-Default-Vault-Name`, **Value**: `Default`

### 3. Save and Enable / 保存并启用
- Click **Confirm** / 点击 **确定**。
- Ensure it is **Enabled** / 确保处于 **启用** 状态。

### 4. Use in Chat / 在聊天中使用
- Start a new chat with a model supporting Tool Calling.
- Confirm tools from `FNS-Service` are selected.
- Now you can say: "List my notes" or "在库中搜索 X"。
