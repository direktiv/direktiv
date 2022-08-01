import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime"
import utc from "dayjs/plugin/utc"
import { useInstances, useNamespaceLogs } from "direktiv-react-hooks"
import { useEffect, useState } from "react"
import { VscCheck, VscChromeClose, VscTerminal } from "react-icons/vsc"
import { Link } from "react-router-dom"
import ContentPanel, { ContentPanelBody, ContentPanelTitle } from "../../components/content-panel"
import FlexBox from "../../components/flexbox"
import HelpIcon from "../../components/help"
import Loader from "../../components/loader"
import Logs, { LogFooterButtons } from "../../components/logs/logs"
import { Config } from "../../util"

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

    const {namespace, noPadding} = props
    const [follow, setFollow] = useState(true)
    const [load, setLoad] = useState(true)

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
                    cd += `[${dayjs.utc(data[i].t).local().format("HH:mm:ss.SSS")}] ${data[i].msg}\n`
                }
                setClipData(cd)
            }
            if(clipData !== data){
                let cd = ""
                for(let i=0; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].t).local().format("HH:mm:ss.SSS")}] ${data[i].msg}\n`

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
                                {data !== null ?
                                    <FlexBox className="col" style={{ ...paddingStyle }}>
                                        <FlexBox className={"logs"}>
                                            <Logs setFollow={setFollow} follow={follow} logItems={data} wordWrap={true}/>
                                        </FlexBox>
                                        <div style={{ height: "40px", backgroundColor: "#223848", color: "white", maxHeight: "40px", minHeight: "40px", padding: "0px 10px 0px 10px", boxShadow: "0px 0px 3px 0px #fcfdfe", alignItems: 'center', borderRadius: " 0px 0px 8px 8px", overflow: "hidden" }}>
                                            <FlexBox className="gap" style={{ width: "100%", flexDirection: "row-reverse", height: "100%", alignItems: "center" }}>
                                                <LogFooterButtons setFollow={setFollow} follow={follow} data={data}/>
                                            </FlexBox>
                                        </div>
                                    </FlexBox>
                                    : <div className="container-alert">Failed to get logs{err ? `: ${err}` : ``}</div>}
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
                                        Successful <span className="hide-1000">Executions</span>
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
                                        Failed <span className="hide-1000">Executions</span>
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
    const [qParams] = useState(["limit=5", "filter.field=STATUS", "filter.type=MATCH", "filter.val=failed"])

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
                            let split = obj.as.split(":")
                            let wf = split[0]
                            let revision = split[1]
                            if(!revision){
                                revision = "latest"
                            }
                            return(
                                <tr className="instance-row">
                                    <td title={obj.as} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/explorer/${wf}?tab=1&revision=${revision}&revtab=0`}>{obj.as}</Link></td>
                                    <td title={obj.id} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/instances/${obj.id}`}>
                                            {obj.id.split("-")[0]}
                                        </Link>
                                    </td>
                                    {/* <td>{dayjs.utc(obj.updatedAt).local().format("HH:mm a")}</td> */}
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
    const [qParams] = useState(["limit=5", "filter.field=STATUS", "filter.type=MATCH", "filter.val=complete"])

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
                            let split = obj.as.split(":")
                            let wf = split[0]
                            let revision = split[1]
                            if(!revision){
                                revision = "latest"
                            }
                            return(
                                <tr className="instance-row">
                                    <td title={obj.as} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/explorer/${wf}?tab=1&revision=${revision}&revtab=0`}>{obj.as}</Link></td>
                                    <td title={obj.id} style={{overflow:"hidden", textOverflow:"ellipsis"}}>
                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/instances/${obj.id}`}>
                                            {obj.id.split("-")[0]}
                                        </Link>
                                    </td>
                                    {/* <td>{dayjs.utc(obj.updatedAt).local().format("HH:mm a")}</td> */}
                                </tr>
                            )
                        })
                    }
                </tbody>
            </table>
        </FlexBox>
    )
}