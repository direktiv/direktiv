import React, { useState } from 'react';
import './style.css';
import AddValueButton from '../../../components/add-button';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import {VscLock, VscTrash} from 'react-icons/vsc'
import Modal, {ButtonDefinition, KeyDownDefinition} from '../../../components/modal';
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';
import {useSecrets} from 'direktiv-react-hooks'
import {Config, GenerateRandomKey} from '../../../util'
import HelpIcon from '../../../components/help';


function SecretsPanel(props){
    const {namespace} = props


    const [keyValue, setKeyValue] = useState("")
    const [vValue, setVValue] = useState("")
    const {data, createSecret, deleteSecret, getSecrets} = useSecrets(Config.url, namespace, localStorage.getItem("apikey"))
   
    // createErr is the error when creating a secret

    return (
        <ContentPanel style={{ height: "100%", minHeight: "180px", width: "100%" }}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscLock />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Secrets
                    </div>
                    <HelpIcon msg={"Encrypted key/value pairs that can be referenced within workflows. Suitable for storing sensitive information (such as tokens) for use in workflows."} />
                </FlexBox>
                <div>
                    <Modal title="New secret" 
                        escapeToCancel
                        titleIcon={<VscLock/>}
                        modalStyle={{
                            maxWidth: "300px"
                        }}

                        onOpen={() => {
                        }}

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
                                await getSecrets()
                            }, true)
                        ]}
                        
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                let err = await createSecret(keyValue, vValue)
                                if(err) return err
                                await  getSecrets()
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
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
                    {data !== null ? 
                        <Secrets deleteSecret={deleteSecret} getSecrets={getSecrets} secrets={data}  />: ""}
                    </FlexBox>
                    <div>
                        <Alert>Once a secret is removed, it can never be restored.</Alert>
                    </div>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default SecretsPanel;

function Secrets(props) {
    const {secrets, deleteSecret, getSecrets} = props

    return(
        <>
            <FlexBox className="col gap" style={{ maxHeight: "236px", overflowY: "auto" }}>
                    {secrets.length === 0 ?
                             <FlexBox className="secret-tuple empty-content" >
                             <FlexBox className="key">No secrets are stored...</FlexBox>
                             <FlexBox className="val"></FlexBox>
                             <FlexBox className="actions">
                             </FlexBox>
                         </FlexBox>
                    :
                    <>
                    {secrets.map((obj)=>{

                        let key = GenerateRandomKey("secret-")

                        return (
                            <FlexBox className="secret-tuple" key={key} id={key}>
                                <FlexBox className="key">{obj.node.name}</FlexBox>
                                <FlexBox className="val"><span>******</span></FlexBox>
                                <FlexBox className="actions">
                                    <Modal 
                                        escapeToCancel
                                        style={{
                                            flexDirection: "row-reverse",
                                            marginRight: "8px"
                                        }}
                                        titleIcon={<VscLock/>}
                                        title="Remove secret" 
                                        button={(
                                            <SecretsDeleteButton/>
                                        )} 
                                        actionButtons={
                                            [
                                                // label, onClick, classList, closesModal, async
                                                ButtonDefinition("Delete", async () => {
                                                    let err = await deleteSecret(obj.node.name)
                                                    if (err) return err
                                                    await getSecrets()
                                                }, "small red", true, false),
                                                ButtonDefinition("Cancel", () => {
                                                }, "small light", true, false)
                                            ]
                                        }   
                                    >
                                        <FlexBox className="col gap">
                                            <FlexBox >
                                                Are you sure you want to delete '{obj.node.name}'?
                                                <br/>
                                                This action cannot be undone.
                                            </FlexBox>
                                        </FlexBox>
                                    </Modal>
                                </FlexBox>
                            </FlexBox>
                        )
                    })}</>}
            </FlexBox>
        </>
    );
}

export function SecretsDeleteButton(props) {
    return (
        <div className="secrets-delete-btn red-text">
            <VscTrash />
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