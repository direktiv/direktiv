import { useInstances, useNamespaceLogs } from "direktiv-react-hooks"
import { useEffect, useState } from "react"
import { VscCheck, VscChromeClose, VscTerminal } from "react-icons/vsc"
import ContentPanel, { ContentPanelBody, ContentPanelTitle } from "../../components/content-panel"
import FlexBox from "../../components/flexbox"
import HelpIcon from "../../components/help"
import Loader from "../../components/loader"
import { Config, copyTextToClipboard } from "../../util"
import { AutoSizer, List, CellMeasurer, CellMeasurerCache } from 'react-virtualized';
import { VscCopy, VscEye, VscEyeClosed } from 'react-icons/vsc';
import {TerminalButton} from '../instance'
import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import { Link } from "react-router-dom"

dayjs.extend(utc)
dayjs.extend(relativeTime);


export default function Monitoring(props) {
    const {namespace} = props
    if(!namespace){
        return ""
    }

    return (
        <div style={{paddingRight:"8px", height:"100%"}}>
            <MonitoringPage noPadding namespace={namespace} />
        </div>
    )
}

function MonitoringPage(props) {
    const cache = new CellMeasurerCache({
        fixedWidth: false,
        defaultHeight: 20
    })
    const {namespace, noPadding} = props
    const [follow, setFollow] = useState(true)
    const [load, setLoad] = useState(true)
    const [width,] = useState(window.innerWidth);

    const [clipData, setClipData] = useState(null)
    const {data, err} = useNamespaceLogs(Config.url, true, namespace, localStorage.getItem('apikey'))
    let paddingStyle = { padding: "12px" }
    if (noPadding) {
        paddingStyle = { padding: "0px" }
    }

    useEffect(()=>{
        if(data !== null) {
            if(clipData === null) {
                let cd = ""
                for(let i=0; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`
                }
                setClipData(cd)
            }
            if(clipData !== data){
                let cd = ""
                for(let i=0; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`

                }
                setClipData(cd)
            }
        }
    },[data, clipData, setClipData])

    useEffect(()=>{
        if(data !== null || err !== null) {
            setLoad(false)
        }
    },[data, err])

    function rowRenderer({index, parent, key, style}) {
        if(!data[index]){
            return ""
        }

        return (
        <CellMeasurer
            key={key}
            cache={cache}
            parent={parent}
            columnIndex={0}
            rowIndex={index}
        >
          <div style={style}>
            <div style={{display:"inline-block",minWidth:"112px", color:"#b5b5b5"}}>
                <div className="log-timestamp">
                    <div>[</div>
                        <div style={{display: "flex", flex: "auto", justifyContent: "center"}}>{dayjs.utc(data[index].node.t).local().format("HH:mm:ss.SSS")}</div>
                    <div>]</div>
                </div>
            </div> 
            <span style={{marginLeft:"5px"}}>
                {data[index].node.msg}
            </span>
            <div style={{height: `fit-content`}}></div>
          </div>
          </CellMeasurer>
        );
    }

    return (
        <Loader load={load} timer={3000}>
            <FlexBox className="wrap gap" style={{paddingRight:"8px", height:"100%"}}>
                <FlexBox style={{width:"600px", flex: 2}}>
                    <ContentPanel style={{width:"100%"}}>
                        <ContentPanelTitle>
                            <FlexBox className="gap" style={{alignItems:"center"}}>
                                <VscTerminal/>
                                <div>
                                    Namespace Logs
                                </div>
                                <HelpIcon msg={"Namespace logs details action happening throughout the namespace"} />
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <>
                                <FlexBox className="col" style={{...paddingStyle}}>
                                    <FlexBox style={{ backgroundColor: "#002240", color: "white", borderRadius: "8px 8px 0px 0px", overflow: "hidden", padding: "8px" }}>
                                        <div style={{flex:"1 1 auto", lineHeight: "20px"}}>
                                            <AutoSizer>
                                                {({height, width})=>(
                                                    <div style={{height: "100%", minHeight: "100%"}}>
                                                    <List
                                                    width={width}
                                                    height={height}
                                                        // style={{
                                                        //     minHeight: "100%"
                                                        //     // maxHeight: "100%"
                                                        // }}
                                                        rowRenderer={rowRenderer}
                                                        deferredMeasurementCache={cache}
                                                        scrollToIndex={follow ? data.length - 1: 0}
                                                        rowCount={data.length}
                                                        rowHeight={cache.rowHeight}
                                                        />
                                                    </div>
                                                )}
                                            </AutoSizer>
                                        </div>
                                    </FlexBox>
                                    <div style={{ height: "40px", backgroundColor: "#223848", color: "white", maxHeight: "40px", minHeight: "40px", padding: "0px 10px 0px 10px", boxShadow: "0px 0px 3px 0px #fcfdfe", alignItems:'center', borderRadius: " 0px 0px 8px 8px", overflow: "hidden" }}>
                                        <FlexBox className="gap" style={{width: "100%", flexDirection: "row-reverse", height: "100%", alignItems: "center"}}>
                                            <TerminalButton onClick={()=>{
                                                copyTextToClipboard(clipData)
                                            }}>
                                                    <VscCopy/> Copy {width > 999 ? <span>to Clipboard</span>:""}
                                            </TerminalButton>
                                            {follow ?
                                                <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"}>
                                                    <VscEyeClosed/> Stop {width > 999 ? <span>watching</span>: ""}
                                                </TerminalButton>
                                                :
                                                <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"} >
                                                    <VscEye/> <div>Follow {width > 999 ? <span>logs</span>: ""}</div>
                                                </TerminalButton>
                                            }
                                        </FlexBox>
                                    </div>
                                </FlexBox>
                            </>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox className="gap" style={{flexDirection: "column", flex:1}}>
                    <FlexBox>
                        <ContentPanel style={{width:"100%"}}>
                            <ContentPanelTitle>
                                <FlexBox className="gap" style={{alignItems:"center"}}>
                                    <VscCheck fill={"var(--theme-green)"}/>
                                    <div>
                                        Successful <span className="hide-on-med">Executions</span>
                                    </div>
                                    <HelpIcon msg={"A list of the latest successful executions"} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody>
                                <SuccessfulExecutions namespace={namespace} />
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                    <FlexBox>
                        <ContentPanel style={{width:"100%"}}>
                            <ContentPanelTitle>
                                <FlexBox className="gap" style={{alignItems:"center"}}>
                                    <VscChromeClose fill={"var(--theme-red)"}/>
                                    <div>
                                        Failed <span className="hide-on-med">Executions</span>
                                    </div>
                                    <HelpIcon msg={"A list of the latest failed executions"} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody>
                                <FailedExecutions namespace={namespace} />
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </Loader>
    )
}

