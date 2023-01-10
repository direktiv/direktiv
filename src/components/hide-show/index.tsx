import React from 'react';
import { BsEye, BsEyeSlash } from 'react-icons/bs';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css'
import './style.css';

export interface HideShowButtonProps {
    /**
    * Sets field name that will be used on tooltip.
    */
    field: string
    /**
    * z-index of component.
    */
    zIndex: number
    /**
    * If button is currenlty is show or hide state.
    */
    show: boolean
    /**
    * OnClick callback to set show state. The show state should be declared on the parent.
    */
    setShow?: React.Dispatch<React.SetStateAction<boolean>>
}

/**
* Simple Hide/Show button that can be used for input fields with sensitive data.
*/
function HideShowButton({field = "field", zIndex = 10, show = false, setShow}: HideShowButtonProps) {
    return (
        <>
            <Tippy content={show ? `Hide ${field}` : `Show ${field}`} trigger={'mouseenter focus'} zIndex={zIndex}>
                <div className={"show-hide-icon"} onClick={() => {
                    if (!setShow) {
                        console.warn("setShow prop missing")
                        return
                    }

                    setShow(!show)
                }}>
                    {show ?
                        <BsEye className="grey-text" />
                        :
                        <BsEyeSlash className="grey-text" />
                    }
                </div>
            </Tippy>
        </>
    )
}

export default HideShowButton;