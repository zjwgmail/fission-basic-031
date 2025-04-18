/**
 * 导入全局样式文件
 * 按照规范要求的顺序导入样式文件
 */
import './styles/css-reset.less';
import './styles/css-modules.less';
import './styles/variables/variables.less';
import './styles/themes/main-theme.less';
import './styles/css-global.less';
import styles from './bootstrap.less';

import { useEffect, useRef } from 'react';
import { createRoot } from "react-dom/client";
import { useReactive } from 'ahooks';
// import { Toast } from 'antd-mobile';

// 导入工具函数
import { adaptionWebViewPort, postElement } from './utils/m';
import { onEventWinResize, postEvent } from './utils/events';
import { copyText, queryParams, safeJSONparse, safeStringify } from './utils';
import { handlerInternationalizationTransform, i18n, i18n_mode1_5, LANGUAGE_MODE, queryActivityRules, queryInternationLang, querySwitchLangthModal, queryWhatsppMessageLang, switchLangthModal } from './stores/i18n';
import { chunkDecrypt, chunkEncrypt, decrypt } from './utils/encry/rsa';
import { decomposeCode, fetchSDKTrackingPoint } from './bizs';
import { fetchGet, fetchPost } from './utils/http/onFetchRequest';
import { openLinkStore, openPointLink, openWebWhatsApp } from './bizs/mlbb';
import { rafSetTimeout } from './utils/performance';

import LanguageModal from './components/LanguageModal';
import SelectLange from './components/select-lange';
import SocialChannelsComponent from './components/SocialChannels';

export const CONSTANT_OPTIONS = {
  "projectId": 2810196,
};

/**
 * MCGG 主应用组件
 * @param {Object} props - 组件属性
 */
