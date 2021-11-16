import React, {useState} from 'react';
import './style.css';
import Button from '../button';
import ContentPanel, {ContentPanelTitle, ContentPanelBody, ContentPanelTitleIcon} from '../../components/content-panel';
import { IoLockClosedOutline, IoCloseCircleSharp } from 'react-icons/io5';
import FlexBox from '../flexbox';

function Modal(props) {

    let {withCloseButton} = props;
    const [visible, setVisible] = useState(false);

    function closeModal() {
        setVisible(false);
    }

    let overlay = (<></>);
    if (visible) {
        overlay = (<ModalOverlay withCloseButton={withCloseButton} callback={closeModal} />)
    }

    return(
        <>
            {overlay}
            <Button onClick={() => {
                setVisible(true)
            }}>
                Modal Open
            </Button>
        </>
    );
}

export default Modal;

function ModalOverlay(props) {

    let {callback, withCloseButton} = props;

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
    } else {
        overlayClasses += "clickable"
    }

    return(
        <>
            <div className={"modal-overlay " + overlayClasses} />
            <div className={"modal-container " + overlayClasses} onClick={() => {
                if (!withCloseButton) {
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
                                    Modal Title   
                                </FlexBox>
                                <FlexBox>
                                    {closeButton}
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody >
                                <FlexBox>
                                    Contents...
                                </FlexBox>
                            </ContentPanelBody>
                        </ContentPanel>
                    </div>
                </FlexBox>
            </div>
        </>
    )
}