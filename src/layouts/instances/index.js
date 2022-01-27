import React, { useCallback, useEffect, useState } from 'react';
import './style.css';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import Pagination from '../../components/pagination';
import FlexBox from '../../components/flexbox';
import { VscVmRunning } from 'react-icons/vsc';
import { BsDot } from 'react-icons/bs';
import HelpIcon from '../../components/help';
import { useInstances } from 'direktiv-react-hooks';
import { Config, GenerateRandomKey } from '../../util';

import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import { useNavigate } from 'react-router-dom';
import Loader from '../../components/loader';

const PAGE_SIZE = 10

dayjs.extend(utc)
dayjs.extend(relativeTime);

function InstancesPage(props) {
    const {namespace} = props
    if(!namespace) {
        return <></>
    }
    return(
        <div style={{ paddingRight: "8px" }}>
            <InstancesTable namespace={namespace}/>
        </div>
    );
}

export default InstancesPage;

function InstancesTable(props) {
    const {namespace} = props
    const [load, setLoad] = useState(true)
    
    const [queryParams, setQueryParams] = useState([`first=${PAGE_SIZE}`])
    const {data, err, pageInfo, totalCount} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), ...queryParams)

    const updatePage = useCallback((newParam)=>{
        setQueryParams(newParam)
    }, [])

    useEffect(()=>{
        if(data !== null || err !== null) {
            setLoad(false)
        }
    },[data, err])
    return(
        <Loader load={load} timer={3000}>

        <ContentPanel>
        <ContentPanelTitle>
            <ContentPanelTitleIcon>
                <VscVmRunning/>
            </ContentPanelTitleIcon>
            <FlexBox className="gap" style={{ alignItems: "center" }}>
                <div>
                    Instances
                </div>
                <HelpIcon msg={"A list of recently executed instances."} />
            </FlexBox>
        </ContentPanelTitle>
        <ContentPanelBody>
        <>
        {
            data !== null && data.length === 0 ? 
                <div style={{paddingLeft:"10px", fontSize:"10pt"}}>No instances have been recently executed. Recent instances will appear here.</div>
            :
                <table className="instances-table" style={{width: "100%"}}>
                    <thead>
                        <tr>
                            <th className="center-align" style={{maxWidth: "120px", minWidth: "120px", width: "120px"}}>
                                State
                            </th>
                            <th className="center-align">
                                Name
                            </th>
                            <th className="center-align">
                                Revision ID
                            </th>
                            <th className="center-align">
                                Started <span className="hide-on-med">at</span>
                            </th>
                            <th className="center-align">
                                <span className="hide-on-med">Last</span> Updated
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {data !== null ? 
                        <>
                            <>
                            {data.map((obj)=>{
                            return(
                                <InstanceRow 
                                    key={GenerateRandomKey()}
                                    namespace={namespace}
                                    state={obj.node.status} 
                                    name={obj.node.as} 
                                    id={obj.node.id}
                                    started={dayjs.utc(obj.node.createdAt).local().format("HH:mm a")} 
                                    startedFrom={dayjs.utc(obj.node.createdAt).local().fromNow()}
                                    finished={dayjs.utc(obj.node.updatedAt).local().format("HH:mm a")}
                                    finishedFrom={dayjs.utc(obj.node.updatedAt).local().fromNow()}
                                />
                            )
                            })}</>
                        </>
                        :""}
                    </tbody>
                </table>
        }
        </>
        </ContentPanelBody>
    </ContentPanel>
    <FlexBox>
        {!!totalCount && <Pagination pageSize={PAGE_SIZE} pageInfo={pageInfo} updatePage={updatePage} total={totalCount}/>}
    </FlexBox>
    </Loader>
        
    );
}

const success = "complete";
const fail = "failed";
const crashed = "crashed";
// there is no cancelled state
const cancelled = "cancelled";
const running = "pending";

export function InstanceRow(props) {
    let {state, name, wf, started,  finished,  id, namespace} = props;
    const navigate = useNavigate()

    let label;
    if (state === success) {
        label = <SuccessState />
    } else if (state === cancelled) {
        label = <CancelledState />
    } else if (state === fail || state === crashed) {
        label = <FailState />
    }  else  if (state === running) {
        label = <RunningState />
    }

    let wfStr = name.split(':')[0]
    let revStr = name.split(':')[1]

    let pathwf = wfStr.split("/")
    let wfname = pathwf[pathwf.length-1]
    pathwf.splice(pathwf.length-1, pathwf.length-1)
    
    return(
    
    <tr onClick={()=>{
        navigate(`/n/${namespace}/instances/${id}`)
    }} className="instance-row" style={{cursor: "pointer"}}>
        <td className="label-cell">
            {label}
        </td>
        {!wf ? 
        <td title={wfStr} className="center-align" style={{ fontSize: "12px", lineHeight: "20px", display:"flex", justifyContent:"center", marginTop:"12px"}}>
            <div style={{marginLeft:"10px", textOverflow:"ellipsis", overflow:"hidden"}}>
                /{pathwf.join("/")}
            </div>
            {pathwf.length !== 1 ?
            <div>
                /{wfname}
            </div>:""}
        </td>: ""}
        <td title={revStr} style={{ fontSize: "12px", lineHeight: "20px", textOverflow:"ellipsis", overflow:"hidden" }} className="center-align">
            {revStr}
        </td>
        <td className="center-align">
            {started}
            {/* <span style={{fontSize:"12px", marginLeft:"3px"}} className="grey-text hide-on-med">({startedFrom})</span> */}
        </td>
        <td className="center-align">
            {finished}
            {/* <span style={{fontSize:"12px", marginLeft:"3px"}} className="grey-text hide-on-med">({finishedFrom})</span> */}
        </td>
    </tr>
    )
}

function StateLabel(props) {

    let {className, label} = props;
    className += " label-cell"

    return (
        <div>
            <FlexBox className={className} style={{ alignItems: "center", padding: "0px", width: "fit-content" }} >
                <BsDot style={{ height: "32px", width: "32px" }} />
                <div className="hide-on-med" style={{marginLeft: "-8px", marginRight: "16px"}}>{label}</div>
            </FlexBox>
        </div>
    )
}

export function SuccessState() {
    return (
        <StateLabel className={"success-label"} label={"Successful"} />
    )
}

export function FailState() {
    return (
        <StateLabel className={"fail-label"} label={"Failed"} />
    )
}

export function CancelledState() {
    return (
        <StateLabel className={"cancel-label"} label={"Cancelled"} />
    )
}

export function RunningState() {
    return (
        <StateLabel className={"running-label"} label={"Running"} />
    )
}

