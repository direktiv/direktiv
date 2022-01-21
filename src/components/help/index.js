import React from 'react';
import { VscInfo } from 'react-icons/vsc';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css'
import './style.css';

function HelpIcon(props) {

    let {msg} = props;
    if (!msg) {
        msg = "No help text provided."
    }

    return (
        <>
            <Tippy content={msg} trigger={'mouseenter focus click'} zIndex={10}>
                <div className={"iconWrapper"}>
                    <VscInfo className="grey-text"/>
                </div>
            </Tippy>
        </>
    )
}

export default HelpIcon;