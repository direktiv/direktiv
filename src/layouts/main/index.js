import React, { useEffect, useState } from 'react';
import './style.css';
import Breadcrumbs from '../../components/breadcrumbs';
import Settings from '../settings';
import Explorer from '../explorer';
import FlexBox from '../../components/flexbox';
import NavBar from '../../components/navbar';

import { useNamespaces } from 'direktiv-react-hooks' 
import { Config } from '../../util'
import { BrowserRouter, Routes, Route, useNavigate} from 'react-router-dom'
import InstancesPage from '../instances';
import JQPlayground from '../jqplayground';
import GlobalRegistriesPanel from '../global-registries';
import NamespaceServices from '../namespace-services';
import NamespaceRevisions from '../namespace-services/revisions';
import PodPanel from '../namespace-services/pod';
import GlobalServicesPanel from '../global-services';
import GlobalRevisionsPanel from '../global-services/revisions';
import GlobalPodPanel from '../global-services/pod'


function NamespaceNavigation(props){
    const {namespaces, namespace, setNamespace, deleteNamespace, deleteErr} = props

    const [load, setLoad] = useState(true)
    const navigate = useNavigate()

    // on mount check if namespace is stored in local storage and exists in the response given back
    useEffect(()=>{

        // only do this check if its not provided in the params
        if (namespaces !== null && namespaces.length > 0) {
            let urlpath = window.location.pathname.split("/")
            let ns = localStorage.getItem('namespace')
            if(urlpath[1] && urlpath[1] === "n"){
                // urlpath[2] would be the namespace 
                ns = urlpath[2]
            } 
            if (ns) {
                let found = false
                for(let i=0; i < namespaces.length; i++) {
                    if(namespaces[i].node.name === ns){
                        found = true
                        break
                    }
                }
                if (!found) {
                    // not found set it to the index page
                    setNamespace("")
                    setLoad(false)
                    localStorage.setItem('namespace', "")
                    navigate("/", {replace: true})
                    return
                } 
            } else {
                // locally stored namespace didn't exist in array so choose the 1st element
                ns = namespaces[0].node.name   
            }
            // namespace is good and found go to this one
            localStorage.setItem('namespace', ns)
            setNamespace(ns)
            setLoad(false)
            if(window.location.pathname === "/") {
                navigate(`/n/${ns}`, {replace: true})
            }
        } else  {
            // no namespaces should we should reset namespace back to ""
            if(!load) {
                setNamespace("")
            }
            setLoad(false)
        }

        if(namespaces !== null && namespaces.length === 0 && window.location.pathname !== "/") {
            navigate("/", {replace: true})
        }
    },[namespaces, navigate, setNamespace, namespace, load])

    if(load) {
        return ""
    }

    return(
        <FlexBox className="content-col col">
            <FlexBox className="breadcrumbs-row">
                <Breadcrumbs namespace={namespace}/>
            </FlexBox>
            <FlexBox className="col" style={{paddingBottom: "8px"}}>
                {namespaces !== null ? 
                <Routes>
                    <Route path="/" element={<div>index route:)</div>} />
                    {/* Explorer routing */}
                    <Route path="/n/:namespace" element={<Explorer namespace={namespace} />} />
                    <Route path="/n/:namespace/:type/*" element={<Explorer namespace={namespace} />} />

                    <Route path="/n/:namespace/monitoring" element={<div>monitor</div>}/>
                    {/* <Route path="/n/:namespace/builder" element={<WorkflowBuilder namespace={namespace}/>}/> */}
                    <Route path="/n/:namespace/instances" element={<InstancesPage namespace={namespace} />}/>
                    <Route path="/n/:namespace/instances/:id" element={<div>instance id</div>} />
                    <Route path="/n/:namespace/permissions" element={<div>permissions</div>} />
                    <Route path="/n/:namespace/services" element={<NamespaceServices namespace={namespace}/>}/>
                    <Route path="/n/:namespace/services/:service" element={<NamespaceRevisions namespace={namespace}/>}/>
                    <Route path="/n/:namespace/services/:service/:revision" element={<PodPanel namespace={namespace}/>}/>
                    <Route path="/n/:namespace/settings" element={<Settings deleteErr={deleteErr} namespace={namespace} deleteNamespace={deleteNamespace}/>} />
                    <Route path="/n/:namespace/events" element={<div>events</div>}/>

                    {/* non-namespace routes */}
                    <Route path="/jq" element={<JQPlayground />} />

                    <Route path="/g/services" element={<GlobalServicesPanel/>} />
                    <Route path="/g/services/:service" element={<GlobalRevisionsPanel/>} />
                    <Route path="/g/services/:service/:revision" element={<GlobalPodPanel/>} />

                    <Route path="/g/registries" element={<GlobalRegistriesPanel />} />
                </Routes>:""}
            </FlexBox>
        </FlexBox>
    )
}

function MainLayout(props) {
    let {onClick, style, className} = props;

    const { data, err, createErr, deleteErr, createNamespace, deleteNamespace } = useNamespaces(Config.url, true)
    const [namespace, setNamespace] = useState(null)


    // TODO work out how to handle this error when listing namespaces
    if(err !== null) {
        // createNamespace, deleteNamespace or listing namespaces has an error
        console.log(err)
    }
    // if (data === null) {
    //     // still loading :)
    //     return(
    //         <div>we loading</div>
    //     )
    // }

    return(
        <div id="main-layout" onClick={onClick} style={style} className={className}>
            <FlexBox className="row gap tall" style={{minHeight: "100vh"}}>
                {/* 
                    Left col: navigation
                    Right : page contents 
                */}

                <BrowserRouter>
                    <FlexBox className="navigation-col">
                        <NavBar setNamespace={setNamespace} namespace={namespace} createErr={createErr} createNamespace={createNamespace} deleteNamespace={deleteNamespace} namespaces={data} />
                    </FlexBox>
                    <NamespaceNavigation deleteErr={deleteErr} deleteNamespace={deleteNamespace} namespace={namespace} setNamespace={setNamespace} namespaces={data}/>
                </BrowserRouter>

            </FlexBox>
        </div>
    );
}

export default MainLayout;