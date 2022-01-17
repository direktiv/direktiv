import { useNamespaceLogs } from "direktiv-react-hooks"
import { useEffect, useState } from "react"
import {  VscTerminal } from "react-icons/vsc"
import ContentPanel, { ContentPanelBody, ContentPanelTitle } from "../../components/content-panel"
import FlexBox from "../../components/flexbox"
import HelpIcon from "../../components/help"
import Loader from "../../components/loader"
import { Config } from "../../util"

export default function Monitoring(props) {
    const {namespace} = props
    if(!namespace){
        return ""
    }

    return (
        <div style={{paddingRight:"8px", height:"100%"}}>
            <MonitoringPage namespace={namespace} />
        </div>
    )
}

function MonitoringPage(props) {
    const {namespace} = props

    const [load, setLoad] = useState(true)
    const {data, err} = useNamespaceLogs(Config.url, true, namespace, localStorage.getItem('apikey'))
    console.log(data, err)
    useEffect(()=>{
        if(data !== null || err !== null) {
            setLoad(false)
        }
    },[data, err])

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
                            x
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
                                        Last X Success Workflows
                                    </div>
                                    <HelpIcon msg={"Namespace logs details action happening throughout the namespace"} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody>
                                x
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                    <FlexBox>
                        <ContentPanel style={{width:"100%"}}>
                            <ContentPanelTitle>
                                <FlexBox className="gap" style={{alignItems:"center"}}>
                                    <VscTerminal/>
                                    <div>
                                        Last X Failed Workflows
                                    </div>
                                    <HelpIcon msg={"Namespace logs details action happening throughout the namespace"} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody>
                                x
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </Loader>
    )
}