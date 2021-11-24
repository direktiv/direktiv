import React, { useCallback, useState } from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoCloudDownloadSharp, IoCloudUpload, IoCloudUploadSharp, IoLockClosedOutline } from 'react-icons/io5';
import FlexBox from '../../../components/flexbox';
import Modal, { ButtonDefinition } from '../../../components/modal';
import AddValueButton from '../../../components/add-button';
import { useNamespaceVariables } from 'direktiv-react-hooks';
import { Config } from '../../../util';
import DirektivEditor from '../../../components/editor';
import Button from '../../../components/button';
import {useDropzone} from 'react-dropzone'
import { SecretsDeleteButton } from '../secrets-panel';
import Tabs from '../../../components/tabs';

function VariablesPanel(props){

    const {namespace} = props
    const [keyValue, setKeyValue] = useState("")
    const [dValue, setDValue] = useState("")
    const [file, setFile] = useState(null)

    const {data, err, setNamespaceVariable, getNamespaceVariable, deleteNamespaceVariable} = useNamespaceVariables(Config.url, true, namespace)

    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Variables   
                </FlexBox>
                <div>
                    <Modal title="New variable" 
                        escapeToCancel
                        button={(
                            <AddValueButton label=" " />
                        )}  
                        onClose={()=>{
                            setKeyValue("")
                            setDValue("")
                            setFile(null)
                        }}
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                if(document.getElementById("file-picker")){
                                    let err = await setNamespaceVariable(keyValue, file)
                                    if (err) return err
                                } else {
                                    let err = await setNamespaceVariable(keyValue, dValue)
                                    if (err) return err
                                }
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", true, false)
                        ]}
                    >
                        <AddVariablePanel file={file} setFile={setFile} setKeyValue={setKeyValue} keyValue={keyValue} dValue={dValue} setDValue={setDValue}/>
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody style={{minHeight:"180px"}}>
                {data !== null ?
                <Variables deleteNamespaceVariable={deleteNamespaceVariable} setNamespaceVariable={setNamespaceVariable} getNamespaceVariable={getNamespaceVariable} variables={data}/>:""}
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default VariablesPanel;


function VariableFilePicker(props) {
    const {file, setFile} = props

    const onDrop = useCallback(acceptedFiles => {
        setFile(acceptedFiles[0])
    },[])
    
    const {getRootProps, getInputProps} = useDropzone({onDrop, multiple: false})

    return (
        <FlexBox className="file-input" style={{flexDirection:"column"}} {...getRootProps()}>
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
        </FlexBox>
    )
}

function AddVariablePanel(props) {
    const {keyValue, setKeyValue, dValue, setDValue, file, setFile} = props

    return(
        <Tabs 
            style={{minHeight: "400px", minWidth: "450px"}}
            headers={["Manual", "Upload"]}
            tabs={[(
                <FlexBox id="written" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <FlexBox className="gap" style={{maxHeight: "300px"}}>
                        <FlexBox style={{overflow:"hidden"}}>
                            <DirektivEditor dlang={"shell"} width={"450px"} dvalue={dValue} setDValue={setDValue} height={"300px"}/>
                        </FlexBox>
                    </FlexBox>
                </FlexBox>
            ),(
                <FlexBox id="file-picker" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <FlexBox className="gap">
                        <VariableFilePicker file={file} setFile={setFile} />
                    </FlexBox>
                </FlexBox>
            )]}
        />
    )
}

