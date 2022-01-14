import { useInstances, useNamespaceLogs } from "direktiv-react-hooks"
import { useEffect, useState } from "react"
import { VscGraph, VscTerminal } from "react-icons/vsc"
import ContentPanel, { ContentPanelBody, ContentPanelTitle } from "../../components/content-panel"
import FlexBox from "../../components/flexbox"
import HelpIcon from "../../components/help"
import Loader from "../../components/loader"
import { Config, copyTextToClipboard } from "../../util"
import * as dayjs from "dayjs"
import { AutoSizer, List, CellMeasurer, CellMeasurerCache, WindowScroller } from 'react-virtualized';
import { IoCopy, IoEye, IoEyeOff } from "react-icons/io5"
import {TerminalButton} from '../instance'

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
    const [width, setWidth] = useState(window.innerWidth);

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
            <FlexBox className="gap" style={{paddingRight:"8px", height:"100%"}}>
                <FlexBox style={{width:"600px"}}>
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
                                                    <IoCopy/> Copy {width > 999 ? <span>to Clipboard</span>:""}
                                            </TerminalButton>
                                            {follow ?
                                                <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"}>
                                                    <IoEyeOff/> Stop {width > 999 ? <span>watching</span>: ""}
                                                </TerminalButton>
                                                :
                                                <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"} >
                                                        <IoEye/> <div>Follow {width > 999 ? <span>logs</span>: ""}</div>
                                                </TerminalButton>
                                            }
                                        </FlexBox>
                                    </div>
                                </FlexBox>
                            </>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox className="gap" style={{flexDirection: "column"}}>
                    <FlexBox>
                        <ContentPanel style={{width:"100%"}}>
                            <ContentPanelTitle>
                                <FlexBox className="gap" style={{alignItems:"center"}}>
                                    <VscTerminal/>
                                    <div>
                                        Successful Executions
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
                                    <VscTerminal/>
                                    <div>
                                        Failed Executions
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
    const [qParams, setQParams] = useState(["first=10", "filter.field=STATUS", "filter-type=MATCH", "filter.val=failed"])

    const {data, err} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), qParams)
    console.log(data, err)
    
    return (
        ""
    )
}

function SuccessfulExecutions(props) {
    const {namespace} = props
    const [qParams, setQParams] = useState(["first=10", "filter.field=STATUS", "filter.type=MATCH", "filter.val=complete"])

    const {data, err} = useInstances(Config.url, true, namespace, localStorage.getItem("apikey"), ...qParams)
    console.log(data, err)

    return (
        ""
    )
}