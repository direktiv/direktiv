import { useInstances } from 'direktiv-react-hooks';
import { useEffect, useMemo, useState } from 'react';
import { BsDot } from 'react-icons/bs';
import { VscClose, VscVmRunning } from 'react-icons/vsc';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import HelpIcon from '../../components/help';
import Pagination, { usePageHandler } from '../../components/pagination';
import { Config, GenerateRandomKey } from '../../util';
import './style.css';

import Tippy from '@tippyjs/react';
import * as dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc";
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
        <FlexBox className="col gap" style={{ paddingRight: "8px" }}>
            <InstancesTable namespace={namespace}/>
        </FlexBox>
    );
}

export default InstancesPage;

export function InstancesTable(props) {
    const {namespace, mini, hideTitle, panelStyle, bodyStyle, filter, placeholder} = props
    const [load, setLoad] = useState(true)

    const [filterName, setFilterName] = useState("")
    const [filterCreatedBefore, setFilterCreatedBefore] = useState("")
    const [filterCreatedAfter, setFilterCreatedAfter] = useState("")
    const [filterState, setFilterState] = useState("")
    const [filterInvoker, setFilterInvoker] = useState("")
    const queryFilters = useMemo(()=>{
        if (filter) {
            return filter
        }

        let newFilters = []
        if (filterName !== ""){
            newFilters.push(`filter.field=AS&filter.type=CONTAINS&filter.val=${filterName}`)
        }

        if (filterCreatedBefore !== ""){
            newFilters.push(`filter.field=CREATED&filter.type=BEFORE&filter.val=${encodeURIComponent(new Date(filterCreatedBefore).toISOString())}`)
        }

        if (filterCreatedAfter !== ""){
            newFilters.push(`filter.field=CREATED&filter.type=AFTER&filter.val=${encodeURIComponent(new Date(filterCreatedAfter).toISOString())}`)
        }


        if (filterState !== ""){
            newFilters.push(`filter.field=STATUS&filter.type=MATCH&filter.val=${filterState}`)
        }

        if (filterInvoker !== ""){
            newFilters.push(`filter.field=TRIGGER&filter.type=CONTAINS&filter.val=${filterInvoker}`)
        }

        return newFilters
    }, [filter, filterName, filterCreatedBefore, filterCreatedAfter, filterState, filterInvoker])

    const pageHandler = usePageHandler(PAGE_SIZE)
    const goToFirstPage = pageHandler.goToFirstPage
    const {data, err, pageInfo} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), pageHandler.pageParams, ...queryFilters)

    useEffect(()=>{
        if(data !== null || err !== null) {
            setLoad(false)
        }
    },[data, err])

    // Reset Page to start when filters changes
    useEffect(()=>{
        // TODO: This will interfere with page position if initPage > 1
        goToFirstPage()
    },[queryFilters, goToFirstPage])

    return(
        <Loader load={load} timer={3000}>
        {hideTitle ?<></>:
        <FlexBox className="gap instance-filter" style={{justifyContent: "space-between", alignItems: "center", paddingBottom: "8px", flexGrow:"0"}}>
            <FlexBox className="col gap">
                <FlexBox className="row center-y gap">
                    Filter Name
                    {filterName === "" ?  <></> : <div className="filter-close-btn" onClick={()=>{setFilterName("")}}><VscClose/></div>}
                </FlexBox>
                <input type="search" placeholder="Instance Name" value={filterName} onChange={e=>{
                    setFilterName(e.target.value)
                }}/>
            </FlexBox>
            <FlexBox className="col gap" >
                <FlexBox className="row center-y gap">
                    Filter State
                    {filterState === "" ?  <></> : <div className="filter-close-btn" onClick={()=>{setFilterState("")}}><VscClose/></div>}
                </FlexBox>
                <select value={filterState} style={{color: `${filterState === "" ? "gray": "#082032"}`}} onChange={(e)=>{
                        setFilterState(e.target.value)
                        }}>
                    <option value="" disabled selected hidden>Choose State</option>
                    <option value="complete">Complete</option>
                    <option value="failed">Failed</option>
                    <option value="pending">Running</option>
                    <option value="cancelled">Cancelled</option>
                </select>
            </FlexBox>
            <FlexBox className="col gap" >
                <FlexBox className="row center-y gap">
                    Filter Invoker
                    {filterInvoker === "" ?  <></> : <div className="filter-close-btn" onClick={()=>{setFilterInvoker("")}}><VscClose/></div>}
                </FlexBox>
                <select value={filterInvoker} style={{color: `${filterInvoker === "" ? "gray": "#082032"}`}} onChange={(e)=>{
                        setFilterInvoker(e.target.value)
                        }}>
                    <option value="" disabled selected hidden>Choose filter option</option>
                    <option value="api">API</option>
                    <option value="cloudevent">Cloud Event</option>
                    <option value="instance">Instance</option>
                    <option value="cron">Cron</option>
                </select>
            </FlexBox>
            <FlexBox className="col gap" >
                <FlexBox className="row center-y gap">
                    Filter Created Before
                    {filterCreatedBefore === "" ?  <></> : <div className="filter-close-btn" onClick={()=>{setFilterCreatedBefore("")}}><VscClose/></div>}
                </FlexBox>
                <input type="datetime-local" style={{color: `${filterCreatedBefore === "" ? "gray": "#082032"}`}} value={filterCreatedBefore} required onChange={e=>{
                    setFilterCreatedBefore(e.target.value)
                }}/>
            </FlexBox>
            <FlexBox className="col gap" >
                <FlexBox className="row center-y gap">
                    Filter Created After
                    {filterCreatedAfter === "" ?  <></> : <div className="filter-close-btn" onClick={()=>{setFilterCreatedAfter("")}}><VscClose/></div>}
                </FlexBox>
                <input type="datetime-local" style={{color: `${filterCreatedAfter === "" ? "gray": "#082032"}`}} value={filterCreatedAfter} required onChange={e=>{
                    setFilterCreatedAfter(e.target.value)
                }}/>
            </FlexBox>
        </FlexBox>}

        <ContentPanel style={{...panelStyle}}>
        {hideTitle ?<></>:
        <>
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
        </>}
        <ContentPanelBody style={{...bodyStyle}}>
        {
            data !== null && data.length === 0 ? 
                <div style={{paddingLeft:"10px", fontSize:"10pt"}}>{`${placeholder ? placeholder : "No instances have been recently executed. Recent instances will appear here."}`}</div>
            :
                <table className="instances-table" style={{width: "100%"}}>
                    <thead>
                        <tr>
                            
                            <th className="center-align" style={{maxWidth: "120px", minWidth: "120px", width: "120px"}}>State</th>
                            <th className="center-align">Name</th>
                            {mini ? <></>:<th className="center-align">Revision ID</th>}
                            {mini ? <></>:<th className="center-align">Invoker</th>}
                            <th className="center-align">Started <span className="hide-1000">at</span></th>
                            {mini ? <></>:<th className="center-align"><span className="hide-1000">Last</span> Updated</th>}
                        </tr>
                    </thead>
                    <tbody>
                        {data !== null ? 
                        <>
                            <>
                            {data.map((obj)=>{
                            return(
                                <InstanceRow 
                                    mini={mini}
                                    key={GenerateRandomKey()}
                                    namespace={namespace}
                                    state={obj.status} 
                                    name={obj.as} 
                                    id={obj.id}
                                    invoker={obj.invoker}
                                    startedDate={dayjs.utc(obj.createdAt).local().format("DD MMM YY")} 
                                    startedTime={dayjs.utc(obj.createdAt).local().format("HH:mm a")} 
                                    finishedDate={dayjs.utc(obj.updatedAt).local().format("DD MMM YY")}
                                    finishedTime={dayjs.utc(obj.updatedAt).local().format("HH:mm a")} 
                                />
                            )
                            })}</>
                        </>
                        :""}
                    </tbody>
                </table>
        }
        </ContentPanelBody>
        </ContentPanel>
        <FlexBox className="row" style={{justifyContent:"flex-end", paddingBottom:"1em", flexGrow: 0}}>
            <Pagination pageHandler={pageHandler} pageInfo={pageInfo}/>
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
    let {state, name, wf, startedDate,  finishedDate, startedTime, finishedTime,  id, namespace, mini, invoker} = props;
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
    pathwf.pop()
    
    return(
    
    <tr onClick={()=>{
        navigate(`/n/${namespace}/instances/${id}`)
    }} className="instance-row" style={{cursor: "pointer"}}>
        <td className="label-cell">
            {label}
        </td>
        {!wf ? 
        <Tippy content={`/${wfStr}`} trigger={'mouseenter focus'} zIndex={10}>
        <td className="center-align" style={{ fontSize: "12px", lineHeight: "20px", display:"flex", justifyContent:"center", marginTop:"12px", whiteSpace: "nowrap"}}>
            {pathwf.length > 0 ?
            <div style={{marginLeft:"10px", textOverflow:"ellipsis", overflow:"hidden"}}>
                /{pathwf.join("/")}
            </div> :
            <></>
            }
            <div>
                /{wfname}
            </div>
            
        </td>
        </Tippy>: ""}
        {mini ? <></>:<td title={revStr} style={{ fontSize: "12px", lineHeight: "20px", textOverflow:"ellipsis", overflow:"hidden", color: revStr !== undefined ? "" : "var(--theme-dark-gray-text)" }} className="center-align">
            {revStr !== undefined ? revStr : "ROUTER"}
        </td>}
        {mini || wf ? <></>:<td title={invoker} style={{ fontSize: "12px", lineHeight: "20px", textOverflow:"ellipsis", overflow:"hidden" }} className="center-align">
            {/* Trim instance id from invoker label */}
            {invoker !== undefined &&  invoker !== "" ? invoker.split(":")[0] : "NA"}
        </td>}
        <td className="center-align">
            <span className="hide-864">{startedDate}, </span>
            {startedTime}
        </td>
        {mini ? <></>:<td className="center-align">
            <span className="hide-864">{finishedDate}, </span>
            {finishedTime}
        </td>}
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
                <div className="hide-1000" style={{marginLeft: "-8px", marginRight: "16px"}}>{label}</div>
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