function AppComponent(props = {}) {
  const { qp = {} } = props;
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
  const viewData = useReactive({
    queryParams: {},
    l: {},
    ln: "02",
    lIdx: 0
  });

  useEffect(() => {
    // demo start
    // languageModalRef.current?.open();
    // let code = chunkDecrypt(`IUsQy5Uf4nPK3wZQyvqh4dLNI07iWdSmZTvwdT8ZWxSVBV18/BKzJA2alEAZq4Um+oRddY1R9DdkzTi4sn8hVYID/fvCWRFwCkpcnxkmkorfiW6dsrWhuZcjnxiesIiVniCVb4Wuhp7QQf0UXWGjHuSzo7FDMXrlGhNd/FldbIs=`);
    // console.log(`code`, code);
    // demo end

    onLoad();

    return () => { }
  }, []);

  // 初始化函数
  async function onLoad() {
    // 处理国际化数据 - 1、3、6、9、10、12
    await handlerInternationalizationTransform(i18n.data);
    // 处理奖励列表
    // await handlerInternationalizationTransform(i18n_mode1_5.data);
    // 切换语言弹窗
    await handlerInternationalizationTransform(switchLangthModal.data);

    viewData.queryParams = qp;
    // if (viewData.queryParams?.cdk) {
    //   viewData.queryParams.cdk = decrypt(viewData.queryParams.cdk); // 解密
    // }

    viewData.l = queryInternationLang(viewData.queryParams.lang || "02", viewData.queryParams.mode); // 获得语言配置
    viewData.ln = viewData.queryParams.lang || "02"; // 语言类型
    viewData.arl = queryActivityRules(viewData.queryParams.lang || "02");
    viewData.slm = querySwitchLangthModal(viewData.queryParams.lang || "02");

    // 埋点上报
    onCareOfIt(`reward${viewData.queryParams?.mode}PageExposure`);
  }
  // 切换语言
  function onSwitchLang(item, index) {
    let originLanguage = viewData.ln;
    viewData.l = queryInternationLang(item.value, viewData.queryParams.mode);
    viewData.ln = item.value;
    viewData.queryParams.lang = item.value;
    viewData.arl = queryActivityRules(item.value || "02");
    viewData.slm = querySwitchLangthModal(item.value || "02");
    viewData.lIdx = index;

    let decomposeCodeOptions = decomposeCode(viewData.queryParams?.code); // 开团集结码
    let lang = viewData.ln; // 语言
    let queryCode = viewData.queryParams?.code;
    // 相同就不要执行了
    if (originLanguage === lang) {
      return;
    }
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25031",
      "behavior": "switchLanguageClick",
      "lang": lang || "02",  //语言 01中文 02英语 03马来语
      "channel": decomposeCodeOptions?.channel || "", //渠道  a 端内-128通路 b 端内-邮件推送 c 端内-任务达人 d 端外-fb e 端外-ins f 端外-ua加热 g 端外-备用1 h 端外-备用2
      "url": window.location.href,
      "code": queryCode,
      "originLanguage": originLanguage,
      "targetLanguage": lang
    });

    languageModalRef.current?.open();
  }
  // 复制游戏码
  function onHandlerGameCode() {
    onCareOfIt(`reward${viewData.queryParams?.mode}ButtonClick`);
    copyText(viewData.queryParams?.cdk ?? "", () => {

    });
  }
  // 3人奖励、5人奖励和8人奖励地址,控制参数定义
  async function onCareOfIt(behavior = "") {
    if (!viewData.queryParams?.mode) { // rewardpage
      return;
    }
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25031",
      "behavior": behavior,
      "lang": viewData.queryParams?.lang || "02",  //语言 01中文 02英语 03马来语
      "channel": viewData.queryParams?.channel || "", //渠道
      "url": globalThis.location.href,
      "code": viewData.queryParams.code
    });
  }

  // 切换语言弹窗 - 确定
  function onModalYes() {
    // let decomposeCodeOptions = decomposeCode(viewData.queryParams?.code); // 开团集结码
    let lang = viewData.ln; // 语言
    // let queryCode = viewData.queryParams?.code;
    // fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    //   "proj": "mlbb",
    //   "act_type": "mlbb25031",
    //   "behavior": "switchLanguageClick",
    //   "lang": lang || "02",  //语言 01中文 02英语 03马来语
    //   "channel": decomposeCodeOptions?.channel || "", //渠道  a 端内-128通路 b 端内-邮件推送 c 端内-任务达人 d 端外-fb e 端外-ins f 端外-ua加热 g 端外-备用1 h 端外-备用2
    //   "url": window.location.href,
    //   "code": queryCode
    // });
    // console.log(`decomposeCodeOptions`, decomposeCodeOptions, `${decomposeCodeOptions.channel}${lang}${decomposeCodeOptions.algebra}${decomposeCodeOptions.playerCode}`);
    let msgLangConfig = queryWhatsppMessageLang(lang || "02");
    rafSetTimeout(() => {
      openWebWhatsApp(msgLangConfig?.message({ code: `a${lang}0100000` }));
    }, 300);
  }

  // 语言选择页面（社媒渠道）
  if (viewData.queryParams.socialchannel == 1) {
    return <SocialChannelsComponent />
  }

  // 界面呈现 - 邀请活动规则
  if (viewData.queryParams.gpt == 11) {
    return (
      <section className={`pr f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper} ${styles.InvitationActivityRulesWrapper}`}>
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

        <LanguageModal ref={languageModalRef} slm={viewData.slm} lang={viewData.ln} onYes={onModalYes} />
      </section>
    )
  }

  if (viewData.queryParams.lp) {
    return (
      <section className={`pr f16 oh ${styles[`langStyle${viewData.ln}`]} ${styles.AppContainerWrapper}`}>
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
        <div className={`bannerBox bannerBox${viewData.queryParams?.mode}`}></div>
        <div onClick={onHandlerGameCode} className={`pr packageCodeBox packageCodeBox${viewData.queryParams?.mode} packageCodeBoxMode${viewData.queryParams?.mode}`}>
          <span className="pa df gameCode">{viewData.queryParams?.cdk}</span>
          <a className="pa urlLink" href="https://r8qs.adj.st/appinvites/UI_CDKey?adj_t=1i7vuwle_1iz74412&adjust_deeplink=mobilelegends%3A%2F%2Fappinvites%2FUI_CDKey" rel="noopener noreferrer" target="_blank"></a>
        </div>
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

        <LanguageModal ref={languageModalRef} slm={viewData.slm} lang={viewData.ln} qp={viewData.queryParams} onYes={onModalYes} />
      </section>
    );
  }

  return "";
}
/**
 * 非落地页场景的初始化逻辑
 */
