import React, { useState } from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoLockClosedOutline } from 'react-icons/io5';
import FlexBox from '../../../components/flexbox';
import Modal, { ButtonDefinition } from '../../../components/modal';
import AddValueButton from '../../../components/add-button';
import { useNamespaceVariables } from 'direktiv-react-hooks';
import { Config } from '../../../util';
import DirektivEditor from '../../../components/editor';
import Button from '../../../components/button';

function VariablesPanel(props){

    const {namespace} = props
    const [keyValue, setKeyValue] = useState("")
    const [dValue, setDValue] = useState("")

    const {data, err, setNamespaceVariable, getNamespaceVariable, deleteNamespaceVariable} = useNamespaceVariables(Config.url, true, namespace)
    console.log(data, err, "VARIABLES NAMESPACE")

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
                        }}
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                let err = await setNamespaceVariable(keyValue, dValue)
                                if (err) return err
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", true, false)
                        ]}
                    >
                        <AddVariablePanel setKeyValue={setKeyValue} keyValue={keyValue} dValue={dValue} setDValue={setDValue}/>
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody >
                {data !== null ?
                <Variables variables={data}/>:""}
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default VariablesPanel;

function AddVariablePanel(props) {
    const {keyValue, setKeyValue, dValue, setDValue} = props
    return(
        <FlexBox className="col gap" style={{fontSize: "12px"}}>
            <FlexBox className="gap">
                <FlexBox>
                    <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox style={{overflow:"hidden"}}>
                    <DirektivEditor dlang={"shell"} width={"450px"} dvalue={dValue} setDValue={setDValue} height={"300px"}/>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )
}

function Variables(props) {
    const {variables} = props
    return(
        <FlexBox>
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
                            <tr>
                                <td>{obj.node.name}</td>
                                <td className="muted-text">
                                    <Modal
                                        escapeToCancel
                                        style={{
                                            flexDirection: "row-reverse",
                                            marginRight: "8px"
                                        }}
                                        title="View Variable" 
                                        button={(
                                            <Button className="small">Show value</Button>
                                            )}
                                    >
                                        
                                    </Modal>
                                </td>
                                <td>{fileSize(obj.node.size)}</td>
                                <td></td>
                            </tr>
                        )
                    })}
                </tbody>
            </table>
        </FlexBox>
    );
}

function fileSize(size) {
    var i = Math.floor(Math.log(size) / Math.log(1024));
    return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i];
}