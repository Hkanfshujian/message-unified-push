

// 定义一些常量名
const CONSTANT = {
    PAGE: 1,
    // PAGE_SIZE: 8,
    TOTAL: 0,
    DEFALUT_SITE_CONFIG: JSON.stringify({ "logo": "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 128 128' width='200' height='200'><defs><linearGradient id='g' x1='0' y1='0' x2='1' y2='1'><stop offset='0%' stop-color='#0EA5E9'/><stop offset='100%' stop-color='#2563EB'/></linearGradient></defs><rect x='8' y='8' width='112' height='112' rx='30' fill='url(#g)'/><path d='M32 94V34h16l16 24 16-24h16v60H80V60L64 84 48 60v34H32z' fill='#FFFFFF'/></svg>", "pagesize": "8", "slogan": "Message Unified Push Platform", "title": "消息统一推送中台", "login_title": "消 息 统 一 推 送 中 台", "theme_color": "blue" }),
    LOG_TASK_ID: "T-IM1GBswSRY",
    STORE_TOKEN_NAME: '__message_nest_token__',
    STORE_AUTH_SOURCE_NAME: '__message_nest_auth_source__',
    STORE_RBAC_AUTHZ_NAME: '__message_nest_rbac_authz__',
    STORE_CUSTOM_NAME: '__message_nest_custom_site__',
    STORE_OPEN_TABS_NAME: '__message_nest_open_tabs_v1',
    NO_AUTH_URL: [
        '/auth',
        '/auth/register',
        '/auth/public-config',
    ],
    WAYS_DATA: [
        {
            type: 'Email',
            label: '邮箱',
            // 动态接收者配置
            dynamicRecipient: {
                support: true,              // 是否支持动态接收者
                field: 'to_account',        // 接收者字段名
                label: '收件邮箱',           // 接收者字段标签
                desc: '邮箱地址',            // 接收者字段描述
            },
            inputs: [
                { subLabel: 'smtp服务地址', value: '', col: 'server', desc: "smtp@xyz.com" },
                { subLabel: 'smtp服务端口', value: '', col: 'port', desc: "port" },
                { subLabel: '邮箱账号', value: '', col: 'account', desc: "邮箱账号" },
                { subLabel: '邮箱密码', value: '', col: 'passwd', desc: "邮箱密码" },
                { subLabel: '发信人名称', value: '', col: 'from_name', desc: "想要设置的发信人名字" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'html', content: 'html' },
            ],
            taskInsInputs: [
                { value: '', col: 'to_account', desc: '收件邮箱', label: '收件邮箱' },
            ],
        },
        {
            type: 'Dtalk',
            label: '钉钉',
            inputs: [
                { subLabel: 'access_token', value: '', col: 'access_token', desc: "钉钉webhook中的access_token" },
                { subLabel: '加签', value: '', col: 'secret', desc: "加签的签名，SEC开头" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "输入框说明", desc: "钉钉支持加签和关键字过滤，如果是配置了关键字过滤，只需要消息里面包含了关键字，就会发送"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'QyWeiXin',
            label: '企业微信机器人',
            inputs: [
                { subLabel: 'token', value: '', col: 'access_token', desc: "企业微信webhook中的token" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Feishu',
            label: '飞书机器人',
            inputs: [
                { subLabel: 'access_token', value: '', col: 'access_token', desc: "飞书webhook中的access_token" },
                { subLabel: '加签', value: '', col: 'secret', desc: "加签的签名" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "输入框说明", desc: "飞书支持加签和关键字过滤，如果是配置了关键字过滤，只需要消息里面包含了关键字，就会发送"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Custom',
            label: '自定义推送',
            inputs: [
                { subLabel: 'webhook地址', value: '', col: 'webhook', desc: "自定义webhook地址" },
                { subLabel: '请求体', value: '', col: 'body', desc: "text内容请使用 TEXT 进行占位\n例如：{\"message\": \"TEXT\", \"foo\": \"bar\"}", isTextArea: true },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "自定义webhook说明", desc: "自定义webhook暂时只支持text，消息将解析TEXT占位标识进行替换，暂时只支持POST方式"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'WeChatOFAccount',
            label: '微信测试公众号模板',
            // 动态接收者配置
            dynamicRecipient: {
                support: true,              // 是否支持动态接收者
                field: 'to_account',        // 接收者字段名
                label: '接收者OpenId',       // 接收者字段标签
                desc: 'OpenId',             // 接收者字段描述
            },
            inputs: [
                { subLabel: 'appID', value: '', col: 'appID', desc: "公众号appid" },
                { subLabel: 'appsecret', value: '', col: 'appsecret', desc: "公众号appsecret" },
                { subLabel: '模板id', value: '', col: 'tempid', desc: "模板消息id" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "公众号消息说明", desc: "微信测试公众号模板消息发送，token使用内存缓存，<br />秘钥请访问 https://mp.weixin.qq.com/debug/cgi-bin/sandboxinfo?action=showinfo&t=sandbox/index"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
                { value: '', col: 'to_account', desc: '接收者OpenId', label: '接收者OpenId' },
            ],
        },
        // 暂时屏蔽阿里云短信入口
        {
            type: 'AliyunSMS',
            label: '阿里云短信',
            // 动态接收者配置
            dynamicRecipient: {
                support: true,              // 是否支持动态接收者
                field: 'phone_number',      // 接收者字段名
                label: '手机号码',           // 接收者字段标签
                desc: '手机号码',            // 接收者字段描述
            },
            inputs: [
                { subLabel: 'AccessKeyId', value: '', col: 'access_key_id', desc: "阿里云AccessKeyId" },
                { subLabel: 'AccessKeySecret', value: '', col: 'access_key_secret', desc: "阿里云AccessKeySecret" },
                { subLabel: 'RegionId', value: 'cn-hangzhou', col: 'region_id', desc: "阿里云区域ID，如cn-hangzhou" },
                { subLabel: '短信签名', value: '', col: 'sign_name', desc: "短信签名名称" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "阿里云短信说明", desc: "使用阿里云短信服务发送短信，需要在阿里云控制台申请短信签名和模板。<br />AccessKey请在阿里云控制台获取。"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
                { value: '', col: 'phone_number', desc: "手机号码（接收短信的手机号）", label: '手机号码' },
                { value: '', col: 'template_code', desc: "短信模板CODE（在阿里云短信控制台获取）", label: '短信模板CODE' },
            ],
        },
        {
            type: 'Telegram',
            label: 'Telegram机器人',
            inputs: [
                { subLabel: 'Bot Token', value: '', col: 'bot_token', desc: "Telegram Bot Token" },
                { subLabel: 'Chat ID', value: '', col: 'chat_id', desc: "接收消息的Chat ID或User ID" },
                { subLabel: '自定义API地址', value: '', col: 'api_host', desc: "可选，自定义Telegram API地址（优先级最高）" },
                { subLabel: '代理地址', value: '', col: 'proxy_url', desc: "可选，支持 HTTP/HTTPS/SOCKS5 代理" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "Telegram机器人说明",
                desc: "使用Telegram Bot发送消息。<br />1. 通过 @BotFather 创建机器人获取Bot Token<br />2. Chat ID可以是用户ID、群组ID或频道ID<br />3. <strong>代理配置说明：</strong><br />&nbsp;&nbsp;• <strong>自定义API地址</strong>（优先级最高）：适用于自建代理服务器，如 https://api.example.com<br />&nbsp;&nbsp;• <strong>代理地址</strong>（优先级较低）：支持以下格式<br />&nbsp;&nbsp;&nbsp;&nbsp;- HTTP代理：http://127.0.0.1:7890<br />&nbsp;&nbsp;&nbsp;&nbsp;- HTTPS代理：https://proxy.example.com:8080<br />&nbsp;&nbsp;&nbsp;&nbsp;- SOCKS5代理：socks5://127.0.0.1:1080<br />&nbsp;&nbsp;&nbsp;&nbsp;- 带认证的SOCKS5：socks5://user:pass@host:1080<br />&nbsp;&nbsp;• 如果同时配置，将优先使用自定义API地址，代理地址会被忽略"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
                { subLabel: 'html', content: 'html' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Bark',
            label: 'Bark推送',
            inputs: [
                { subLabel: 'Push Key', value: '', col: 'push_key', desc: "Bark设备码或自建IP，例：DxHcxxxxxRxxxxxxcm" },
                { subLabel: '存档', value: '', col: 'archive', desc: "1（存档）或 0（不存档），可选" },
                { subLabel: '分组', value: '', col: 'group', desc: "推送分组，可选" },
                { subLabel: '推送声音', value: '', col: 'sound', desc: "推送铃声，可选" },
                { subLabel: '推送图标', value: '', col: 'icon', desc: "推送图标URL，可选" },
                { subLabel: '推送时效', value: '', col: 'level', desc: "active/timeSensitive/passive，可选" },
                { subLabel: '跳转URL', value: '', col: 'url', desc: "点击推送跳转的URL，可选" },
                { subLabel: '加密Key', value: '', col: 'key', desc: "AES加密Key，16位，可选" },
                { subLabel: '加密IV', value: '', col: 'iv', desc: "AES加密IV，16位，可选" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "Bark推送说明",
                desc: "Bark 是一款 iOS 专用的推送软件。<br />1. 在 iPhone 上安装 Bark App<br />2. 复制你的设备 Key 或自建服务器地址<br />3. 详细参数说明可以查看 Bark 官方文档"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'PushMe',
            label: 'PushMe推送',
            inputs: [
                { subLabel: 'Push Key', value: '', col: 'push_key', desc: "PushMe的push_key" },
                { subLabel: '自定义API地址', value: '', col: 'url', desc: "默认: https://push.i-i.me/，可选" },
                { subLabel: '日期', value: '', col: 'date', desc: "日期，可选" },
                { subLabel: '类型', value: '', col: 'type', desc: "类型，可选" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "PushMe推送说明",
                desc: "PushMe 是一款推送服务。<br />1. 获取你的 Push Key<br />2. 如果有自建服务，填写自定义 API 地址"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Ntfy',
            label: 'Ntfy推送',
            inputs: [
                { subLabel: 'Topic', value: '', col: 'topic', desc: "Ntfy的Topic名字，必填" },
                { subLabel: '自定义API地址', value: '', col: 'url', desc: "默认: https://ntfy.sh/，可选" },
                { subLabel: '优先级', value: '3', col: 'priority', desc: "1(min) 到 5(max)，可选" },
                { subLabel: '图标URL', value: '', col: 'icon', desc: "推送显示的图标URL，可选" },
                { subLabel: 'Token', value: '', col: 'token', desc: "认证Token，可选" },
                { subLabel: '用户名', value: '', col: 'username', desc: "Basic认证用户名，可选" },
                { subLabel: '密码', value: '', col: 'password', desc: "Basic认证密码，可选" },
                { subLabel: 'Actions', value: '', col: 'actions', desc: "JSON格式的Action配置，可选" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "Ntfy推送说明",
                desc: "Ntfy 是一款基于 HTTP 的发布-订阅通知服务。<br />1. 填写你的 Topic<br />2. 如果是自建服务，填写自定义 API 地址<br />3. 支持 Token 或 用户名密码 认证"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'Gotify',
            label: 'Gotify推送',
            inputs: [
                { subLabel: 'Gotify服务地址', value: '', col: 'url', desc: "Gotify服务的URL，例如: https://gotify.example.com" },
                { subLabel: 'Token', value: '', col: 'token', desc: "APP的Token" },
                { subLabel: '优先级', value: '5', col: 'priority', desc: "消息优先级，可选" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "Gotify推送说明",
                desc: "Gotify 是一款自托管的消息推送服务。<br />1. 在 Gotify 中创建 APP 获取 Token<br />2. 填写 Gotify 服务的完整 URL"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
            ],
            taskInsInputs: [
            ],
        },
        {
            type: 'QyWeiXinApp',
            label: '企业微信应用',
            inputs: [
                { subLabel: '企业ID', value: '', col: 'corp_id', desc: "企业微信的CorpID" },
                { subLabel: '应用Secret', value: '', col: 'corp_secret', desc: "应用的Secret" },
                { subLabel: '应用ID', value: '', col: 'agent_id', desc: "应用的AgentID" },
                { subLabel: '渠道名', value: '', col: 'name', desc: "想要设置的渠道名字" },
            ],
            tips: {
                text: "企业微信应用说明", desc: "通过企业微信自建应用发送消息。支持发送文本和Markdown。"
            },
            taskInsRadios: [
                { subLabel: 'text', content: 'text' },
                { subLabel: 'markdown', content: 'markdown' },
            ],
            taskInsInputs: [
            ],
        },
    ],
}


// 转换渠道map
CONSTANT.WAYS_DATA_MAP = {};
CONSTANT.WAYS_DATA.forEach(element => {
    CONSTANT.WAYS_DATA_MAP[element.type] = element
});

export { CONSTANT }
