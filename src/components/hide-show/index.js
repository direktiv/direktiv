import React from 'react';
import { BsEye, BsEyeSlash } from 'react-icons/bs';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css'
import './style.css';

function HideShowButton(props) {

    const {field, zIndex, show, setShow} = props;

    return (
        <>
            <Tippy content={show ? `Hide ${field ? field : "field" }` : `Show ${field ?  field : "field"}`} trigger={'mouseenter focus'} zIndex={zIndex ? zIndex : 10}>
                <div className={"show-hide-icon"} onClick={()=>{
                    if (!setShow){
                        console.warn("setShow prop missing")
                        return
                    }
                    
                    setShow(!show)
                }}>
                    {show ?
                        <BsEye className="grey-text"/>
                        :
                        <BsEyeSlash className="grey-text"/>
                    }
                </div>
            </Tippy>
        </>
    )
}

export default HideShowButton;