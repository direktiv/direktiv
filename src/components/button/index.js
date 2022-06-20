import React, { useMemo } from 'react';
import './style.css';
import Tippy from '@tippyjs/react';

function Button(props) {

    let { children, onClick, style, className, title, tip, disabledTip, loading, disabled } = props;

    const tippyMsg = useMemo(() => {
        if (disabledTip && disabled) {
            return disabledTip
        }

        if (tip) {
            return tip
        }

        return ""
    }, [tip, disabledTip, disabled])

    const tippyTrigger = useMemo(() => {
        if (tippyMsg !== "") {
            return 'mouseenter focus'
        }

        return ""
    }, [tippyMsg])

    const btnClassName = useMemo(() => {
        let newClassName = "btn " + className

        if (loading) {
            newClassName += " loading"
        }

        if (disabled) {
            newClassName += " disabled"
        }

        return newClassName
    }, [loading, disabled, className])

    const btnWrapperClassName = useMemo(() => {
        let newClassName = "btn-wrapper"

        if (loading) {
            newClassName += " loading"
        }

        return newClassName
    }, [loading])

    return (
        <>
            <Tippy content={tippyMsg} trigger={tippyTrigger} zIndex={10} delay={[200, 0]}>
                <div className={btnWrapperClassName}>
                    <div onClick={onClick} className={btnClassName} style={style} title={title}>
                        {children}
                    </div>
                </div>
            </Tippy>
        </>
    );
}

export default Button;

// ${loading ? " loading" : " loading"}