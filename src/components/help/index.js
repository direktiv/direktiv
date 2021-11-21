import React from 'react';
import {IoMdHelpCircle} from 'react-icons/io';

function HelpIcon(props) {

    let {msg} = props;
    if (!msg) {
        msg = "No help text provided."
    }

    return (<>
        <IoMdHelpCircle className="grey-text" title={msg} />
    </>)
}

export default HelpIcon;