import React, { useState } from 'react';
import './style.css';
import { VscAdd, VscLock, VscTrash } from 'react-icons/vsc';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import DirektivEditor from '../../components/editor';
import FlexBox from '../../components/flexbox';
import Modal, { ButtonDefinition } from '../../components/modal';
import { GenerateRandomKey } from '../../util';
import Select from 'react-select';

function PermissionsPageWrapper(props) {

    let {namespace} = props;
    if (!namespace) {
        return <></>
    }

    return(<PermissionsPage />)
}

export default PermissionsPageWrapper;

function PermissionsPage(props) {

    return (<>
        <FlexBox className="wrap gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="wrap gap col">
                <FlexBox style={{ maxHeight: "160px" }}>
                    <PermissionsPanel />
                </FlexBox>
                <FlexBox>
                    <GroupPoliciesPanel />
                </FlexBox>
            </FlexBox>
            <FlexBox style={{ minWidth: "48%" }}>
                <OPAEditorPanel />
            </FlexBox>
        </FlexBox>
    </>)
}

function PermissionsPanel(props) {

    let dummyData = ["listNodes", "createNodes", "deleteNodes", "executeWorkflows", "updateWorkflows"]

    let policyList = []
    for (let i = 0; i < dummyData.length; i++) {
        policyList.push(
            <FlexBox className="gap group-policy-list-item" style={{ cursor: "unset" }} >
                <FlexBox>
                    {dummyData[i]}
                </FlexBox>
                <div>
                    <input type="checkbox" />
                </div>
            </FlexBox>
        )
    }

    return (
        <>
            <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscLock/>
                    </ContentPanelTitleIcon>
                    <div>
                        Permissions
                    </div>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <FlexBox className="col">
                        <FlexBox className="subtle-border-bottom">
                            <FlexBox style={{justifyContent: "left", alignItems: "center"}}>
                                <div style={{maxWidth: "100%", wordWrap: "break-word"}}>
                                    <b>Set group policies</b><br/>
                                    <span style={{fontSize: "12px"}}>Existing policies can be seen in the 'Open Policy Agent' editor.</span>
                                </div>
                            </FlexBox>
                            <FlexBox style={{justifyContent: "right", alignItems: "center", maxWidth: "45px"}}>
                                <Modal
                                    title="Set group policies"
                                    style={{
                                        justifyContent: "right",
                                        alignItems: "right"
                                    }}
                                    button={(
                                        <Button className="small light">
                                            <VscAdd style={{marginTop: "2px"}} />
                                        </Button>
                                    )}
                                    actionButtons = {[
                                        // label, onClick, classList, closesModal, async
                                        ButtonDefinition("Save", () => {
                                        }, "small", true, false),
                                        ButtonDefinition("Cancel", () => {
                                        }, "small light", true, false)
                                    ]}
                                >
                                    <FlexBox className="col gap" style={{maxWidth: "400px"}}>
                                        <div className="center-align subtle-border-bottom" style={{paddingBottom: "8px"}}>
                                            Edit group policies.
                                            Existing policies can be seen in the 'Open Policy Agent' editor.
                                        </div>
                                        <div className="center-align" style={{ fontSize: "12px" }}>
                                            <b className="red-text">IMPORTANT: </b> <br/>This will overwrite any existing policy assignments to the targeted group.
                                        </div>
                                        <div style={{paddingRight: "8px"}}>
                                            <input placeholder="Group name" type="" />
                                        </div>
                                        <div style={{paddingRight: "8px"}}>
                                            <input placeholder="Policies" type="" />
                                        </div>
                                    </FlexBox>
                                </Modal>
                            </FlexBox>
                        </FlexBox>
                        <FlexBox style={{ justifyContent: "space-between", alignItems: "center" }}>
                            <div style={{maxWidth: "100%", wordWrap: "break-word"}}>
                                <b>Create a new access token</b><br/>
                                <span style={{fontSize: "12px"}}>Access tokens are used to authenticate/authorize API requests.</span>
                            </div>
                            <FlexBox style={{justifyContent: "right", alignItems: "center", maxWidth: "45px"}}>
                                <Modal
                                    title="Create Access Token"
                                    style={{
                                        justifyContent: "right",
                                        alignItems: "right"
                                    }}
                                    button={(
                                        <Button className="small light">
                                            <VscAdd style={{marginTop: "2px"}} />
                                        </Button>
                                    )}
                                    actionButtons = {[
                                        // label, onClick, classList, closesModal, async
                                        ButtonDefinition("Create", () => {
                                            console.log("create auth token");
                                        }, "small", true, false),
                                        ButtonDefinition("Cancel", () => {
                                        }, "small light", true, false)
                                    ]}
                                >
                                    <FlexBox className="col gap" style={{maxWidth: "400px"}}>
                                        <div className="center-align">
                                            Create a new Access Token, which can be used to 
                                            authenticate API requests sent to the Direktiv server.
                                        </div>
                                        <div style={{paddingRight: "8px"}}>
                                            <input placeholder="Lifetime (seconds)" type="" />
                                        </div>
                                        <FlexBox className="gap col" style={{ maxHeight: "300px" }}>
                                            {policyList}
                                        </FlexBox>
                                    </FlexBox>
                                </Modal>
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </>
    )
}