function FailedExecutions(props) {
    const {namespace} = props
    const [qParams] = useState(["first=5", "filter.field=STATUS", "filter-type=MATCH", "filter.val=failed"])

    const {data} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), ...qParams)
    
    // todo implement loading
    if(data === null) {
        return ""
    }

    return (
        <FlexBox style={{width:"100%"}}>
            <table style={{width:"100%", height:"fit-content"}}>
                <thead>
                    <tr>
                        <th>Workflow</th>
                        <th>Instance</th>
                        {/* <th>Updated</th> */}
                    </tr>
                </thead>
                <tbody>
                    {
                        data.map((obj)=>{
                            let split = obj.node.as.split(":")
                            let wf = split[0]
                            let revision = split[1]
                            if(!revision){
                                revision = "latest"
                            }
                            return(
                                <tr className="instance-row">
                                    <td title={obj.node.as} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/explorer/${wf}?tab=1&revision=${revision}&revtab=0`}>{obj.node.as}</Link></td>
                                    <td title={obj.node.id} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/instances/${obj.node.id}`}>
                                            {obj.node.id.split("-")[0]}
                                        </Link>
                                    </td>
                                    {/* <td>{dayjs.utc(obj.node.updatedAt).local().format("HH:mm a")}</td> */}
                                </tr>
                            )
                        })
                    }
                </tbody>
            </table>
        </FlexBox>
    )
}

function SuccessfulExecutions(props) {
    const {namespace} = props
    const [qParams] = useState(["first=5", "filter.field=STATUS", "filter.type=MATCH", "filter.val=complete"])

    const {data} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), ...qParams)
    // todo implement loading
    if(data === null) {
        return ""
    }

    return (
        <FlexBox style={{width:"100%"}}>
            <table style={{width:"100%", height:"fit-content"}}>
                <thead>
                    <tr>
                        <th>Workflow</th>
                        <th>Instance</th>
                        {/* <th>Updated</th> */}
                    </tr>
                </thead>
                <tbody>
                    {
                        data.map((obj)=>{
                            let split = obj.node.as.split(":")
                            let wf = split[0]
                            let revision = split[1]
                            if(!revision){
                                revision = "latest"
                            }
                            return(
                                <tr className="instance-row">
                                    <td title={obj.node.as} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/explorer/${wf}?tab=1&revision=${revision}&revtab=0`}>{obj.node.as}</Link></td>
                                    <td title={obj.node.id} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/instances/${obj.node.id}`}>
                                            {obj.node.id.split("-")[0]}
                                        </Link>
                                    </td>
                                    {/* <td>{dayjs.utc(obj.node.updatedAt).local().format("HH:mm a")}</td> */}
                                </tr>
                            )
                        })
                    }
                </tbody>
            </table>
        </FlexBox>
    )
}