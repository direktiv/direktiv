import React, {useState} from 'react';
import './style.css';
import AddValueButton from '../add-button';
import Button from '../button';
import ContentPanel, {ContentPanelTitle, ContentPanelBody, ContentPanelTitleIcon} from '../../components/content-panel';
import { IoLockClosedOutline, IoCloseCircleSharp } from 'react-icons/io5';
import FlexBox from '../flexbox';


function Modal(props) {

    let {title, children, button, withCloseButton, activeOverlay, label} = props;
    let {actionButtonLabel, actionButtonFunc} = props;
    const [visible, setVisible] = useState(false);

    if (!title) {
        title = "Modal Title"
    }

    if (!button) {
        
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
                        actionButtonLabel={actionButtonLabel}
                        actionButtonFunc={actionButtonFunc}
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
        <FlexBox onClick={() => {
            setVisible(true)
        }}>
            {button}
        </FlexBox>
        </>
    )
}

export default Modal;

function ModalOverlay(props) {

    let {title, children, callback, activeOverlay, withCloseButton} = props;
    let {actionButtonLabel, actionButtonFunc} = props;

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
                        <ContentPanel style={{ maxWidth: "250px" }}>
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
                                    { actionButtonLabel ? (
                                        <FlexBox className="gap" style={{flexDirection: "row-reverse"}}>
                                            <Button className="small red" onClick={() => {
                                                actionButtonFunc()
                                                callback()
                                            }}>{actionButtonLabel}</Button>
                                        </FlexBox>
                                    ) :<></>}
                                </FlexBox>
                            </ContentPanelBody>
                        </ContentPanel>
                    </div>
                </FlexBox>
            </div>
        </>
    )
}