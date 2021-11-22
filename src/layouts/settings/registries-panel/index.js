import React, { useState } from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import Modal, { ButtonDefinition, KeyDownDefinition } from '../../../components/modal';
import { IoLogoDocker } from 'react-icons/io5';
import AddValueButton from '../../../components/add-button';
import FlexBox from '../../../components/flexbox';
import {SecretsDeleteButton} from '../secrets-panel';
import Alert from '../../../components/alert';
import { useRegistries } from 'direktiv-react-hooks';
import { Config } from '../../../util';

function RegistriesPanel(props){
    const {namespace} = props
    const {data, err, getRegistries, createRegistry, deleteRegistry}  = useRegistries(Config.url, namespace)

    const [url, setURL] = useState("")
    const [username, setUsername] = useState("")
    const [token, setToken] = useState("")

    console.log("Registries", err)
    return (
        <ContentPanel style={{width: "100%", minHeight: "180px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Container Registries  
                </FlexBox>
                <div>
                    <Modal title="New registry"
                        escapeToCancel
                        button={(
                            <AddValueButton label=" " />
                        )} 
                        onClose={()=>{
                            setURL("")
                            setToken("")
                            setUsername("")
                        }}
                        keyDownActions={[
                            KeyDownDefinition("Enter", async () => {
                                let err = await createRegistry(url, `${username}:${token}`)
                                if(err) return err
                                await getRegistries()
                            }, true)
                        ]}
                        actionButtons={[
                            ButtonDefinition("Add", async() => {
                                let err = await createRegistry(url, `${username}:${token}`)
                                if(err) return err
                                await  getRegistries()
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", true, false)
                        ]}
                    >
                        <AddRegistryPanel token={token} setToken={setToken} username={username} setUsername={setUsername} url={url} setURL={setURL}/>    
                    </Modal> 
                </div>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox>
                        {data !== null ? 
                        <Registries deleteRegistry={deleteRegistry} getRegistries={getRegistries} registries={data}/>
                            :""}
                    </FlexBox>
                    <FlexBox style={{maxHeight: "44px"}}>
                        <Alert>Once a registry is removed, it can never be restored.</Alert>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default RegistriesPanel;

function AddRegistryPanel(props) {
    const {url, setURL, token, setToken, username, setUsername} = props

    return (
        <FlexBox className="col gap" style={{fontSize: "12px"}}>
            <FlexBox className="gap">
                <FlexBox>
                    <input value={url} onChange={(e)=>setURL(e.target.value)} autoFocus placeholder="Enter URL" />
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox><input value={username} onChange={(e)=>setUsername(e.target.value)} placeholder="Enter username" /></FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox><input value={token} onChange={(e)=>setToken(e.target.value)} type="password" placeholder="Enter token" /></FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

function Registries(props) {
    const {registries, deleteRegistry, getRegistries} = props

    return(
        <>
            <FlexBox className="col gap">
            {registries.map((obj)=>{
                    return (
                        <FlexBox key={obj.name} className="secret-tuple">
                            <FlexBox className="key">{obj.name}</FlexBox>
                            <FlexBox className="val"><span>******</span></FlexBox>
                            <FlexBox className="val"><span>******</span></FlexBox>
                            <FlexBox className="actions">
                                <Modal 
                                    escapeToCancel
                                    style={{
                                        flexDirection: "row-reverse",
                                        marginRight: "8px"
                                    }}
                                    title="Remove registry" 
                                    button={(
                                        <SecretsDeleteButton/>
                                    )} 
                                    actionButtons={
                                        [
                                            // label, onClick, classList, closesModal, async
                                            ButtonDefinition("Delete", async () => {
                                                let err = await deleteRegistry(obj.name)
                                                if (err) return err
                                                await getRegistries()
                                            }, "small red", true, false),
                                            ButtonDefinition("Cancel", () => {
                                            }, "small light", true, false)
                                        ]
                                    }   
                                >
                                    <FlexBox className="col gap">
                                        <FlexBox>
                                            Are you sure you want to remove '{obj.name}'?
                                            <br/>
                                            This action cannot be undone.
                                        </FlexBox>
                                    </FlexBox>
                                </Modal>
                            </FlexBox>
                        </FlexBox>
                    )
                })}
            </FlexBox>
        </>
    );
}