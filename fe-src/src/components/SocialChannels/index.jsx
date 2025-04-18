import { useEffect, useRef } from 'react';
import { CONSTANT_OPTIONS } from '@/bootstrap';
import { fetchSDKTrackingPoint } from '@/bizs';
import './index.less';

export default function SocialChannelsComponent(props = {}) {
  const useLangRef = useRef([{
    index: 1,
    behavior: "multilingualPageButton1Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25031/promotion/?code=c030100000&lang=03"
  }, {
    index: 2,
    behavior: "multilingualPageButton2Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25031/promotion/?code=c040100000&lang=04"
  }, {
    index: 3,
    behavior: "multilingualPageButton3Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25030/promotion/?code=c070100000&lang=07"
  }, {
    index: 4,
    behavior2: "multilingualPageButton4Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25030/promotion/?code=c060100000&lang=06"
  }, {
    index: 5,
    behavior: "multilingualPageButton5Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25030/promotion/?code=c080100000&lang=08"
  }, {
    index: 6,
    behavior: "multilingualPageButton6Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25031/promotion/?code=c020100000&lang=02"
  }, {
    index: 7,
    behavior: "multilingualPageButton7Click",
    linkUrl: "https://sg-play.mobilelegends.com/events/mlbb25030/promotion/?code=c020100000&lang=02"
  }]);

  useEffect(() => {
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25031",
      "behavior": `multilingualPage`,
      "url": window.location.href
    });
  }, []);

  function onLangClick(index) {
    const item = useLangRef.current.find(it => it.index == index);
    fetchSDKTrackingPoint(CONSTANT_OPTIONS.projectId, {
      "proj": "mlbb",
      "act_type": "mlbb25030",
      "behavior": item.behavior,
      "url": window.location.href
    });

    window.open(item.linkUrl, '_blank');
  }
  return (
    <div className="df social-channels-wrapper">
      <div className="banner-box"></div>
      <div className="tac desc-box">Please select your event <strong>language</strong></div>
      <div className="df social-btns-box">
        {/* 点击按钮就跳转七个按钮对应的链接 */}
        <a className="social-btn link-lang-03" onClick={() => onLangClick(1)}></a>
        <a className="social-btn link-lang-04" onClick={() => onLangClick(2)}></a>
        <a className="social-btn link-lang-07" onClick={() => onLangClick(3)}></a>
        <a className="social-btn link-lang-06" onClick={() => onLangClick(4)}></a>
        <a className="social-btn link-lang-08" onClick={() => onLangClick(5)}></a>

        <span className="social-btn-split"></span>
        <a className="social-btn link-lang-031-02" onClick={() => onLangClick(6)}></a>
        <a className="social-btn link-lang-030-02" onClick={() => onLangClick(7)}></a>
      </div>
    </div>
  );
}