function GroupPoliciesPanel(props) {
    return (
        <>
            <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscLock/>
                    </ContentPanelTitleIcon>
                    <div>
                        Group Policies
                    </div>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <GroupPoliciesList />
                </ContentPanelBody>
            </ContentPanel>
        </>
    )
}

function GroupPoliciesList(props) {

    let dummyData = [{
        group: "jon",
        policies: ["do-this", "do-that", "do-the-other"]
    }, {
        group: "jacob",
        policies: ["jingleheimer", "schmidt"]
    }, {
        group: "superman",
        policies: ["kryptonian"]
    }]

    return(
        <FlexBox className="col gap">
            {dummyData.map((obj) => {
                return (
                    <GroupPoliciesListItem key={GenerateRandomKey("")} obj={obj} />  
                )
            })}
        </FlexBox>
    )
}

function GroupPoliciesListItem(props) {

    let {obj} = props;

    return (
        <Modal
            activeOverlay
            escapeToCancel
            title={`Policies for '${obj.group}'`}
            style={{
                maxHeight: "32px"
            }}
            button={(
                <FlexBox className="gap group-policy-list-item">
                    <div>
                        <b>{obj.group}</b>
                    </div>
                    <FlexBox style={{justifyContent: "right"}}>
                        <span className="group-policy-list-item-count">{`${obj.policies.length} polic`}{obj.policies.length > 1 ? "ies" : "y"}</span>
                    </FlexBox>
                </FlexBox>
            )}
            actionButtons = {[
                // label, onClick, classList, closesModal, async
                ButtonDefinition("Save", () => {
                }, "small", true, false),
                ButtonDefinition("Cancel", () => {
                }, "small light", true, false)
            ]}
        >
            <FlexBox className="col gap" style={{maxHeight: "80vh", overflowY: "auto"}}>
                <FlexBox className="gap" style={{ alignItems: "center" }}>
                    <FlexBox style={{ padding: "1px" }}>
                        <input placeholder="Add group policy" type="text" />
                    </FlexBox>
                    <div style={{ marginRight: "2px" }}>
                        <Button className="small light">
                            <VscAdd className="green-text" style={{ marginTop: "2px" }} />
                        </Button>
                    </div>
                </FlexBox>
                {obj.policies.map((item) => {
                    return (<FlexBox className="policy-name" key={GenerateRandomKey} style={{ alignItems: "center" }}>
                        <div>
                            {item}
                        </div>
                        <FlexBox style={{ justifyContent: "right" }}>
                            <Button className="small light">
                                <VscTrash className="red-text" />
                            </Button>
                        </FlexBox>
                    </FlexBox>)
                })}
            </FlexBox>
        </Modal>
    )
}

const dummyDataOPA = `package direktiv.authz

is_in_group[g] {
  some i
  group := input.groups[i]
  g = data.groups[_]
  g.name = group
}
  
authorizeAPI {
  some group
  is_in_group[group]
  bits.and(group.perm, input.action) != 0
}

namespaceOwner {
  some group
  is_in_group[group]
  bits.and(group.perm, input.namespaceOwner) != 0
}

viewNamespace {
  authorizeAPI
}
 
listServices {
  authorizeAPI
}
 
deleteNamespace {
  authorizeAPI
}
 
getNode {
  authorizeAPI
}
 
mkdir {
  authorizeAPI
}
 
createWorkflow {
  authorizeAPI
}
 
updateWorkflow {
  authorizeAPI
}
 
deleteWorkflow {
  authorizeAPI
}
 
deleteNode {
  authorizeAPI
}
 
getWorkflow {
  authorizeAPI
}
 
deleteRevision {
  authorizeAPI
}
 
tag {
  authorizeAPI
}
 
untag {
  authorizeAPI
}
 
validateRef {
  authorizeAPI
}
 
validateRouter {
  authorizeAPI
}
 
sendNamespaceEvent {
  authorizeAPI
}
 
listSecrets {
  authorizeAPI
}
 
createSecret {
  authorizeAPI
}
 
deleteSecret {
  authorizeAPI
}
 
listRegistries {
  authorizeAPI
}
 
createRegistry {
  authorizeAPI
}
 
deleteRegistry {
  authorizeAPI
}
 
executeWorkflow {
  authorizeAPI
}
 
listInstances {
  authorizeAPI
}
 
getInstance {
  authorizeAPI
}
 
cancelInstance {
  authorizeAPI
}
 
createAttribute {
  authorizeAPI
}
 
deleteAttribute {
  authorizeAPI
}
 
listInstanceVariables {
  authorizeAPI
}
 
getInstanceVariable {
  authorizeAPI
}
 
setInstanceVariable {
  authorizeAPI
}
 
listWorkflowVariables {
  authorizeAPI
}
 
getWorkflowVariable {
  authorizeAPI
}
 
setWorkflowVariable {
  authorizeAPI
}
 
listNamespaceVariables {
  authorizeAPI
}
 
getNamespaceVariable {
  authorizeAPI
}
 
setNamespaceVariable {
  authorizeAPI
}
 
getNamespaceLogs {
  authorizeAPI
}
 
getWorkflowLogs {
  authorizeAPI
}
 
getInstanceLogs {
  authorizeAPI
}
 
watchPods {
  authorizeAPI
}
 
watchLogs {
  authorizeAPI
}
 
listPods {
  authorizeAPI
}
 
deleteService {
  authorizeAPI
}
 
getService {
  authorizeAPI
}
 
createService {
  authorizeAPI
}
 
updateService {
  authorizeAPI
}

getGroups {
  namespaceOwner
  authorizeAPI
}

getGroupPermissions {
  authorizeAPI
}

setGroupPermissions {
  namespaceOwner
  authorizeAPI
}

getPoliciesFile {
  namespaceOwner
  authorizeAPI
}

editPoliciesFile {
  namespaceOwner
  authorizeAPI
}

createAuthToken {
  namespaceOwner
  authorizeAPI
}

getMetrics {
  authorizeAPI
}

listWorkflowServices {
  authorizeAPI
}
`

