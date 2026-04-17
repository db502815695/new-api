# 首页 UI 重设计 · Design Spec

- 日期:2026-04-17
- 范围:`web/src/pages/Home/index.jsx` 的默认首页视图(当管理员未配置 `home_page_content` 时显示)
- 风格方向:**SaaS 产品落地页**
- 不改动:`homePageContent` 的覆盖逻辑(markdown / iframe URL 仍优先)、`NoticeModal` 公告弹窗、主题/多语言联动

---

## 1. 目标

替换现有「命令中心」两栏布局(左文右 Base URL 面板 + 底部 3 场景卡),改成模块化的商业 SaaS 落地页。要求:

- 主动线清晰:Hero → 能力 → 接入示例 → 三步上手 → FAQ → Footer
- 结构可拆分可维护:每段独立子组件,`Home/index.jsx` 只做装配与分支
- 保留开发者友好体验:Base URL 复制、路径切换、多语言代码示例
- 无新增第三方依赖(代码高亮用内置 `<pre><code>` + 基础着色)

## 2. 页面骨架

顺序自上而下:

```
HeroSection            全宽,渐变背景 + 网格纹理
FeaturesSection        4 卡核心能力
IntegrationSection     左:Base URL 复制+路径轮播  右:curl/Python/Node Tab 代码块
QuickStartSection      水平 3 步(移动端竖排)
FaqSection             Accordion(默认全收起)
FooterSection          GitHub / Docs / 版本号 / 版权
```

Hero 上方/下方的 `NoticeModal` 行为不变。

> **明确去除**的模块:供应商 Logo 墙(用户指定不要)、旧版场景卡(Scenario Cards)、Hero 右侧面板。

## 3. 各段设计细节

### 3.1 HeroSection

元素从上到下:

1. **Kicker 小字**:`AI Gateway · 统一入口`(大写间距字体)
2. **主标题**:两行
   - 第一行:常规色
   - 第二行:主色→紫色 `background-clip: text` 渐变
3. **副文案**:≤ 2 行,介绍「一站式接入 40+ 模型、保留 OpenAI 兼容调用方式」
4. **CTA 组**(≤ 3 个,桌面端同排,移动端堆叠):
   - 主:`获取密钥` → `/console`,Semi `solid primary`
   - 次:`模型广场` → `/pricing`,Semi 默认按钮
   - 第三(条件式):
     - `isDemoSiteMode && statusState.status.version` → 版本号按钮,点击跳 `https://github.com/QuantumNous/new-api`
     - 否则若 `docsLink` 非空 → `查看文档` 按钮,点击打开 `docsLink`
     - 两者皆无 → 不渲染第三按钮
5. **信任条(Hero 底部)**:3 列
   - `40+ / 主流模型`
   - `OpenAI 兼容 / SDK 无缝迁移`
   - `多客户端 / 一处配置多处调用`

视觉要点:

- 背景沿用现有 `home-command-shell` 的径向光晕 + 网格纹理,**降低饱和度**
- 最大宽度容器(`max-w-6xl` 或等价),左右留白
- 移动端:标题字号降档,CTA 堆叠

### 3.2 FeaturesSection

4 列栅格(≥ md),2 列(< md),每卡:

| 卡 | 图标 | 标题 | 描述(≤ 24 字) |
|---|---|---|---|
| 1 | `IconServer` / `IconCode` | 多模型统一接入 | OpenAI 兼容,40+ 供应商无缝切换 |
| 2 | `IconCoinMoneyStroked` | 用量与计费 | 按模型、按渠道灵活配置倍率 |
| 3 | `IconShield` | 限流与渠道分流 | Token / RPM / TPM 多维限制 |
| 4 | `IconHistogram` | 调用日志与分析 | 实时查看每条请求与耗时 |

卡样式:1px 边框 + 12px 圆角 + 默认白/深色底;hover 提升边框为主色、加浅阴影、图标色块填充;不要 3D/重阴影。

### 3.3 IntegrationSection

两栏(≥ md),堆叠(< md)。

**左栏 · 接入入口**

- 标题 `接入入口`
- Base URL 块:只读代码框 + 右上复制按钮(复用 `handleCopyBaseURL`)
- 推荐路径(轮播):继续复用 Semi `ScrollList` + `ScrollItem`,每 3s 自动切换
- 底部显示「组合后示例」`${serverAddress}${currentEndpoint}`

**右栏 · 代码示例**

- 顶部 Tab:`curl` / `Python` / `Node.js`,激活态下划线 + 主色
- 代码块:`<pre><code>` 等宽字体,带行号(CSS counter),底色 `var(--semi-color-fill-0)`
- 右上角复制按钮
- 模板中 `{baseURL}`、`{endpoint}` 运行时替换为当前值

代码模板(存入 `web/src/pages/Home/sections/codeSnippets.js`):

