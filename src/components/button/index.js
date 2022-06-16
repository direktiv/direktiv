import React from 'react';
import './style.css';
import Tippy from '@tippyjs/react';

function Button(props) {

    let { children, onClick, style, className, title, tip } = props;
    let loader = (<></>);

    if (className) {
        if (className.includes("btn-loading")) {
            loader = (
                <div className="btn-loader" />
            )
        }
    }

    return (
        <>
            {tip ?
                <Tippy content={tip} trigger={'mouseenter focus'} zIndex={10} delay={[500, 0]}>
                    <div onClick={onClick} className={"btn " + className} style={style} title={title}>
                        {children}
                        {loader}
                    </div>
                </Tippy>
                :
                <div onClick={onClick} className={"btn " + className} style={style} title={title}>
                    {children}
                    {loader}
                </div>
            }
        </>
    );
}

export default Button;