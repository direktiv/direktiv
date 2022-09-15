import React, { useState } from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import {VscLock, VscTrash} from 'react-icons/vsc'
import Modal  from '../../../components/modal';
import {useDropzone} from 'react-dropzone'
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';
import {useSecrets} from 'direktiv-react-hooks'
import {Config, GenerateRandomKey} from '../../../util'
import HelpIcon from '../../../components/help';
import Tabs from '../../../components/tabs'
import DirektivEditor from '../../../components/editor';
import { AutoSizer } from 'react-virtualized';

import { VscAdd } from 'react-icons/vsc';



function SecretsPanel(props){
    const {namespace} = props

    const [keyValue, setKeyValue] = useState("")
    const [file, setFile] = useState(null)
    const [vValue, setVValue] = useState("")
    const {data, createSecret, deleteSecret, getSecrets} = useSecrets(Config.url, namespace, localStorage.getItem("apikey"))

    return (
        <ContentPanel style={{ height: "100%", minHeight: "180px", width: "100%" }}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscLock />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} gap>
                    <div>
                        Secrets
                    </div>
                    <HelpIcon msg={"Encrypted key/value pairs that can be referenced within workflows. Suitable for storing sensitive information (such as tokens) for use in workflows."} />
                </FlexBox>
                <div>
                    <Modal title="New secret" 
                        escapeToCancel
                        titleIcon={<VscLock/>}
                        modalStyle={{width: "600px"}}

                        onOpen={() => {
                        }}

                        onClose={()=>{
                            setKeyValue("")
                            setVValue("")
                            setFile(null)
                        }}
                        
                        button={(
                            <VscAdd/>
                        )}
                        buttonProps={{
                            auto: true,
                        }}
                        actionButtons={[
                            {
                                label: "Add",

                                onClick: async () => {
                                    if(document.getElementById("file-picker")){
                                        if(keyValue.trim() === "") {
                                            throw new Error("Secret key name needs to be provided.")
                                        }
                                        if(!file) {
                                            throw new Error("Please add or select file")
                                        }
                                        await createSecret(keyValue, file)
                                    } else {
                                        if(keyValue.trim() === "") {
                                            throw new Error("Secret key name needs to be provided.")
                                        }
                                        if(vValue.trim() === "") {
                                            throw new Error("Secret value needs to be provided.")
                                        }
                                        await createSecret(keyValue, vValue)
                                    }
                                    await  getSecrets()
                                },

                                buttonProps: {variant: "contained", color: "primary"},
                                errFunc: ()=>{},
                                closesModal: true,
                                validate: true
                            },
                            {
                                label: "Cancel",

                                onClick: () => {
                                },

                                buttonProps: {},
                                errFunc: ()=>{},
                                closesModal: true
                            }
                        ]}
                        requiredFields={[
                            { tip: "secret key is required", value: keyValue }
                        ]}
                    >
                         <Tabs
            style={{minHeight: "100px", minWidth: "400px"}}
            headers={["Manual", "Upload"]}
            tabs={[(     
                <AddSecretPanel keyValue={keyValue} vValue={vValue} setKeyValue={setKeyValue} setVValue={setVValue} />
            ),(
                <FlexBox id="file-picker" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                    <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter key" />
                    </div>
                    <FlexBox id="file-picker" gap>
                        <SecretFilePicker file={file} setFile={setFile} id="add-secret-panel"/>
                    </FlexBox>
                </FlexBox>
            )]}
        />
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox col gap>
                    <FlexBox className="secrets-list"> 
                    {data !== null ? 
                        <Secrets deleteSecret={deleteSecret} getSecrets={getSecrets} secrets={data}  />: ""}
                    </FlexBox>
                    <div>
                        <Alert severity="info">Once a secret is removed, it can never be restored.</Alert>
                    </div>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    );
}

export default SecretsPanel;

export function SecretFilePicker(props) {
    const {file, setFile, id} = props

    const onDrop = acceptedFiles => {
        setFile(acceptedFiles[0])
    }
    
    const {getRootProps, getInputProps} = useDropzone({onDrop, multiple: false})

    return (
        <div {...getRootProps()} className="file-input" id={id} style={{display:"flex", flex:"auto", flexDirection:"column"}} >
            <div>
                <input {...getInputProps()} />
                <p>Drag 'n' drop the file here, or click to select file</p>
                {
                    file !== null ?
                    <p style={{margin:"0px"}}>Selected file: '{file.path}'</p>
                    :
                    ""
                }
            </div>
        </div>
    )
}

function Secrets(props) {
    const {secrets, deleteSecret, getSecrets} = props

    return <>
        <FlexBox col gap style={{ maxHeight: "236px", overflowY: "auto" }}>
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
                            <FlexBox className="key">{obj.name}</FlexBox>
                            <FlexBox className="val"><span>******</span></FlexBox>
                            <FlexBox className="actions">
                                <Modal 
                                    modalStyle={{width: "360px"}}
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
                                    buttonProps={{
                                        variant: "text",
                                        color: "info"
                                    }} 
                                    actionButtons={
                                        [
                                            {
                                                label: "Delete",

                                                onClick: async () => {
                                                    await deleteSecret(obj.name)
                                                    await getSecrets()
                                                },

                                                buttonProps: {variant: "contained", color: "error"},
                                                errFunc: ()=>{},
                                                closesModal: true
                                            },
                                            {
                                                label: "Cancel",

                                                onClick: () => {
                                                },

                                                buttonProps: {},
                                                errFunc: ()=>{},
                                                closesModal: true
                                            }
                                        ]
                                    }   
                                >
                                    <FlexBox col gap>
                                        <FlexBox >
                                            Are you sure you want to delete '{obj.name}'?
                                            <br/>
                                            This action cannot be undone.
                                        </FlexBox>
                                    </FlexBox>
                                </Modal>
                            </FlexBox>
                        </FlexBox>
                    );
                })}</>}
        </FlexBox>
    </>;
}

export function SecretsDeleteButton(props) {
    return (
        <div className="red-text" style={{display: "flex", alignItems: "center", height: "100%"}}>
            <VscTrash />
        </div>
    )
}

function AddSecretPanel(props) {
    const {keyValue, vValue, setKeyValue, setVValue} = props

    return (
        <FlexBox col gap style={{fontSize: "12px", width: "400px"}}>
            <FlexBox gap>
                <FlexBox>
                    <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter key" />
                </FlexBox>
            </FlexBox>
            <FlexBox gap style={{minHeight:"250px"}}>
                <FlexBox style={{overflow:"hidden"}}>
                    <AutoSizer>
                        {({height, width})=>(
                            <DirektivEditor width={width} dvalue={vValue} setDValue={setVValue} height={height}/>
                        )}
                        </AutoSizer>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
    );
}