function Variables(props) {
    const {variables, getNamespaceVariable, setNamespaceVariable, deleteNamespaceVariable} = props

    const [val, setValue] = useState("")
    const [mimeType, setType] = useState("")
    const [file, setFile] = useState(null)

    const onDrop = useCallback(acceptedFiles => {
        setFile(acceptedFiles[0])
    },[])
    
    const {getRootProps, getInputProps} = useDropzone({onDrop, multiple: false})


    return(
        <FlexBox>
            {variables.length === 0  ? "":
            <table className="variables-table">
                <thead>
                    <tr className="header-row">
                        <th className="left-align" style={{ width: "180px", maxWidth: "180px" }}>
                            Name
                        </th>
                        <th>
                            Value
                        </th>
                        <th className="left-align" style={{ width: "80px", maxWidth: "80px" }}>
                            Size
                        </th>
                        <th className="center-align" style={{ width: "120px", maxWidth: "120px" }}>
                            Action
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {variables.map((obj)=>{
                        return(
                            <tr key={`${obj.node.name}${obj.node.size}`}>
                                <td>{obj.node.name}</td>
                                <td className="muted-text">
                                    {obj.node.size <= 2500000 ? 
                                        <Modal
                                            escapeToCancel
                                            style={{
                                                flexDirection: "row-reverse",
                                                marginRight: "8px"
                                            }}
                                            title="View Variable" 
                                            onClose={()=>{
                                                setType("")
                                                setValue("")
                                            }}
                                            onOpen={async ()=>{
                                                let data = await getNamespaceVariable(obj.node.name)
                                                setType(data.contentType)
                                                setValue(data.data)
                                            }}
                                            button={(
                                                <Button className="small">Show value</Button>
                                            )}
                                            actionButtons={
                                                [
                                                    ButtonDefinition("Edit", async () => {
                                                        let err = await setNamespaceVariable(obj.node.name, val , mimeType)
                                                        if (err) return err
                                                    }, "small blue", true, false),
                                                    ButtonDefinition("Cancel", () => {
                                                    }, "small light", true, false)
                                                ]
                                            } 
                                        >
                                            <FlexBox className="col gap" style={{fontSize: "12px"}}>
                                                <FlexBox className="gap">
                                                    <FlexBox style={{overflow:"hidden"}}>
                                                        <DirektivEditor dlang={"shell"} width={"450px"} dvalue={val} setDValue={setValue} height={"300px"}/>
                                                    </FlexBox>
                                                </FlexBox>
                                                <FlexBox className="gap">
                                                    <FlexBox>
                                                        <input value={mimeType} onChange={(e)=>setType(e.target.value)} placeholder="Enter mimetype for variable" />
                                                    </FlexBox>
                                                </FlexBox>
                                            </FlexBox>
                                        </Modal>:
                                        "Cannot show filesize greater than 2.5MiB"
                                        }
                                </td>
                                <td>{fileSize(obj.node.size)}</td>
                                <td>
                                    <FlexBox style={{justifyContent:"center"}} >
                                        <FlexBox>
                                            <VariablesDownloadButton />
                                        </FlexBox>
                                        <Modal
                                            escapeToCancel
                                            style={{
                                                flexDirection: "row-reverse",
                                            }}
                                            onClose={()=>{
                                                setFile(null)
                                            }}
                                            title="Upload to a variable" 
                                            button={(
                                                <VariablesUploadButton />
                                            )}
                                            actionButtons={
                                                [
                                                    ButtonDefinition("Upload", async () => {
                                                        let err = await setNamespaceVariable(obj.node.name, file, mimeType)
                                                        if (err) return err
                                                    }, "small blue", true, false),
                                                    ButtonDefinition("Cancel", () => {
                                                    }, "small light", true, false)
                                                ]
                                            } 
                                        >
                                            <FlexBox className="col gap">
                                                <VariableFilePicker file={file} setFile={setFile} />
                                            </FlexBox>
                                        </Modal>
                                        <Modal
                                            escapeToCancel
                                            style={{
                                                flexDirection: "row-reverse",
                                            }}
                                            title="Delete a variable" 
                                            button={(
                                                <SecretsDeleteButton/>
                                            )}
                                            actionButtons={
                                                [
                                                    ButtonDefinition("Delete", async () => {
                                                        let err = await deleteNamespaceVariable(obj.node.name)
                                                        if (err) return err
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
                                </td>
                            </tr>
                        )
                    })}
                </tbody>
            </table>}
        </FlexBox>
    );
}

function VariablesUploadButton() {
    return (
        <div className="secrets-delete-btn grey-text">
            <IoCloudUploadSharp/>
        </div>
    )
}
function VariablesDownloadButton() {
    return (
        <div className="secrets-delete-btn grey-text">
            <IoCloudDownloadSharp/>
        </div>
    )
}

function fileSize(size) {
    var i = Math.floor(Math.log(size) / Math.log(1024));
    return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
}