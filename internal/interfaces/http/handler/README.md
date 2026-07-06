# handler

HTTP 适配层（Controller/Handler）：
- 负责解析请求参数与返回响应
- 不承载业务规则
- 调用 application service 完成用例

后续按上下文拆分：auth、article、category、tag、user、setting、file、dashboard、archive。
