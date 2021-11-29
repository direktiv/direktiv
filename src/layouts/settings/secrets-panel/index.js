import React, { useState } from 'react';
import './style.css';
import AddValueButton from '../../../components/add-button';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoLockClosedOutline } from 'react-icons/io5';
import {RiDeleteBin2Line} from 'react-icons/ri';
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
    const {data, err, createSecret, deleteSecret, getSecrets} = useSecrets(Config.url, namespace)
   
    console.log("Secrets", err, data)
    // createErr is the error when creating a secret

    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
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

                        modalStyle={{
                            maxWidth: "300px"
                        }}

                        onOpen={() => {
                            console.log("ON OPEN");
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
    const {secrets, deleteSecret, getSecrets} = props

    return(
        <>
            <FlexBox className="col gap" style={{ maxHeight: "236px", overflowY: "auto" }}>
                    {secrets.length === 0 ?
                             <FlexBox className="secret-tuple" >
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
                                        title="Remove secret" 
                                        button={(
                                            <SecretsDeleteButton/>
                                        )} 
                                        actionButtons={
                                            [
                                                // label, onClick, classList, closesModal, async
                                                ButtonDefinition("Delete", async () => {
                                                    console.log("DELETE FUNC");
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
            <RiDeleteBin2Line/>
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