async function onMounted(qp = {}) {
  // let queryParams = queryParams();
  let gamePageType = qp.gpt;    // 游戏页面类型
  let queryCode = qp?.code;     // 开团集结码 (rallyCode)
  let gameChannel = decomposeCode(queryCode)?.channel; // 投放渠道 // qp.channel;
  let gameLang = qp?.lang; // 语言

  // 从 前往预约新游 进入
  if (!!queryCode && gamePageType == 1) {
    alert(`从 前往预约新游 按钮进来的`);
    alert(`先执行上报逻辑，然后执行：跳转功能？跳转到任务链接地址（需要先进入到游戏内？一个页面地址？再返回时看到完成开团信息？）`); // 疑问TODO：链接到哪？
    // let decry = decrypt(decodeURIComponent(queryCode));
    // if (typeof decry === "string") {
    //   decry = safeJSONparse(decry) || {};
    // }
    // alert(`decry` + "," + decry + "," + decodeURIComponent(queryCode) + "," + decomposeCode(decry.rally_code));
    // alert(`集结码拆解：${ safeStringify(decomposeCode(decry.rally_code))}`);
    // alert(`code解密：${ safeStringify(decry) }`);
    alert(`code解码后：${decodeURIComponent(queryCode)}`);
    alert(`发送助力接口，入参 - param -> 解码后：${queryCode}`);

    // fetchPost('/events/mlbb25031/gateway/activity/help', {
    //   param: queryCode
    // }, {
    //   notTipBizCodeMsg: true
    // }).then(resp => {
    //   if (resp.code !== 200) {
    //     alert(`resp.code !== 200：${resp?.message}`);
    //     return;
    //   }
    //   alert(`模拟 前往预约新游 的操作已完成，准备跳转至whatsapp`);
    //   alert('/events/mlbb25031/gateway/activity/help--POST', resp);

    //   fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
    //     // "projectId": 2810196,
    //     "proj": "mlbb",
    //     "act_type": "mlbb25031",
    //     "behavior": "appointmentLinkExposure",
    //     "lang": qp?.lang || "02",  //语言 01中文 02英语 03马来语
    //     "channel": gameChannel || "", //渠道
    //     "url": globalThis.location.href,
    //     "code": queryCode
    //   });

    //   openPointLink('https://8ufa.adj.st/?adj_t=1kyuom1r_1k7aid97&adj_redirect_ios=https%3A%2F%2Fapps.apple.com%2Fus%2Fapp%2Fmagic-chess-go-go%2Fid6612014908%3Fppid%3D88f9f6ab-4be0-46c7-8ad0-f9bf2e82b312');
    //   // openLinkStore('com.mobile.legends', 'id1160056295');
    // }).catch(err => {
    //   alert('err', err?.message);
    // });
    // boss.whatsapp.openWebWhatsApp(""); // boss.utils.rsa.decryptData(queryCode)
    return;
  }
  // 从 前往游戏内兑换 进入规则页？
  if ([10, 11, '10', '11'].includes(gamePageType)) {
    alert(`从 前往游戏内兑换 按钮进来的`);
    alert(`疑问TODO：1.需要执行上报逻辑？2.需要从web网页打开游戏？`); // 疑问TODO：1.需要执行上报逻辑？2.需要从web网页打开游戏？
    let preUrl = `${globalThis.location.origin}${globalThis.location.pathname}`;
    alert(`通过 领取方式：${preUrl} ?code=${queryCode} & gpt=${gamePageType} & lp=1 & cdk=?? CDK1001 进入最终奖励领取页面。`);
    globalThis.location.href = `${preUrl}?code=${queryCode}&gpt=${gamePageType}&lp=1`;
    // alert(`模拟 进入游戏`);
    // boss.whatsapp.openPlatformApp('', 'com.mobile.legends', 'id1160056295');
    return;
  }
  // 开团人的条件，没有用户的gpt，但是存在code集结码，作为开团人状态。
  if (!!queryCode && !gamePageType) {
    console.log(gameLang, gameChannel);
    alert(`进入渠道对应的投放页面。${JSON.stringify(qp)}，投放渠道：${gameChannel}。`);
    alert(`准备跳转whatsapp并且执行：我要参与MLBB组队活动，抽vivo手机、RM5000现金和限定皮肤等奖励！\n我的活动码：${queryCode}`);
    // 上报数据
    // let decomposeCodeOptions = decomposeCode(qp?.code); // 开团集结码
    let lang = qp?.lang; // 语言
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25031",
      "behavior": "promotionLinkExposure",
      "lang": lang || "02",  //语言 01中文 02英语 03马来语
      "channel": gameChannel, // decomposeCodeOptions?.channel || "", //渠道  a 端内-128通路 b 端内-邮件推送 c 端内-任务达人 d 端外-fb e 端外-ins f 端外-ua加热 g 端外-备用1 h 端外-备用2
      "url": window.location.href,
      "code": queryCode
    });

    let msgLangConfig = queryWhatsppMessageLang(gameLang || "02");
    alert(`对应发送whatsapp消息：${msgLangConfig?.message({ code: queryCode })}`);
    rafSetTimeout(() => {
      openWebWhatsApp(msgLangConfig?.message({ code: queryCode }));
    }, 300);
    return;
  }
}

