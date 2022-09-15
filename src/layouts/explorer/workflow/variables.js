import { useWorkflowVariables } from 'direktiv-react-hooks';
import { useEffect, useState } from 'react';

import { VscCloudDownload, VscCloudUpload, VscEye, VscLoading, VscTrash, VscVariableGroup, VscAdd } from 'react-icons/vsc';

import Tippy from '@tippyjs/react';
import { saveAs } from 'file-saver';
import { AutoSizer } from 'react-virtualized';
import { SearchBar } from '..';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import DirektivEditor from '../../../components/editor';
import FlexBox from '../../../components/flexbox';
import HelpIcon from "../../../components/help";
import Modal from '../../../components/modal';
import Pagination, { usePageHandler } from '../../../components/pagination';
import Tabs from '../../../components/tabs';
import { CanPreviewMimeType, Config, MimeTypeFileExtension } from '../../../util';
import { VariableFilePicker } from '../../settings/variables-panel';

import Button from '../../../components/button';


const PAGE_SIZE = 10 ;

function AddWorkflowVariablePanel(props) {

    const {namespace, workflow} = props
    const [keyValue, setKeyValue] = useState("")
    const [dValue, setDValue] = useState("")
    const [file, setFile] = useState(null)
    const [uploading, setUploading] = useState(false)
    const [mimeType, setMimeType] = useState("application/json")
    const [search, setSearch] = useState("")

    let wfVar = workflow.substring(1)

    const pageHandler = usePageHandler(PAGE_SIZE)
    const goToFirstPage = pageHandler.goToFirstPage
    const {data, pageInfo, setWorkflowVariable, getWorkflowVariable, getWorkflowVariableBlob, deleteWorkflowVariable} = useWorkflowVariables(Config.url, true, namespace, wfVar, localStorage.getItem("apikey"), pageHandler.pageParams, `filter.field=NAME`, `filter.val=${search}`, `filter.type=CONTAINS`)

    // Reset Page to start when filters changes
    useEffect(() => {
        // TODO: This will interfere with page position if initPage > 1
        goToFirstPage()
    }, [search, goToFirstPage])

    if (data === null) {
        return <></>
    }

    return (
        <ContentPanel style={{width: "100%", height: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscVariableGroup />
                </ContentPanelTitleIcon>
                <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
                    <div>
                        Variables
                    </div>
                    <HelpIcon msg={"List of variables for that workflow."} />
                </FlexBox>
                <SearchBar setSearch={setSearch} style={{ height: "26px" }} />
                <div>
                    <Modal title="New variable"
                        modalStyle={{ width: "600px" }}
                        style={{ maxWidth: "42px" }}
                        escapeToCancel
                        button={(
                            <VscAdd />
                        )}
                        buttonProps={{
                            auto: true
                        }}
                        onClose={() => {
                            setKeyValue("")
                            setDValue("")
                            setFile(null)
                            setUploading(false)
                            setMimeType("application/json")
                        }}
                        actionButtons={[
                            {
                                label: "Add",

                                onClick: async () => {
                                    if (document.getElementById("file-picker")) {
                                        setUploading(true)
                                        await setWorkflowVariable(encodeURIComponent(keyValue), file, mimeType)
                                    } else {
                                        await setWorkflowVariable(encodeURIComponent(keyValue), dValue, mimeType)
                                    }
                                },

                                buttonProps: {variant: "contained", color: "primary", loading: uploading},
                                errFunc: () => { setUploading(false) },
                                closesModal: true,
                                validate: true
                            },
                            {
                                label: "Cancel",

                                onClick: () => {
                                },

                                buttonProps: {},
                                errFunc: () => { },
                                closesModal: true
                            }
                        ]}

                        requiredFields={[
                            { tip: "variable key name is required", value: keyValue }
                        ]}
                    >
                        <AddVariablePanel mimeType={mimeType} setMimeType={setMimeType} file={file} setFile={setFile} setKeyValue={setKeyValue} keyValue={keyValue} dvalue={dValue} setDValue={setDValue} />
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody>
                {data !== null ?
                <FlexBox col>
                    <div>
                    <Variables namespace={namespace} deleteWorkflowVariable={deleteWorkflowVariable} setWorkflowVariable={setWorkflowVariable} getWorkflowVariable={getWorkflowVariable} getWorkflowVariableBlob={getWorkflowVariableBlob} variables={data}/>
                    </div>
                    <FlexBox row style={{justifyContent:"flex-end", paddingBottom:"1em", flexGrow: 0}}>
                        <Pagination pageHandler={pageHandler} pageInfo={pageInfo}/>
                    </FlexBox>
                </FlexBox>:<></>}
            </ContentPanelBody>
        </ContentPanel>
    );
}

export default AddWorkflowVariablePanel;

function AddVariablePanel(props) {
    const {keyValue, setKeyValue, dValue, setDValue, file, setFile, mimeType, setMimeType} = props

    let lang = ""

    switch(mimeType){
    case "application/json":
        lang = "json"
        break
    case "application/x-sh":
        lang = "shell"
        break
    case "text/html":
        lang = "html"
        break
    case "text/css":
        lang = "css"
        break
    case "application/yaml":
        lang = "yaml"
        break
    default:
        lang = "plain"
    }

    return(
        <Tabs
            style={{minHeight: "500px", minWidth: "90%"}}
            headers={["Manual", "Upload"]}
            tabs={[(
                <FlexBox id="written" className="col gap" style={{fontSize: "12px", minWidth: "300px"}}>
                    <div style={{width: "100%", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <div style={{width: "100%", display: "flex"}}>
                        <select style={{width:"100%"}} defaultValue={mimeType} onChange={(e)=>setMimeType(e.target.value)}>
                            <option value="">Choose a mimetype</option>
                            <option value="application/json">json</option>
                            <option value="application/yaml">yaml</option>
                            <option value="application/x-sh">shell</option>
                            <option value="text/plain">plaintext</option>
                            <option value="text/html">html</option>
                            <option value="text/css">css</option>
                        </select>
                    </div>
                    <FlexBox gap style={{maxHeight: "600px"}}>
                        <FlexBox style={{overflow:"hidden"}}>
                        <AutoSizer>
                            {({height, width})=>(
                            <DirektivEditor dlang={lang} width={width} dvalue={dValue} setDValue={setDValue} height={height}/>
                            )}
                        </AutoSizer>
                        </FlexBox>
                    </FlexBox>
                </FlexBox>
            ),(
                <FlexBox id="file-picker" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <FlexBox gap>
                        <VariableFilePicker setKeyValue={setKeyValue} setMimeType={setMimeType} mimeType={mimeType} file={file} setFile={setFile} id="add-variable-panel" />
                    </FlexBox>
                </FlexBox>
            )]}
        />
    )
}

function Variables(props) {

    const {variables, namespace, getWorkflowVariable, setWorkflowVariable, deleteWorkflowVariable, getWorkflowVariableBlob} = props;

    return(
        <FlexBox>
            {variables.length === 0  ? <div style={{paddingLeft:"10px", fontSize:"10pt"}}>No variables are stored...</div>:
            <table className="variables-table">
                <tbody>
                    {variables.map((obj)=>{
                        return(
                            <Variable namespace={namespace} obj={obj} getWorkflowVariableBlob={getWorkflowVariableBlob} getWorkflowVariable={getWorkflowVariable} deleteWorkflowVariable={deleteWorkflowVariable} setWorkflowVariable={setWorkflowVariable}/>
                        )
                    })}
                </tbody>
            </table>
            }
        </FlexBox>
    );
}

function Variable(props) {
    const {obj, getWorkflowVariable, setWorkflowVariable, deleteWorkflowVariable, getWorkflowVariableBlob} = props
    const [val, setValue] = useState("")
    const [mimeType, setType] = useState("")
    const [file, setFile] = useState(null)
    const [downloading, setDownloading] = useState(false)
    const [uploading, setUploading] = useState(false)

    let lang = MimeTypeFileExtension(mimeType)

    return (
        <tr className="body-row" key={`var-${obj.name}${obj.size}`}>
        <td className="wrap-word variable-name" style={{ width: "180px", maxWidth: "180px", textOverflow:"ellipsis",  overflow:"hidden" }}>
            <Tippy content={obj.name} trigger={'mouseenter focus'} zIndex={10}>
                <div className={"variable-name"} style={{width: "fit-content", maxWidth: "180px", textOverflow:"ellipsis",  overflow:"hidden"}}>
                    {obj.name}
                </div>
            </Tippy>
        </td>
        <td className="muted-text show-variable">
            <FlexBox className="center-x">
                {obj.size <= 2500000 ? 
                    <Modal
                        modalStyle={{height: "90vh",width: "600px"}}
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
                            let data = await getWorkflowVariable(obj.name)
                            setType(data.contentType)
                            setValue(data.data)
                        }}
                        button={(
                            <FlexBox className={"gap"} style={{fontWeight:"bold"}}>
                                <VscEye className="auto-margin" />
                                Show <span className="hide-600">value</span>
                            </FlexBox>
                        )}
                        buttonProps={{
                            color: "info",
                        }}
                        actionButtons={
                            [
                                {
                                    label: "Save",

                                    onClick: async () => {
                                        await setWorkflowVariable(obj.name, val , mimeType)
                                    },

                                    buttonProps: {variant: "contained", color: "primary"},
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
                        <FlexBox col gap style={{fontSize: "12px",minHeight: "500px"}}>
                            <FlexBox gap style={{flexGrow: 1}}>
                                <FlexBox style={{overflow:"hidden"}}>
                                {CanPreviewMimeType(mimeType) ?   
                                <AutoSizer>
                                    {({height, width})=>(
                                    <DirektivEditor dlang={lang} width={width} dvalue={val} setDValue={setValue} height={height}/>
                                    )}
                                </AutoSizer>
                                :
                                <div style={{width: "100%", display:"flex", justifyContent: "center", alignItems:"center"}}>
                                    <p style={{fontSize:"11pt"}}>
                                        Cannot preview variable with mime-type: {mimeType}
                                    </p>
                                </div>
                                }
                                </FlexBox>
                            </FlexBox>
                            <FlexBox gap style={{flexGrow: 0, flexShrink: 1}}>
                                <FlexBox>
                                    <select style={{width:"100%"}} defaultValue={mimeType} onChange={(e)=>setType(e.target.value)}>
                                        <option value="">Choose a mimetype</option>
                                        <option value="application/json">json</option>
                                        <option value="application/yaml">yaml</option>
                                        <option value="application/x-sh">shell</option>
                                        <option value="text/plain">plaintext</option>
                                        <option value="text/html">html</option>
                                        <option value="text/css">css</option>
                                    </select>
                                </FlexBox>
                            </FlexBox>
                        </FlexBox>
                    </Modal>:<div style={{textAlign:"center"}}>Cannot show filesize greater than 2.5MiB</div>
                }
            </FlexBox>
        </td>
        <td style={{ width: "80px", maxWidth: "80px", textAlign: "center" }}>{fileSize(obj.size)}</td>
        <td style={{ width: "120px", maxWidth: "120px", paddingLeft: "12px" }}> 
            <FlexBox style={{gap: "2px", justifyContent:"flex-end"}}>
                <div>
                    {!downloading? 
                    <VariablesDownloadButton onClick={async()=>{
                        setDownloading(true)

                        const variableData = await getWorkflowVariableBlob(obj.name)
                        const extension = MimeTypeFileExtension(variableData.contentType)
                        saveAs(variableData.data, obj.name + `${extension ? `.${extension}`: ""}`)

                        setDownloading(false)
                    }}/>:<VariablesDownloadingButton />}
                </div>
                <div>
                    <Modal
                        modalStyle={{width: "500px"}}
                        escapeToCancel
                        style={{
                            flexDirection: "row-reverse",
                        }}
                        onClose={()=>{
                            setFile(null)
                        }}
                        title="Replace variable" 
                        button={(
                            <VariablesUploadButton />
                        )}
                        buttonProps={{
                            auto: true,
                            variant: "text",
                            color: "info"
                        }}
                        actionButtons={
                            [
                                {
                                    label: "Upload",

                                    onClick: async () => {
                                        setUploading(true)
                                        await setWorkflowVariable(obj.name, file, mimeType)
                                    },

                                    buttonProps: {variant: "contained", color: "primary", loading:uploading},
                                    errFunc: ()=>{setUploading(false)},
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
                            ]
                        } 

                        requiredFields={[
                            {tip: "file is required", value: file}
                        ]}
                    >
                        <FlexBox col gap>
                            <VariableFilePicker setMimeType={setType} id="modal-file-picker" file={file} setFile={setFile} />
                        </FlexBox>
                    </Modal>
                </div>
                <div>

                
                <Modal
                    escapeToCancel
                    style={{
                        flexDirection: "row-reverse",
                    }}
                    modalStyle={{width: "400px"}}
                    title="Delete a variable" 
                    button={(
                        <VariablesDeleteButton/>
                    )}
                    buttonProps={{
                        auto: true,
                        variant: "text",
                        color: "info"
                    }}
                    actionButtons={
                        [
                            {
                                label: "Delete",

                                onClick: async () => {
                                        await deleteWorkflowVariable(obj.name)
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
                </div>
            </FlexBox>
        </td>
    </tr>
    );
}

function VariablesUploadButton() {
    return (
        <div className="secrets-delete-btn grey-text auto-margin" style={{display: "flex", alignItems: "center", height: "100%"}}>
            <VscCloudUpload className="auto-margin"/>
        </div>
    )
}

function VariablesDownloadButton(props) {
    const {onClick} = props

    return (
        <Button onClick={onClick} auto variant={"text"} color={"info"}>
            <VscCloudDownload/>
        </Button>
    )
}

function VariablesDownloadingButton(props) {

    return (
        <VscLoading style={{animation: "spin 2s linear infinite"}}/>
    )
}


function VariablesDeleteButton() {
    return (
        <div className="red-text" style={{display: "flex", alignItems: "center", height: "100%"}}>
            <VscTrash className="auto-margin" />
        </div>
    )
}

function fileSize(size) {
    if (size <= 0) {
        return "0 B"
    }
    var i = Math.floor(Math.log(size) / Math.log(1024));
    return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
}