import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody} from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import { IoLockClosedOutline } from 'react-icons/io5';
import Alert from '../../../components/alert';
import Button from '../../../components/button';
import Modal, { ButtonDefinition } from '../../../components/modal';

function ScarySettings(props) {
    const {deleteNamespace, namespace} = props
    return (<>
        <ContentPanel className="scary-panel">
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline className="red-text" />
                </ContentPanelTitleIcon>
                <FlexBox className="red-text">
                    Important Settings   
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox className="scary-settings"> 
                        <Scary namespace={namespace} deleteNamespace={deleteNamespace}/>
                    </FlexBox>
                    <FlexBox>
                        <Alert className="critical">The following settings are super dangerous! Use at your own risk!</Alert>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    </>)
}

export default ScarySettings;

function Scary(props) {
    const {deleteNamespace, namespace} = props
    return(
        <>
        <FlexBox>
            <FlexBox className="auto-margin" style={{fontSize: "12px", maxWidth: "300px"}}>
                This will permanently delete the current active namespace and all resources associated with it.
            </FlexBox>
            <FlexBox>
                <Modal title="Delete namespace" 
                        escapeToCancel
                        button={(
                            <Button className="auto-margin small red">
                                Delete Namespace
                            </Button>
                        )}  
                        actionButtons={[
                            ButtonDefinition("Delete", () => {
                                deleteNamespace(namespace)
                            }, "small red", true, false),
                            ButtonDefinition("Cancel", () => {
                                console.log("close modal");
                            }, "small light", true, false)
                        ]}
                    >
                        <DeleteNamespaceConfirmationPanel />
                    </Modal>
            </FlexBox>
        </FlexBox>
        </>
    );
}


function DeleteNamespaceConfirmationPanel(props) {

    return (
        <FlexBox style={{fontSize: "12px"}}>
            <p>
                Are you sure you want to delete this namespace?<br/> This action <b>can not be undone!</b>
            </p>
        </FlexBox>
    );
}