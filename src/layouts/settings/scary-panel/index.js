import React, { useState } from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody} from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';
import Button from '../../../components/button';
import Modal, { ButtonDefinition } from '../../../components/modal';
import { VscSettingsGear, VscTrash } from 'react-icons/vsc';

function ScarySettings(props) {
    const {deleteNamespace, namespace, deleteErr} = props
    return (<>
        <div>
            <ContentPanel className="scary-panel">
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscSettingsGear className="red-text" />
                    </ContentPanelTitleIcon>
                    <FlexBox className="red-text">
                        Important Settings   
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel">
                    <FlexBox className="gap col">
                        <FlexBox className="scary-settings"> 
                            <Scary namespace={namespace} deleteErr={deleteErr} deleteNamespace={deleteNamespace}/>
                        </FlexBox>
                        <FlexBox>
                            <Alert style={{width: "100%"}} className="critical">These settings are super dangerous! Use at your own risk!</Alert>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </div>
    </>)
}

export default ScarySettings;

function Scary(props) {
    const {deleteNamespace, namespace} = props
    const [delButtonEnabled, setDelButtonEnabled] = useState(false)
    // deleteErr gets filled in when someone attempts to delete a namespace and an error happens

    let delBtnClasses = "small red";
    if (!delButtonEnabled) {
        delBtnClasses += " disabled"
    }

    return(
        <>
        <FlexBox>
            <FlexBox className="auto-margin" style={{fontSize: "12px", maxWidth: "300px"}}>
                This will permanently delete the current active namespace and all resources associated with it.
            </FlexBox>
            <FlexBox>
                <Modal title="Delete namespace" 
                        escapeToCancel
                        titleIcon={<VscTrash/>}
                        button={(
                            <Button className="auto-margin small red" style={{
                                minWidth: "120px",
                                textAlign: "center"
                            }}>
                                Delete Namespace
                            </Button>
                        )}  
                        actionButtons={[
                            ButtonDefinition("Delete", async () => {
                                await deleteNamespace(namespace)
                            }, delBtnClasses,()=>{}, true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light",()=>{}, true, false)
                        ]}
                    >
                        <DeleteNamespaceConfirmationPanel namespace={namespace} setDelButtonEnabled={setDelButtonEnabled} />
                    </Modal>
            </FlexBox>
        </FlexBox>
        </>
    );
}


function DeleteNamespaceConfirmationPanel(props) {

    let {namespace, setDelButtonEnabled} = props;

    const [inputValue, setInputValue] = useState("")

    return (
        <FlexBox className="col" style={{fontSize: "12px"}}>
            <FlexBox className="col">
                <p>
                    Are you sure you want to delete this namespace?<br/> This action <b>can not be undone!</b>
                </p>
                <br/>
                <p>
                    Please type <b>{namespace}</b> to confirm.
                </p>
            </FlexBox>
            <FlexBox>
                <input ovalue={inputValue} onChange={(e)=> {
                    setInputValue(e.target.value)
                    if (e.target.value === namespace) {
                        setDelButtonEnabled(true)
                    } else {
                        setDelButtonEnabled(false)
                    }
                }} type="text"></input>
            </FlexBox>
        </FlexBox>
    );
}