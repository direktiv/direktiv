import React, {useEffect, useState} from 'react';
import './style.css';
import Button from '../button';
import ContentPanel, {ContentPanelTitle, ContentPanelBody, ContentPanelTitleIcon, ContentPanelFooter} from '../../components/content-panel';

import { VscDiffAdded } from 'react-icons/vsc';

import FlexBox from '../flexbox';
import Alert from '../alert';
import { VscClose } from 'react-icons/vsc';

import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css'

export function ModalHeadless(props) {
    let {maximised, noPadding, titleIcon, title, children, withCloseButton, activeOverlay, label} = props;
    let {modalStyle, actionButtons, keyDownActions, escapeToCancel, onClose, onOpen, requiredFields, visible, setVisible } = props;

    if (!title) {
        title = "Modal Title"
    }
    
    if (!label) {
        label = "Click me"
    }

    function closeModal() {
        setVisible(false);
    }

    

    let callback = function() {
        if (onClose) {
            onClose()
        }
        closeModal()
    }

    let overlay = (<></>);
    if (visible) {
        overlay = (<ModalOverlay 
                        maximised={maximised}
                        modalStyle={modalStyle}
                        children={children} 
                        title={title} 
                        activeOverlay={activeOverlay} 
                        withCloseButton={withCloseButton} 
                        callback={callback} 
                        onOpen={onOpen}
                        actionButtons={actionButtons}
                        requiredFields={requiredFields}
                        escapeToCancel={escapeToCancel}
                        keyDownActions={keyDownActions}
                        titleIcon={titleIcon}
                        noPadding={noPadding}
                    />)
    }

    return (<>{overlay}</>)


}

function Modal(props) {

    let {maximised, noPadding, titleIcon, title, children, button, withCloseButton, activeOverlay, label, buttonDisabled} = props;
    let {modalStyle, style, actionButtons, keyDownActions, escapeToCancel, onClose, onOpen, requiredFields } = props;
    const [visible, setVisible] = useState(false);
    if (!button) {
        return(
            <div>
                 <ModalHeadless 
        setVisible={setVisible}
        visible={visible}
        maximised={maximised}
        noPadding={noPadding}
        titleIcon={titleIcon}
        title={title}
        children={children}
        button={button}
        withCloseButton={withCloseButton}
        activeOverlay={activeOverlay}
        label={label}
        buttonDisabled={buttonDisabled}
        modalStyle={modalStyle}
        style={style}
        actionButtons={actionButtons}
        keyDownActions={keyDownActions}
        escapeToCancel={escapeToCancel}
        onClose={onClose}
        onOpen={onOpen}
        requiredFields={requiredFields}/>
                <Button style={{pointerEvents: buttonDisabled ? "none" : ""}} onClick={(ev) => {
                    setVisible(true)
                    ev.stopPropagation()
                }}>
                    {label}
                </Button>
            </div>
        );
    }

    return (
        <>
        
        <ModalHeadless 
        setVisible={setVisible}
        visible={visible}
        maximised={maximised}
        noPadding={noPadding}
        titleIcon={titleIcon}
        title={title}
        children={children}
        button={button}
        withCloseButton={withCloseButton}
        activeOverlay={activeOverlay}
        label={label}
        buttonDisabled={buttonDisabled}
        modalStyle={modalStyle}
        style={style}
        actionButtons={actionButtons}
        keyDownActions={keyDownActions}
        escapeToCancel={escapeToCancel}
        onClose={onClose}
        onOpen={onOpen}
        requiredFields={requiredFields}/>
        <FlexBox style={{...style}}>
            <div style={{width: "100%", display:'flex', justifyContent: "center", pointerEvents: buttonDisabled ? "none" : ""}} onClick={async(ev) => {
                if(onOpen){
                    await onOpen()
                }
                setVisible(true)
                ev.stopPropagation()
            }}>
                {button}
            </div>
        </FlexBox>
        </>
    )
        }
export default Modal;