/**
 * 生成并渲染React应用
 * 用于落地页场景
 */
function generateElement(qp = {}) {
  const container = createRoot(document.querySelector("#root"));
  container.render(<AppComponent qp={qp} />);
}
/**
 * 应用主入口函数
 * 负责初始化全局配置和响应式适配
 */
async function Main() {
  // 设置全局产品名称
  // globalThis["$ProductionName"] = `mcgg_202501252300`;

  /**
   * 视口变化处理函数
   * 根据屏幕宽度计算html的font-size，实现响应式布局
   * @returns {number} 计算后的font-size值
   */
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

  // 监听窗口大小和方向变化，动态调整布局
  postEvent('windowResizeAndOrientationChange', () => {
    return onViewportChange();
  });
  onEventWinResize('windowResizeAndOrientationChange');

  // 根据URL参数判断是否为落地页
  let _getQueryParams = queryParams() ?? {};
  let landPage = _getQueryParams?.lp;      // 是否是落地页
  let queryCode = _getQueryParams?.code;     // 唯一code码
  let mode = _getQueryParams?.mode;
  let gpt = _getQueryParams?.gpt;
  // let socialChannel = _getQueryParams?.socialchannel;

  // try {
  // } catch (err) {
  //   console.error(`err`, err);
  // }
  if (landPage == 1) {
    // 获取参数
    if (gpt == 1 && !!queryCode) {
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

    generateElement(_getQueryParams);  // 落地页场景
  } else {
    onMounted(_getQueryParams);       // 非落地页场景
  }

}
// 启动应用
Main();

// 辅助函数
function alert() { }


// console.log(
//   chunkEncrypt(safeStringify({
//     code: queryCode,
//     mode,
//   })),
//   chunkDecrypt(
//     chunkEncrypt(safeStringify({
//       code: queryCode,
//       mode,
//     }))
//   )
// )
// let QUERY_FETCH_CDK_DATA = await fetchPost("/events//mlbb25031gateway/activity/cdk", {
//   param: chunkEncrypt(safeStringify({
//     code: queryCode,
//     mode: mode - 0,
//   }))
// });
// let data = QUERY_FETCH_CDK_DATA?.data;