```js
export const CODE_SNIPPETS = {
  curl: `curl {baseURL}{endpoint} \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer $API_KEY" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'`,
  python: `from openai import OpenAI

client = OpenAI(base_url="{baseURL}", api_key="$API_KEY")
resp = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello"}],
)
print(resp.choices[0].message.content)`,
  node: `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "{baseURL}",
  apiKey: process.env.API_KEY,
});

const res = await client.chat.completions.create({
  model: "gpt-4o",
  messages: [{ role: "user", content: "Hello" }],
});
console.log(res.choices[0].message.content);`,
};
```

语法高亮:**不引入第三方库**。首版直接用 `<pre><code>` + 等宽字体 + `var(--semi-color-fill-0)` 底色,不做语法着色。若日后要加,单独迭代,不纳入本次实现范围。

### 3.4 QuickStartSection

水平连线 3 步,移动端改竖排:

| 步 | 标题 | 说明 |
|---|---|---|
| 01 | 复制 Base URL | 直接使用当前站点地址 |
| 02 | 选择兼容路径 | 按客户端选 OpenAI 兼容路径 |
| 03 | 创建令牌接入 | 在控制台生成 Key,粘到应用 |

每步卡:大号灰色 index + 标题 + 一行描述;桌面端步之间用淡色分隔线 + 小箭头。

### 3.5 FaqSection

Semi `Collapse`,默认全收起。问题清单:

1. 是否支持 Azure / AWS Bedrock?
2. 定价怎么算?
3. 能否私有化部署?
4. 从 OpenAI SDK 迁移要改什么?
5. 如何启用流式输出 / 函数调用?
6. 出现限流怎么办?

答案:每条 1-2 句,指向文档/控制台具体页面的链接走 `docsLink`。

### 3.6 FooterSection

单行(桌面),两行(移动):

```
new-api · © QuantumNous 2026        GitHub · Docs · v{version}
```

- 左:品牌名 + 版权年(年份动态取 `new Date().getFullYear()`)
- 右:GitHub 链接(固定 `https://github.com/QuantumNous/new-api`)、文档链接(`docsLink`,空则不显示)、版本号(`statusState?.status?.version`,空则不显示)
- 分隔线:`1px solid var(--semi-color-border)`

## 4. 代码结构

```
web/src/pages/Home/
  index.jsx                ← 只做 homePageContent 分支 + 装配各 Section
  home.css                 ← 新增,集中放新首页所有样式(迁移并精简旧 .home-command-*)
  sections/
    HeroSection.jsx
    FeaturesSection.jsx
    IntegrationSection.jsx
    QuickStartSection.jsx
    FaqSection.jsx
    FooterSection.jsx
    codeSnippets.js
```

- `web/src/index.css` 里原 `.home-command-*` 规则块整体迁移到 `home.css`,index.css 只保留 `@import './pages/Home/home.css';` 或由 Home 组件内 `import`
- Tailwind class 与定制 class 共用:布局/间距走 Tailwind,视觉细节走自定义 class

## 5. 国际化

- 全部新文案通过 `useTranslation()` 的 `t()` 调用
- 新 key 以中文作为主键,沿用项目既有约定(`web/src/i18n/locales/zh.json` 自动同步)
- 运行 `bun run i18n:extract && bun run i18n:sync` 补齐其它语言占位

## 6. 行为与交互保留

- `NoticeModal`:装载位置不变,仍在 `Home` 根节点
- 管理员自定义首页内容:
  - `https://` 开头 → iframe,postMessage 主题/语言(保持现状)
  - 其它 → `marked.parse` → `dangerouslySetInnerHTML`(保持现状)
  - 仅当后端返回空字符串时才渲染新版首页
- `localStorage('home_page_content')` 缓存策略保持现状
- `useIsMobile`、`useActualTheme`、`StatusContext` 用法保持现状

## 7. 验收标准

1. 默认首页(无 `home_page_content`)渲染 6 个 section,顺序与本文一致
2. 桌面/移动端均无横向滚动条(保留 `overflow-x-hidden`)
3. Base URL 复制、路径轮播、代码示例 Tab 切换、代码复制 均可用
4. 切换暗/亮主题时 Hero 渐变、卡片边框、代码块底色无错位或文字不可读
5. 管理员后台设置 `home_page_content` 后,自定义内容优先,与现状一致
6. Lint 与 build 通过:`bun run lint`、`bun run eslint`、`bun run build`
7. 保留 `nеw-аρi` 与 `QuаntumΝоuѕ` 全部品牌标识(Rule 5)

## 8. 非目标(Out of Scope)

- 不新增依赖(Prism、Framer Motion 等均不引入)
- 不改动后端 API、`/api/home_page_content` 合约不变
- 不重做 `/console`、`/pricing` 等外部页面
- 不调整导航栏 `PageLayout`
- 不做埋点/分析接入
