import React, { useState } from 'react';
import './style.css';
import AddValueButton from '../../../components/add-button';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoCloseCircleSharp, IoLockClosedOutline } from 'react-icons/io5';
import Modal, {ButtonDefinition, KeyDownDefinition} from '../../../components/modal';
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';
import {useSecrets} from 'direktiv-react-hooks'
import {Config} from '../../../util'

function SecretsPanel(props){
    const {namespace} = props


    const [keyValue, setKeyValue] = useState("")
    const [vValue, setVValue] = useState("")
    const {data, err, createSecret, deleteSecret, getSecrets} = useSecrets(Config.url, namespace)
   
    console.log("Secrets", err, data)
    // createErr is the error when creating a secret

    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Secrets   
                </FlexBox>
                <div>
                    <Modal title="New secret" 
                        escapeToCancel
                        onClose={()=>{
                            setKeyValue("")
                            setVValue("")
                        }}
                        button={(
                            <AddValueButton label=" " />
                        )}  
                        keyDownActions={[
                            KeyDownDefinition("Enter", async () => {
                                let err = await createSecret(keyValue, vValue)
                                if(err) return err
                            }, true)
                        ]}
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                let err = await createSecret(keyValue, vValue)
                                if(err) return err
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                                setKeyValue("")
                                setVValue("")
                            }, "small light", true, false)
                        ]}
                    >
                        <AddSecretPanel keyValue={keyValue} vValue={vValue} setKeyValue={setKeyValue} setVValue={setVValue} />
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox className="secrets-list"> 
                        <Secrets  />
                    </FlexBox>
                    <FlexBox style={{maxHeight: "44px"}}>
                        <Alert>Once a secret is removed, it can never be restored.</Alert>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default SecretsPanel;

function Secrets(props) {

    return(
        <>
            <FlexBox className="col gap">
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">randomKey</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            escapeToCancel
                            style={{
                                flexDirection: "row-reverse",
                                marginRight: "8px"
                            }}
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )} 
                            actionButtons={
                                [
                                    // label, onClick, classList, closesModal, async
                                    ButtonDefinition("Delete", () => {
                                        console.log("DELETE FUNC");
                                    }, "small red", true, false),
                                    ButtonDefinition("Cancel", () => {
                                        console.log("DONT DELETE");
                                    }, "small light", true, false)
                                ]
                            }   
                        >
                            <FlexBox className="col gap">
                                <FlexBox >
                                    Are you sure you want to delete 'SECRET_NAME_HERE'?
                                    <br/>
                                    This action cannot be undone.
                                </FlexBox>
                            </FlexBox>
                        </Modal>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    );
}

export function SecretsDeleteButton(props) {
    return (
        <div className="secrets-delete-btn red-text">
            <IoCloseCircleSharp/>
        </div>
    )
}

function AddSecretPanel(props) {
    const {keyValue, vValue, setKeyValue, setVValue} = props


    return (
        <FlexBox className="col gap" style={{fontSize: "12px"}}>
            <FlexBox className="gap">
                <FlexBox>
                    <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter key" />
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox><input type="password"  value={vValue} onChange={(e)=>setVValue(e.target.value)} placeholder="Enter value" /></FlexBox>
            </FlexBox>
        </FlexBox>
    );
}