function ModalOverlay(props) {
    function validateFields(requiredFields) {
        let tipMessages = []
        
        if (!requiredFields) {
            return {tips: tipMessages, valid: tipMessages.length===0}
        }

        for (let i = 0; i < requiredFields.length; i++) {
            const rField = requiredFields[i];
            if (rField.value === null || rField.value === "") {
                tipMessages.push(rField.tip)
            }
        }
    
        return {tips: tipMessages, valid: tipMessages.length===0}
    }

    let {maximised, noPadding, titleIcon, modalStyle, title, children, callback, activeOverlay, withCloseButton} = props;
    let {actionButtons, escapeToCancel, keyDownActions, requiredFields} = props;
    const [displayAlert, setDisplayAlert] = useState(false);
    const [alertMessage, setAlertMessage] = useState("");

    const validateResults = validateFields(requiredFields)

    
    useEffect(()=>{
        function closeModal(e){
            if (e.keyCode === 27) {
                callback(false)
            }
        }

        let removeListeners = [];

        if (escapeToCancel) {
            window.addEventListener('keydown', closeModal)
            removeListeners.push({label: 'keydown', fn: closeModal})
        }

        if (keyDownActions) {
            for (let i = 0; i < keyDownActions.length; i++) {
                const action = keyDownActions[i];

                let fn = async (e) => {
                    if (action.id !== undefined && action.id !== e.target.id){
                        return
                    }

                    if (e.code === action.code) {
                        try { 
                            const result = await action.fn() 
                            if (!result?.error && action.closeModal) {
                                callback(false)
                            }
                            if(result?.error){
                                setAlertMessage(result?.msg)
                                setDisplayAlert(true)
                            }
                        } catch(err) {
                            setAlertMessage(err.toString())
                            setDisplayAlert(true)
                        }
                    }
                }

                window.addEventListener('keydown', fn)
                removeListeners.push({label: 'keydown', fn: fn})
            }
        }

        return () => {
            for (let i = 0; i < removeListeners.length; i++) {
                window.removeEventListener(removeListeners[i].label, removeListeners[i].fn)
            }
        }

    },[escapeToCancel, callback, keyDownActions])
    

    let overlayClasses = ""
    let closeButton = (<></>);
    if (withCloseButton) {
        closeButton = (
            <FlexBox className="modal-buttons" style={{ flexDirection: "column-reverse" }}>
                <div>
                    <VscClose onClick={()=>{
                        callback()
                    }} className="auto-margin" style={{marginRight:"8px"}} />
                </div>
            </FlexBox>
        )
    }
    
    if (activeOverlay) {
        overlayClasses += "clickable"
    }

    let buttons
    if (actionButtons){
       buttons = generateButtons(callback, setDisplayAlert, setAlertMessage, actionButtons, validateResults);
    } 

    let contentBodyStyle = {}
    if (!noPadding) {
        contentBodyStyle = {
            padding: "12px"
        }
    }


    let panelStyle = {maxHeight: "90vh", height: "100%", minWidth: "20vw", maxWidth: "80vw", overflowY: "auto"}
    if (maximised) {
        panelStyle = { 
            ...panelStyle, 
            height: "90vh",
            width: "90vw"
        }
    }

    return(
        <>
            <div className={"modal-overlay " + overlayClasses} />
            <div className={"modal-container " + overlayClasses} onClick={() => {
                if (activeOverlay) {
                    callback()
                }
            }}>
                <FlexBox className="tall">
                    <div style={{ display: "flex", width: "100%", justifyContent: "center", ...modalStyle }} className="modal-body auto-margin" onClick={(e) => {
                        e.stopPropagation()
                    }}>
                        <ContentPanel style={{...panelStyle, width: "100%"}}>
                            <ContentPanelTitle>
                                <FlexBox style={{ maxWidth: "18px" }}>
                                    <ContentPanelTitleIcon>
                                        {
                                            titleIcon 
                                            ? 
                                            [titleIcon]
                                            :
                                            <VscDiffAdded />
                                        }
                                    </ContentPanelTitleIcon>
                                </FlexBox>
                                <FlexBox>
                                    {title}   
                                </FlexBox>
                                <FlexBox>
                                    {closeButton}
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody style={{...contentBodyStyle, overflow: "auto"}}>
                                <FlexBox className="col gap">
                                    { displayAlert ?
                                    <Alert  className="critical">{alertMessage}</Alert>
                                    : <></> }
                                    {children}
                                </FlexBox>
                            </ContentPanelBody>
                            { buttons ? 
                            <div>
                                <ContentPanelFooter>
                                    <FlexBox className="gap modal-buttons-container" style={{flexDirection: "row-reverse"}}>
                                        {buttons}
                                    </FlexBox>
                                </ContentPanelFooter>
                            </div>
                            :<></>}
                        </ContentPanel>
                    </div>
                </FlexBox>
            </div>
        </>
    )
}

export function ButtonDefinition(label, onClick, classList, errFunc, closesModal, async, validate) {
    return {
        label: label,
        onClick: onClick,
        classList: classList,
        errFunc: errFunc,
        closesModal: closesModal,
        async: async,
        validate: validate
    }
}
// KeyDownDefinition :
// code : Target Key Event
// fn : callback function
// closeModal : Whether to close modal after fn()
// id : target element id to listen on. If undefined listener is global
export function KeyDownDefinition(code, fn, errFunc, closeModal, targetElementID) {
    return {
        code: code,
        fn: fn,
        errFunc: errFunc,
        closeModal: closeModal,
        id: targetElementID,
    }
}

function generateButtons(closeModal, setDisplayAlert, setAlertMessage, actionButtons, validateResults) {

    // label, onClick, classList, closesModal, async


    let out = [];
    for (let i = 0; i < actionButtons.length; i++) {

        let btn = actionButtons[i];

        let onClick =  async () => {
            try {
                await btn.onClick()
                if(btn.closesModal){
                    closeModal()
                } else {
                    setAlertMessage("")
                    setDisplayAlert(false)
                }
            } catch(e){
                btn.errFunc()
                // handle error
                if(e.message){
                    setAlertMessage(e.message)
                } else {
                    setAlertMessage(e.toString())
                }
                setDisplayAlert(true)
            }
   
        }

        out.push(
            !validateResults.valid && btn.validate ? 
            <Tippy content={`${validateResults.tips.join(", ")}`} trigger={'mouseenter focus click'} zIndex={10}>
                <div>
                <Button key={Array(5).fill().map(()=>"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(Math.random()*62)).join("")} className={`disabled ${btn.classList}`} onClick={onClick}>
                    <div>{btn.label}</div>
                </Button>
                </div>
            </Tippy>
            :
            <Button key={Array(5).fill().map(()=>"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(Math.random()*62)).join("")} className={`${btn.classList}`} onClick={onClick}>
                    <div>{btn.label}</div>
            </Button>
        )
    }

    return out
} 