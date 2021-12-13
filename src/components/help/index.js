import React from 'react';
import { VscInfo } from 'react-icons/vsc';

function HelpIcon(props) {

    let {msg} = props;
    if (!msg) {
        msg = "No help text provided."
    }

    return (<>
        <VscInfo className="grey-text" title={msg} />
    </>)
}

export default HelpIcon;