import { useEffect, useCallback, useImperativeHandle, forwardRef } from 'react';
import { useReactive } from 'ahooks';

import './index.less';
import { LANGUAGE_MODE } from '@/stores/i18n';

const LanguageModal = forwardRef((props, ref) => {
  const viewData = useReactive({
    visible: false,
    ln: "02",
    queryParams: {}
  });

  useEffect(() => {
    // console.log('props.slm', props.slm);
    viewData.ln = props.lang;
  }, [props.lang]);

  const handleClose = useCallback(() => {
    viewData.visible = false;
  }, []);

  const handleNo = useCallback(() => {
    handleClose();
    props.onNo?.();
  }, [props.onNo]);

  const handleYes = useCallback(() => {
    handleClose();
    props.onYes?.();
  }, [props.onYes]);

  useImperativeHandle(ref, () => ({
    open: () => (viewData.visible = true),
    close: handleClose
  }));

  if (!viewData.visible) return null;

  return (
    <div className="pf language-modal">
      <div className="modal-overlay" />
      <div className="modal-content">
        <div className="modal-close-btn" onClick={handleClose} />
        <div className="modal-body">
          {
            props.slm?.langthContent?.map((it, idx) => {
              return (
                <>
                  <div key={idx} className="df itText">
                    <span className="val" dangerouslySetInnerHTML={{ __html: (typeof it.text === "function" ? it.text({ langName: LANGUAGE_MODE[viewData.ln] }) : it.text) }}></span>
                  </div>
                </>
              )
            })
          }
        </div>
        <div className="modal-footer modal-buttons">
          <div className={`modal-btn modal-btn-no-${viewData.ln}`} onClick={handleNo} />
          <div className={`modal-btn modal-btn-yes-${viewData.ln}`} onClick={handleYes} />
        </div>
      </div>
    </div>
  );
});

export default LanguageModal; 