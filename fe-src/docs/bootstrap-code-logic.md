# MLBB25031 项目 Bootstrap 代码逻辑文档

## 目录

1. [概述](#概述)
2. [引入依赖](#引入依赖)
3. [常量定义](#常量定义)
4. [主应用组件](#主应用组件)
   - [状态管理](#状态管理)
   - [生命周期钩子](#生命周期钩子)
   - [初始化函数](#初始化函数)
   - [切换语言功能](#切换语言功能)
   - [其他功能函数](#其他功能函数)
   - [界面渲染](#界面渲染)
5. [非落地页场景初始化](#非落地页场景初始化)
6. [应用主入口函数](#应用主入口函数)
   - [视口响应式处理](#视口响应式处理)
   - [落地页和非落地页逻辑](#落地页和非落地页逻辑)
7. [流程图](#流程图)
8. [关键功能点](#关键功能点)

## 概述

`bootstrap.jsx` 是 MLBB25031 项目的启动文件，负责初始化应用、处理国际化、响应式布局设置以及根据URL参数决定应用的渲染模式。该文件实现了落地页和非落地页两种模式的逻辑处理，并集成了语言切换、WhatsApp 分享、埋点统计等功能。

## 引入依赖

代码首先导入了所需的样式文件和依赖库：

```jsx
// 样式文件按特定顺序导入
import './styles/css-reset.less';
import './styles/css-modules.less';
import './styles/variables/variables.less';
import './styles/themes/main-theme.less';
import './styles/css-global.less';
import styles from './bootstrap.less';

// React 相关库
import { useEffect, useRef } from 'react';
import { createRoot } from "react-dom/client";
import { useReactive } from 'ahooks';

// 工具函数和服务
import { adaptionWebViewPort, postElement } from './utils/m';
import { onEventWinResize, postEvent } from './utils/events';
import { copyText, queryParams, safeJSONparse, safeStringify } from './utils';
import { handlerInternationalizationTransform, i18n, i18n_mode1_5, LANGUAGE_MODE, queryActivityRules, 
         queryInternationLang, querySwitchLangthModal, queryWhatsppMessageLang, 
         switchLangthModal } from './stores/i18n';
import { chunkDecrypt, chunkEncrypt, decrypt } from './utils/encry/rsa';
import { decomposeCode, fetchSDKTrackingPoint } from './bizs';
import { fetchGet, fetchPost } from './utils/http/onFetchRequest';
import { openLinkStore, openPointLink, openWebWhatsApp } from './bizs/mlbb';
import { rafSetTimeout } from './utils/performance';

// 组件
import LanguageModal from './components/LanguageModal';
import SelectLange from './components/select-lange';
import SocialChannelsComponent from './components/SocialChannels';
```

## 常量定义

项目定义了一个常量对象用于配置：

```jsx
export const CONSTANT_OPTIONS = {
  "projectId": 2810196,
};
```

## 主应用组件

`AppComponent` 是主应用组件，接收 URL 查询参数作为属性。

### 状态管理

```jsx
function AppComponent(props = {}) {
  const { qp = {} } = props;
  // 定义支持的语言列表
  const MAP_LANG = useRef([{
    "label": "EN",
    "value": "02"
  }, {
    "label": "MY",
    "value": "03"
  }, {
    "label": "ID",
    "value": "04"
  }]);
  const languageModalRef = useRef(null);
  // 使用 ahooks 的 useReactive 管理响应式状态
  const viewData = useReactive({
    queryParams: {},
    l: {},         // 语言配置
    ln: "02",      // 当前语言代码
    lIdx: 0        // 当前语言索引
  });
```

### 生命周期钩子

组件使用 `useEffect` 在挂载时进行初始化：

```jsx
useEffect(() => {
  onLoad();
  return () => { }
}, []);
```

### 初始化函数

```jsx
async function onLoad() {
  // 处理国际化数据
  await handlerInternationalizationTransform(i18n.data);
  // 处理切换语言弹窗
  await handlerInternationalizationTransform(switchLangthModal.data);

  // 初始化视图数据
  viewData.queryParams = qp;
  viewData.l = queryInternationLang(viewData.queryParams.lang || "02", viewData.queryParams.mode);
  viewData.ln = viewData.queryParams.lang || "02";
  viewData.arl = queryActivityRules(viewData.queryParams.lang || "02");
  viewData.slm = querySwitchLangthModal(viewData.queryParams.lang || "02");

  // 埋点上报
  onCareOfIt(`reward${viewData.queryParams?.mode}PageExposure`);
}
```

### 切换语言功能

```jsx
function onSwitchLang(item, index) {
  let originLanguage = viewData.ln;
  // 更新语言相关配置
  viewData.l = queryInternationLang(item.value, viewData.queryParams.mode);
  viewData.ln = item.value;
  viewData.queryParams.lang = item.value;
  viewData.arl = queryActivityRules(item.value || "02");
  viewData.slm = querySwitchLangthModal(item.value || "02");
  viewData.lIdx = index;

  // 如果语言相同则不执行后续操作
  if (originLanguage === viewData.ln) {
    return;
  }
  
  // 埋点上报语言切换行为
  let decomposeCodeOptions = decomposeCode(viewData.queryParams?.code);
  let lang = viewData.ln;
  let queryCode = viewData.queryParams?.code;
  fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    "proj": "mlbb",
    "act_type": "mlbb25031",
    "behavior": "switchLanguageClick",
    "lang": lang || "02",
    "channel": decomposeCodeOptions?.channel || "",
    "url": window.location.href,
    "code": queryCode,
    "originLanguage": originLanguage,
    "targetLanguage": lang
  });

  // 打开语言切换确认弹窗
  languageModalRef.current?.open();
}
```

### 其他功能函数

```jsx
// 复制游戏码
function onHandlerGameCode() {
  onCareOfIt(`reward${viewData.queryParams?.mode}ButtonClick`);
  copyText(viewData.queryParams?.cdk ?? "", () => {});
}

// 埋点上报
async function onCareOfIt(behavior = "") {
  if (!viewData.queryParams?.mode) {
    return;
  }
  fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    "proj": "mlbb",
    "act_type": "mlbb25031",
    "behavior": behavior,
    "lang": viewData.queryParams?.lang || "02",
    "channel": viewData.queryParams?.channel || "",
    "url": globalThis.location.href,
    "code": viewData.queryParams.code
  });
}

// 切换语言弹窗确认
function onModalYes() {
  let lang = viewData.ln;
  let msgLangConfig = queryWhatsppMessageLang(lang || "02");
  // 延迟打开 WhatsApp 分享
  rafSetTimeout(() => {
    openWebWhatsApp(msgLangConfig?.message({ code: `a${lang}0100000` }));
  }, 300);
}
```

### 界面渲染

组件根据 URL 参数渲染不同的界面：

1. 社交媒体渠道页面：
```jsx
if (viewData.queryParams.socialchannel == 1) {
  return <SocialChannelsComponent />
}
```

2. 邀请活动规则页面：
```jsx
if (viewData.queryParams.gpt == 11) {
  return (
    <section className={`pr f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper} ${styles.InvitationActivityRulesWrapper}`}>
      {/* 头部、语言选择器 */}
      <div className="pa pen pricePopBox"></div>
      <head className="pr df headerBox">
        <div className="df logoBox">
          <div className={`oh h5Logo`} title="MOBA Logo"></div>
        </div>
        <SelectLange
          value={viewData.ln}
          options={MAP_LANG.current}
          onChange={(option, index) => onSwitchLang(option, index)}
        />
      </head>
      {/* 规则内容展示 */}
      <div className={`df ruleTitleBox ruleTitleBox${viewData.ln}`}></div>
      <div className={`pr oh m0a contentBox contentBox${viewData.queryParams?.mode}`}>
        <div className="pa pen ruleContentReaptBox"></div>
        <div className={`activeRuleContentBox`}>
          {
            viewData.arl?.activeRuleContent?.map((it, idx) => {
              return (
                <>
                  <div key={idx} className="df itText">
                    <strong className={`oh tac serial ${idx === 0 ? 'ovh' : ''}`}>{(idx)}</strong>
                    <span className="val" dangerouslySetInnerHTML={{ __html: (typeof it.text === "function" ? it.text(viewData.queryParams) : it.text) }}></span>
                  </div>
                  {
                    it.imgs?.length ? it.imgs.map((imgItem, imgIdx) => {
                      return (
                        <img key={imgIdx} className="itImg" src={imgItem.url} />
                      )
                    }) : ""
                  }
                </>
              )
            })
          }
        </div>
      </div>
      {/* 奖品表格 */}
      <div className={`sphide df ruleTitleBox prizeTableTitleImg${viewData.ln}`}></div>
      <div className={`sphide oh m0a contentBox prizeTableBox contentBox${viewData.queryParams?.mode}`}>
        <div className={`prizeTable`}>
          <table cellPadding={0} cellSpacing={0} border={0}>
            <thead>
              <tr>
                {
                  viewData.arl?.activeRuleWinningInfoContent?.columns?.map((it, idx) => {
                    return <th className="th" key={idx}>{it.thTitle}</th>
                  })
                }
              </tr>
            </thead>
            <tbody>
              {
                viewData.arl?.activeRuleWinningInfoContent?.dataSource?.map((it, idx) => {
                  return <tr key={idx}>
                    <td>{it.whatsappAccount}</td>
                    <td>{it.whatsappName}</td>
                    <td>{it.winningPrize}</td>
                  </tr>
                })
              }
            </tbody>
          </table>
        </div>
      </div>
      {/* 语言切换弹窗 */}
      <LanguageModal ref={languageModalRef} slm={viewData.slm} lang={viewData.ln} onYes={onModalYes} />
    </section>
  )
}
```

3. 落地页：
```jsx
if (viewData.queryParams.lp) {
  return (
    <section className={`pr f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper}`}>
      {/* 头部、语言选择器 */}
      <div className="pa pen pricePopBox"></div>
      <head className="pr df headerBox">
        <div className="df logoBox">
          <div className={`oh h5Logo`} title="MOBA Logo"></div>
        </div>
        <SelectLange
          value={viewData.ln}
          options={MAP_LANG.current}
          onChange={(option, index) => onSwitchLang(option, index)}
        />
      </head>
      {/* Banner */}
      <div className={`bannerBox bannerBox${viewData.queryParams?.mode}`}></div>
      {/* 游戏码复制区域 */}
      <div onClick={onHandlerGameCode} className={`pr packageCodeBox packageCodeBox${viewData.queryParams?.mode} packageCodeBoxMode${viewData.queryParams?.mode}`}>
        <span className="pa df gameCode">{viewData.queryParams?.cdk}</span>
        <a className="pa urlLink" href="https://r8qs.adj.st/appinvites/UI_CDKey?adj_t=1i7vuwle_1iz74412&adjust_deeplink=mobilelegends%3A%2F%2Fappinvites%2FUI_CDKey" rel="noopener noreferrer" target="_blank"></a>
      </div>
      {/* 内容区域 */}
      <div className={`oh m0a contentBox contentBox${viewData.queryParams?.mode}`}>
        <head className="df m0a titBox"></head>
        <div className={`activeRuleContentBox`}>
          {
            viewData.l?.activeRuleContent?.map((it, idx) => {
              return (
                <>
                  <div key={idx} className="df itText">
                    <span className="val" dangerouslySetInnerHTML={{ __html: (typeof it.text === "function" ? it.text(viewData.queryParams) : it.text) }}></span>
                  </div>
                  {
                    it.imgs?.length ? it.imgs.map((imgItem, imgIdx) => {
                      return (
                        <img key={imgIdx} className="itImg" src={imgItem.url} />
                      )
                    }) : ""
                  }
                </>
              )
            })
          }
        </div>
      </div>
      {/* 语言切换弹窗 */}
      <LanguageModal ref={languageModalRef} slm={viewData.slm} lang={viewData.ln} qp={viewData.queryParams} onYes={onModalYes} />
    </section>
  );
}
```

## 非落地页场景初始化

`onMounted` 函数处理非落地页场景的初始化逻辑：

```jsx
async function onMounted(qp = {}) {
  let gamePageType = qp.gpt;    // 游戏页面类型
  let queryCode = qp?.code;     // 开团集结码 (rallyCode)
  let gameChannel = decomposeCode(queryCode)?.channel; // 投放渠道
  let gameLang = qp?.lang;      // 语言

  // 1. 从"前往预约新游"进入
  if (!!queryCode && gamePageType == 1) {
    // 此处实际代码中包含了多个 alert 调试信息，主要逻辑是：
    // - 发送助力接口请求
    // - 执行埋点上报
    // - 跳转到相应链接或应用
    return;
  }
  
  // 2. 从"前往游戏内兑换"进入
  if ([10, 11, '10', '11'].includes(gamePageType)) {
    // 此处代码主要执行以下逻辑：
    // - 构建落地页 URL
    // - 跳转到落地页
    let preUrl = `${globalThis.location.origin}${globalThis.location.pathname}`;
    globalThis.location.href = `${preUrl}?code=${queryCode}&gpt=${gamePageType}&lp=1`;
    return;
  }
  
  // 3. 开团人状态处理
  if (!!queryCode && !gamePageType) {
    // 埋点上报
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25031",
      "behavior": "promotionLinkExposure",
      "lang": gameLang || "02",
      "channel": gameChannel,
      "url": window.location.href,
      "code": queryCode
    });

    // 获取对应语言的消息配置并发送 WhatsApp 消息
    let msgLangConfig = queryWhatsppMessageLang(gameLang || "02");
    rafSetTimeout(() => {
      openWebWhatsApp(msgLangConfig?.message({ code: queryCode }));
    }, 300);
    return;
  }
}
```

## 应用主入口函数

```jsx
async function Main() {
  // 视口响应式处理
  const onViewportChange = () => {
    let _width = globalThis.innerWidth > 1080 ? 1280 : 1080;
    let adaptHtmlSize = adaptionWebViewPort(20, _width, true);
    postElement('html', {
      'font-size': adaptHtmlSize + 'px'
    });
    return adaptHtmlSize;
  }
  
  // 初始化执行一次
  onViewportChange();
  
  // 监听窗口大小和方向变化
  postEvent('windowResizeAndOrientationChange', () => {
    return onViewportChange();
  });
  onEventWinResize('windowResizeAndOrientationChange');
  
  // 获取 URL 参数
  let _getQueryParams = queryParams() ?? {};
  let landPage = _getQueryParams?.lp;      // 是否是落地页
  let queryCode = _getQueryParams?.code;   // 唯一code码
  let mode = _getQueryParams?.mode;
  let gpt = _getQueryParams?.gpt;
  
  // 落地页和非落地页的处理逻辑
  if (landPage == 1) {
    // 落地页处理逻辑
    if (gpt == 1 && !!queryCode) {
      // 获取 CDK 数据
      let QUERY_FETCH_DATA = await fetchPost("/events/mlbb25031/gateway/activity/cdk", {
        param: chunkEncrypt(safeStringify({
          code: chunkDecrypt(queryCode),
          mode: mode - 0
        }))
      });
      let data = QUERY_FETCH_DATA?.data;
      _getQueryParams.lang = data?.language;
      _getQueryParams.channel = data?.channel;
      _getQueryParams.cdk = data?.cdk;
    }
    
    // 渲染落地页
    generateElement(_getQueryParams);
  } else {
    // 非落地页处理
    onMounted(_getQueryParams);
  }
}

// 启动应用
Main();
```

```jsx
// 渲染React应用的辅助函数
function generateElement(qp = {}) {
  const container = createRoot(document.querySelector("#root"));
  container.render(<AppComponent qp={qp} />);
}
```

## 应用启动流程：

1. 调用 `Main()` 函数
2. 初始化视口响应式处理
3. 获取 URL 参数
4. 根据是否为落地页，执行不同的处理流程：
   - 落地页：获取必要数据后渲染 `AppComponent`
   - 非落地页：执行 `onMounted` 处理特定业务场景

`AppComponent` 初始化流程：

1. 组件挂载后执行 `onLoad()` 函数
2. 处理国际化数据
3. 初始化视图数据和语言配置
4. 根据 URL 参数渲染不同界面

## 关键功能点

1. **多语言支持**：支持英语(EN)、马来语(MY)和印尼语(ID)三种语言，通过语言选择器可以切换语言。

2. **响应式布局**：通过 `adaptionWebViewPort` 函数计算基础字体大小，实现响应式布局。

3. **入口场景处理**：
   - 落地页模式(lp=1)：显示完整的活动页面内容
   - 非落地页模式：根据参数执行特定动作（如打开 WhatsApp、跳转到游戏内兑换页面等）

4. **埋点统计**：使用 `fetchSDKTrackingPoint` 函数记录用户行为，如页面曝光、按钮点击、语言切换等。

5. **WhatsApp 集成**：提供 WhatsApp 消息分享功能，支持多语言消息模板。

6. **安全处理**：使用 `chunkEncrypt/chunkDecrypt` 加密解密敏感参数。

7. **界面渲染**：根据不同场景提供三种主要界面：社交媒体渠道页面、邀请活动规则页面和落地页。