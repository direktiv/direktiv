import React, {useState} from 'react';
import './style.css';
import Button from '../button';
import ContentPanel, {ContentPanelTitle, ContentPanelBody, ContentPanelTitleIcon} from '../../components/content-panel';
import { IoLockClosedOutline, IoCloseCircleSharp } from 'react-icons/io5';
import FlexBox from '../flexbox';


function Modal(props) {

    let {title, children, button, withCloseButton, activeOverlay, label} = props;
    let {actionButtons} = props;
    const [visible, setVisible] = useState(false);

    if (!title) {
        title = "Modal Title"
    }
    
    if (!label) {
        label = "Click me"
    }

    function closeModal() {
        setVisible(false);
    }

    let overlay = (<></>);
    if (visible) {
        overlay = (<ModalOverlay 
                        children={children} 
                        title={title} 
                        activeOverlay={activeOverlay} 
                        withCloseButton={withCloseButton} 
                        callback={closeModal} 
                        actionButtons={actionButtons}
                    />)
    }

    if (!button) {
        return(
            <>
                {overlay}
                <Button onClick={() => {
                    setVisible(true)
                }}>
                    {label}
                </Button>
            </>
        );
    }

    return (
        <>
        {overlay}
        <FlexBox>
            <div onClick={() => {
                setVisible(true)
            }}>
                {button}
            </div>
        </FlexBox>
        </>
    )
}

export default Modal;

function ModalOverlay(props) {

    let {title, children, callback, activeOverlay, withCloseButton} = props;
    let {actionButtons} = props;

    let overlayClasses = ""
    let closeButton = (<></>);
    if (withCloseButton) {
        closeButton = (
            <FlexBox className="modal-buttons" style={{ flexDirection: "column-reverse" }}>
                <div>
                    <IoCloseCircleSharp 
                        className="red-text auto-margin" 
                        style={{ marginRight: "8px" }}
                        onClick={() => {
                            callback()
                        }}
                    />
                </div>
            </FlexBox>
        )
    }
    
    if (activeOverlay) {
        overlayClasses += "clickable"
    }

    let buttons
    if (actionButtons){
       buttons = generateButtons(callback, actionButtons);
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
                    <div className="modal-body auto-margin" onClick={(e) => {
                        e.stopPropagation()
                    }}>
                        <ContentPanel>
                            <ContentPanelTitle>
                                <FlexBox style={{ maxWidth: "18px" }}>
                                    <ContentPanelTitleIcon>
                                        <IoLockClosedOutline />
                                    </ContentPanelTitleIcon>
                                </FlexBox>
                                <FlexBox>
                                    {title}   
                                </FlexBox>
                                <FlexBox>
                                    {closeButton}
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody >
                                <FlexBox className="col gap">
                                    {children}
                                    <FlexBox className="gap" style={{flexDirection: "row-reverse"}}>
                                        {buttons}
                                    </FlexBox>
                                </FlexBox>
                            </ContentPanelBody>
                        </ContentPanel>
                    </div>
                </FlexBox>
            </div>
        </>
    )
}

export function ButtonDefinition(label, onClick, classList, closesModal, async) {
    return {
        label: label,
        onClick: onClick,
        classList: classList,
        closesModal: closesModal,
        async: async
    }
}

function generateButtons(closeModal, actionButtons) {

    // label, onClick, classList, closesModal, async

    let out = [];
    for (let i = 0; i < actionButtons.length; i++) {

        let btn = actionButtons[i];
        let onClick = () => {
            btn.onClick()
            if (btn.closesModal) {
                closeModal()
            }
        }

        out.push(
            <Button key={Array(5).fill().map(()=>"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".charAt(Math.random()*62)).join("")} className={btn.classList} onClick={onClick}>
                {btn.label}
            </Button>
        )
    }

    return out
} 