const opaRegExp = /[\w\[\]]*[\s]{[\s\w:=.\[\](),!]*}/gm;
const policyNameRegExp = /[\w_\[\]]*/

function splitOPAData(data) {

  let policies = data.matchAll(opaRegExp);
  let arr = Array.from(policies, m => m[0]);
  let policyMap = {};

  for (let i = 0; i < arr.length; i++) {
    let policyName = arr[i].match(policyNameRegExp)[0];
    policyMap[policyName] = arr[i];
  }

  return policyMap;
}

function OPAEditorPanel(props) {

  const [selectedOption, setSelectedOption] = useState(null)
  const [policyData, setPolicyData] = useState(null)

  let policyMap = splitOPAData(dummyDataOPA);
  let opts = [];

  for (const property in policyMap) {
    opts.push({
      value: property,
      label: property
    })
  }

  let saveOPA = function() {
    if (policyData) {
      let newOPAData = "package direktiv.authz\n\n";
      for (const property in policyMap) {
        if (property == selectedOption.value) {
          newOPAData += policyData + "\n\n"
        } else {
          newOPAData += policyMap[property] + "\n\n"
        }
      }

      // TODO: When we can start implementing EE functionality,
      // the 'newOPAData' value contains the full payload to send
      // to the Direktiv backend!
      console.log(newOPAData);
    }
  }

  let setOption = function(val) {
    setSelectedOption(val)
    setPolicyData(policyMap[val])
  }

  return (
    <>
      <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
        <ContentPanelTitle>
          <ContentPanelTitleIcon>
            <VscLock/>
          </ContentPanelTitleIcon>
          <div>
            Editor - Open Policy Agent
          </div>
        </ContentPanelTitle>
        <ContentPanelBody style={{overflow: "hidden"}}>
          <FlexBox className="col">
            <div style={{paddingBottom: "8px"}}>
              <Select 
                defaultValue={selectedOption}
                onChange={setOption}
                options={opts}
              />
            </div>
            <FlexBox>
              { selectedOption ?
              <FlexBox>
                <DirektivEditor width="300" dlang="css" setDValue={setPolicyData} dvalue={policyMap[selectedOption.value]} style={{
                  borderBottomRightRadius: "0px",
                  borderBottomLeftRadius: "0px"
                }} />
              </FlexBox>
              :<DirektivEditor width="300" value="" readonly={true} />}
            </FlexBox>
            { selectedOption ?
            <FlexBox className="gap" style={{backgroundColor:"#223848", color:"white", height:"44px", maxHeight:"44px", paddingLeft:"10px", minHeight:"44px", alignItems:'center', borderRadius:"0px 0px 8px 8px", overflow: "hidden"}}>
              <div style={{display:"flex", flex:1 }}>
              </div>
              <div style={{display:"flex", flex:1, justifyContent:"center"}}>
              </div>
              <div style={{display:"flex", flex:1, gap :"3px", justifyContent:"flex-end", paddingRight:"10px"}}>
                <div onClick={saveOPA} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                    Save
                </div>
              </div>
            </FlexBox>
            :<></>}
          </FlexBox>
        </ContentPanelBody>
      </ContentPanel>
    </>
  )
}