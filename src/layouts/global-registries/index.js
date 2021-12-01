import React, { useState } from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../components/content-panel';
import Modal, { ButtonDefinition, KeyDownDefinition } from '../../components/modal';
import { IoLogoDocker } from 'react-icons/io5';
import AddValueButton from '../../components/add-button';
import FlexBox from '../../components/flexbox';
import Alert from '../../components/alert';
import { useGlobalRegistries, useGlobalPrivateRegistries } from 'direktiv-react-hooks';
import {AddRegistryPanel, Registries} from '../settings/registries-panel'
import { Config } from '../../util';
import HelpIcon from '../../components/help';


export default function GlobalRegistriesPanel(){
    return(
        <FlexBox className="gap wrap" style={{ paddingRight: "8px" }}>
            <FlexBox  style={{ minWidth: "380px" }}>
                <GlobalRegistries />
            </FlexBox>
            <FlexBox style={{minWidth:"380px"}}>
                <GlobalPrivateRegistries />
            </FlexBox>
        </FlexBox>
    )
}

export function GlobalRegistries(){

    const {data, err, getRegistries, createRegistry, deleteRegistry} = useGlobalRegistries(Config.url, localStorage.getItem("apikey"))

    const [url, setURL] = useState("")
    const [username, setUsername] = useState("")
    const [token, setToken] = useState("")

    console.log(data, err)
    return (
        <ContentPanel style={{width: "100%", minHeight: "180px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Container Registries  
                    </div>
                    <HelpIcon msg={"Add a registry that can be accessed by any service"} />
                </FlexBox>
                <div>
                    <Modal title="New registry"
                        escapeToCancel
                        modalStyle={{
                            maxWidth: "450px",
                            minWidth: "450px"
                        }}
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

export function GlobalPrivateRegistries(){

    const {data, err, getRegistries, createRegistry, deleteRegistry} = useGlobalPrivateRegistries(Config.url)

    const [url, setURL] = useState("")
    const [username, setUsername] = useState("")
    const [token, setToken] = useState("")

    console.log(data, err)
    return (
        <ContentPanel style={{width: "100%", minHeight: "180px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Private Container Registries  
                    </div>
                    <HelpIcon msg={"Add a registry that is only available to global services"} />
                </FlexBox>
                <div>
                    <Modal title="New registry"
                        escapeToCancel
                        modalStyle={{
                            maxWidth: "450px",
                            minWidth: "450px"
                        }}
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