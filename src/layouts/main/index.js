import React, { useEffect, useState } from 'react';
import './style.css';
import SmallLogo from '../../assets/small-logo.jpeg';
import Breadcrumbs from '../../components/breadcrumbs';
import Settings from '../settings';
import Explorer from '../explorer';
import NotFound from '../notfound';
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
import Loader from '../../components/loader';
import Button from '../../components/button';
import { IoMenu } from 'react-icons/io5';
import InstancePageWrapper from '../instance';
// import PermissionsPageWrapper from '../permissions';
import EventsPageWrapper from '../events';
import Monitoring from '../monitoring';
import MirrorPage from '../mirror';


function NamespaceNavigation(props){
    const {namespaces, namespace, setNamespace, deleteNamespace, deleteErr, extraRoutes} = props

    const [load, setLoad] = useState(true)
    const [breadcrumbChildren, setBreadcrumbChildren] = useState(null)
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
        return "xxx"
    }

    return(
        <FlexBox className="content-col col">
            <FlexBox className="breadcrumbs-row">
                <Breadcrumbs namespace={namespace} additionalChildren={breadcrumbChildren}/>
            </FlexBox>
            <FlexBox className="col" style={{paddingBottom: "8px"}}>
                {namespaces !== null ? 
                <Routes>
                    <Route path="/" element={
                        <div className="message-container">
                            <p className="no-workspace-message">
                                You are not a part of any namespaces! Create a namespace to continue using Direktiv.
                            </p>
                        </div>
                    }/>
                    {/* Explorer routing */}
                    <Route path="/n/:namespace" element={<Explorer namespace={namespace} setBreadcrumbChildren={setBreadcrumbChildren} />} >
                        <Route path="explorer/*" element={<Explorer namespace={namespace} setBreadcrumbChildren={setBreadcrumbChildren} />} />
                    </Route>

                    <Route path="/n/:namespace/mirror/*" element={<MirrorPage namespace={namespace} setBreadcrumbChildren={setBreadcrumbChildren}/>} />

                    <Route path="/n/:namespace/monitoring" element={<Monitoring namespace={namespace}/>}/>
                    {/* <Route path="/n/:namespace/builder" element={<WorkflowBuilder namespace={namespace}/>}/> */}
                    <Route path="/n/:namespace/instances" element={<InstancesPage namespace={namespace} />}/>
                    <Route path="/n/:namespace/instances/:id" element={<InstancePageWrapper namespace={namespace} />} />
                    {/* <Route path="/n/:namespace/permissions" element={<PermissionsPageWrapper namespace={namespace} />} /> */}
                    
               
                    {/* namespace services */}
                    <Route path="/n/:namespace/services" element={<NamespaceServices namespace={namespace}/>}/>
                    <Route path="/n/:namespace/services/:service" element={<NamespaceRevisions namespace={namespace}/>}/>
                    <Route path="/n/:namespace/services/:service/:revision" element={<PodPanel namespace={namespace}/>}/>
                    
                    
                    <Route path="/n/:namespace/settings" element={<Settings deleteErr={deleteErr} namespace={namespace} deleteNamespace={deleteNamespace}/>} />
                    <Route path="/n/:namespace/events" element={<EventsPageWrapper namespace={namespace} />}/>

                    {extraRoutes.map((obj)=>{
                        return(
                            <Route path={obj.route} key={obj.route} element={obj.element(namespace)} />
                        )
                    })}

                    {/* non-namespace routes */}
                    <Route path="/jq" element={<JQPlayground />} />

                    <Route path="/g/services" element={<GlobalServicesPanel/>} />
                    <Route path="/g/services/:service" element={<GlobalRevisionsPanel/>} />
                    <Route path="/g/services/:service/:revision" element={<GlobalPodPanel/>} />

                    <Route path="/g/registries" element={<GlobalRegistriesPanel />} />
                    <Route path='*' exact={true} element={<NotFound/>} />
                </Routes>:""}
            </FlexBox>
        </FlexBox>
    )
}


function MainLayout(props) {
    let {onClick, style, className, extraNavigation, extraRoutes, footer, akey, akeyReq} = props;

   
    const [load, setLoad] = useState(true)
 
    const [namespace, setNamespace] = useState(null)
    const [toggleResponsive, setToggleResponsive] = useState(false);
    const {data, err, createNamespace, createMirrorNamespace, deleteNamespace} = useNamespaces(Config.url, true, akey)

    // const [versions, setVersions] = useState(false)
    
    useEffect(()=>{
        if(data !== null) {
            setLoad(false)
        }
        if(err !== null) {
            setLoad(false)
        }
    },[data, err])

    // TODO work out how to handle this error when listing namespaces
    if(err !== null) {
        // createNamespace, deleteNamespace or listing namespaces has an error
        // console.log(err)
    }
    // if (data === null) {
    //     // still loading :)
    //     return(
    //         <div>we loading</div>
    //     )
    // }
    return(
        <div id="main-layout" onClick={onClick} style={style} className={className}>
            <ResponsiveHeaderBar toggleResponsive={toggleResponsive} setToggleResponsive={setToggleResponsive}/>
            <FlexBox id="master-container" className="row gap tall" style={{minHeight: "100vh"}}>
                {/* 
                    Left col: navigation
                    Right : page contents 
                */}
                <Loader load={load} timer={1000} >
                    <BrowserRouter>
                        <FlexBox className="navigation-col">
                        <NavBar akeyReq={akeyReq} footer={footer} extraNavigation={extraNavigation}  toggleResponsive={toggleResponsive} setToggleResponsive={setToggleResponsive} setNamespace={setNamespace} namespace={namespace} createNamespace={createNamespace} createMirrorNamespace={createMirrorNamespace} deleteNamespace={deleteNamespace} namespaces={data} />
                        </FlexBox>
                        <NamespaceNavigation  akey={akey} extraRoutes={extraRoutes} deleteNamespace={deleteNamespace} namespace={namespace} setNamespace={setNamespace} namespaces={data}/>
                    </BrowserRouter>
                </Loader>
            </FlexBox>
        </div>
    );
}

export default MainLayout;


function ResponsiveHeaderBar(props) {

    let {toggleResponsive, setToggleResponsive} = props;

    return(
        <FlexBox id="responsive-bar" className="hide-on-large" style={{
            flexDirection: "row-reverse"
        }}>
            <div style={{minWidth: "50px", maxWidth: "50px"}}>

            </div>
            <FlexBox>
                <img src={SmallLogo} alt="Direktiv" className="auto-margin" style={{
                    height: "42px"
                }}/>
            </FlexBox>
            <div className="menu-toggle-parent" style={{minWidth: "50px", maxWidth: "50px"}}>
                <div style={{float: "right", marginLeft: "12px"}}>
                    <Button onClick={(e) => {
                            setToggleResponsive(!toggleResponsive)
                            e.stopPropagation()
                        }} className="light small" style={{
                            marginTop: "5px",
                            marginLeft: "5px",
                            maxWidth: "32px",
                            paddingBottom: "8px"
                        }}>
                            <IoMenu className="auto-margin" style={{
                                fontSize: "18px"
                            }} />
                    </Button>
                </div>
            </div>
        </FlexBox>
    )
}
