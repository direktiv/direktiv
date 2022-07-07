import React, { useCallback, useState } from 'react';
import { Config } from '../../util';
import {  VscCloud, VscPlay, VscDebugStepInto } from 'react-icons/vsc';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {useEvents} from 'direktiv-react-hooks'
import Modal, { ButtonDefinition } from '../../components/modal';
import DirektivEditor from '../../components/editor';
import { AutoSizer } from 'react-virtualized';
import HelpIcon from "../../components/help";

import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import './style.css'
import { Link } from 'react-router-dom';
import Pagination from '../../components/pagination';

dayjs.extend(utc)
dayjs.extend(relativeTime);



function EventsPageWrapper(props) {

    let {namespace} = props;
    if (!namespace) {
        return <></>
    }

    return (
        <EventsPage namespace={namespace} />
    )
}

export default EventsPageWrapper;

const PAGE_SIZE = 8;

function EventsPage(props) {

    let {namespace} = props;

    // errHistory and errListeners TODO show error if one
    const [listenersParam, setListenersParam] = useState([`first=${PAGE_SIZE}`])
    const [historyParam, setHistoryParam] = useState([`first=${PAGE_SIZE}`])
    let {eventHistory, eventListeners, eventListenersTotalCount, eventListenersPageInfo, eventHistoryTotalCount, eventHistoryPageInfo, sendEvent, replayEvent} = useEvents(Config.url, true, namespace, localStorage.getItem("apikey"), {listeners: listenersParam, history: historyParam})
    const updateEventHistoryPage = useCallback((newParam)=>{
        setHistoryParam([...newParam])
    }, [])
    const updateListenersPage = useCallback((newParam)=>{
        setListenersParam([...newParam])
    }, [])
    return(
        <>
            <FlexBox className="gap col" style={{paddingRight: "8px"}}>
                <FlexBox>
                    <ContentPanel style={{ width: "100%" }}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscCloud/>
                            </ContentPanelTitleIcon>
                            <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                <div>
                                    Cloud Events History
                                </div>
                                <HelpIcon msg={"A history of events that have hit this specific namespace."} />
                            </FlexBox>
                            <SendEventModal sendEvent={sendEvent}/>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <div style={{maxHeight: "40vh", overflowY: "auto", fontSize: "12px", minWidth: "100%"}}>
                                <table className="cloudevents-table" style={{minWidth: "440px", width: "100%"}}>
                                    <thead>
                                        <tr>
                                            <th>
                                                Type
                                            </th>
                                            <th style={{width:"250px"}}>
                                                Source
                                            </th>
                                            <th>
                                                Time
                                            </th>
                                            <th style={{textAlign:'center'}}>
                                                  Actions
                                            </th>
                                        </tr>
                                    </thead>
                                    {eventHistory !== null && typeof eventHistory === typeof [] && eventHistory.length > 0 ?
                                    <tbody>
                                        {eventHistory.map((obj) => {
                                            return <tr style={{ borderBottom: "1px solid #f4f4f4" }}>
                                                <td title={obj.node.type} style={{ textOverflow: "ellipsis", overflow: "hidden" }}>
                                                    {obj.node.type}
                                                </td>
                                                <td title={obj.node.source} style={{ textOverflow: "ellipsis", overflow: "hidden" }}>
                                                    {obj.node.source}
                                                </td>
                                                <td>
                                                    {dayjs.utc(obj.node.receivedAt).local().format("HH:mm:ss a")}
                                                </td>
                                                <td style={{ textAlign: 'center', justifyContent: "center", }}>
                                                    <FlexBox className={"gap center"}>
                                                        <Modal
                                                            className="run-workflow-modal"
                                                            style={{ justifyContent: "flex-end" }}
                                                            modalStyle={{ color: "black", width: "360px" }}
                                                            title="Retrigger Event"
                                                            onClose={() => {
                                                            }}
                                                            btnStyle={{ width: "unset" }}
                                                            button={
                                                                <Button className="small light bold" tip="Retrigger Event">
                                                                    <VscPlay /> <span className='hide-800'>Retrigger</span>
                                                                </Button>
                                                            }
                                                            actionButtons={[
                                                                ButtonDefinition("Retrigger", async () => {
                                                                    await replayEvent(obj.node.id)
                                                                }, "small", () => { }, true, true),
                                                                ButtonDefinition("Cancel", async () => {
                                                                }, "small light", () => { }, true, false)
                                                            ]}
                                                        >
                                                            <FlexBox style={{ overflow: "hidden" }}>
                                                                Are you sure you want to retrigger {obj.node.id}?
                                                            </FlexBox>
                                                        </Modal>
                                                        <Modal
                                                            className="run-workflow-modal"
                                                            modalStyle={{ color: "black", minWidth: "360px", width: "50vw", height: "40vh", minHeight: "680px" }}
                                                            title="View Event"
                                                            onClose={() => {
                                                            }}
                                                            btnStyle={{ width: "unset" }}
                                                            button={
                                                                <Button className="small light bold">
                                                                    View
                                                                </Button>}
                                                            actionButtons={[
                                                                ButtonDefinition("Close", async () => {
                                                                }, "small light", () => { }, true, false)
                                                            ]}
                                                        >
                                                            <FlexBox className="col" style={{ overflow: "hidden" }}>
                                                                <AutoSizer>
                                                                    {({ height, width }) => (
                                                                        <DirektivEditor noBorderRadius value={atob(obj.node.cloudevent)} readonly={true} dlang="plaintext"
                                                                            options={{
                                                                                autoLayout: true
                                                                            }}
                                                                            width={width}
                                                                            height={height}
                                                                        />
                                                                    )}
                                                                </AutoSizer>
                                                            </FlexBox>
                                                        </Modal>
                                                    </FlexBox>
                                                </td>
                                            </tr>
                                        })}
                                    </tbody> : 
                                    <FlexBox className='table-no-content'>
                                        No cloud events history
                                    </FlexBox>
                                }
                                </table>
                            </div>
                            {
                                !!eventHistoryTotalCount && 
                                <Pagination
                                    pageSize={PAGE_SIZE}
                                    total={eventHistoryTotalCount}
                                    pageInfo={eventHistoryPageInfo}
                                    updatePage={updateEventHistoryPage}
                                />
                            }
                        </ContentPanelBody>                 
                    </ContentPanel>
                </FlexBox>
                <FlexBox>
                    <ContentPanel style={{ width: "100%" }}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscDebugStepInto/>
                            </ContentPanelTitleIcon>
                            <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                <div>
                                    Active Event Listeners
                                </div>
                                <HelpIcon msg={"Current listeners in a namespace that are listening for a cloud a event."} />
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <div style={{maxHeight: "40vh", overflowY: "auto", fontSize: "12px"}}>
                                <table className="event-listeners-table" style={{width: "100%"}}>
                                    <thead>
                                        <tr>
                                            <th>
                                                Workflow
                                            </th>
                                            <th>
                                                Type
                                            </th>
                                            <th>
                                                Mode
                                            </th>
                                            <th>
                                                Updated
                                            </th>
                                            <th>
                                                Event Types
                                            </th>
                                        </tr>
                                    </thead>
                                    {eventListeners !== null && typeof eventListeners === typeof [] && eventListeners?.length > 0 ?
                                    <tbody>
                                        {eventListeners.map((obj)=>{
                                            return(
                                                <tr  style={{borderBottom:"1px solid #f4f4f4"}}>
                                                    <td style={{textOverflow:"ellipsis", overflow:"hidden"}}>
                                                        <Link style={{color:"#2396d8"}} to={`/n/${namespace}/explorer${obj.node.workflow}`}>
                                                            {obj.node.workflow}
                                                        </Link> 
                                                    </td>
                                                    <td style={{textOverflow:"ellipsis", overflow:"hidden"}}>
                                                        {obj.node.instance !== "" ? <Link style={{color:"#2396d8"}} to={`/n/${namespace}/instances/${obj.node.instance}`}>{obj.node.instance.split("-")[0]}</Link> : "start"}
                                                    </td>
                                                    <td style={{textOverflow:"ellipsis", overflow:"hidden"}}>
                                                        {obj.node.mode}
                                                    </td>
                                                    <td>
                                                        {dayjs.utc(obj.node.updatedAt).local().format("HH:mm:ss a")}
                                                    </td>
                                                    <td className="event-split">
                                                        {obj.node.events.map((obj)=>{
                                                            return <span>{obj.type}</span>
                                                        })}
                                                    </td>
                                                </tr>
                                            )
                                        })}
                                    </tbody>:
                                        <FlexBox className='table-no-content'>
                                            No active event listeners
                                        </FlexBox>
                                    }
                                </table>
                            </div>
                            {
                                !!eventListenersTotalCount && 
                                <Pagination
                                    pageSize={PAGE_SIZE}
                                    total={eventListenersTotalCount}
                                    pageInfo={eventListenersPageInfo}
                                    updatePage={updateListenersPage}
                                />
                            }
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
            </FlexBox>
        </>
    )
}

function SendEventModal(props) {

    const {sendEvent} = props
    let [eventData, setEventData] = useState(`{
    "specversion" : "1.0",
    "type" : "com.github.pull.create",
    "source" : "https://github.com/cloudevents/spec/pull",
    "subject" : "123",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/xml",
    "data" : "<much wow=\\"xml\\"/>"
}`);

    return (<>
        <Modal
            title="Send New Event"
            button={(
                <ContentPanelHeaderButton>
                    <div>
                        Send New Event
                    </div>
                </ContentPanelHeaderButton>
            )}
            actionButtons={[
                ButtonDefinition("Send", async () => {
                    await sendEvent(eventData)
                }, "small", ()=>{}, true, false),
                ButtonDefinition("Cancel", () => {}, "small light", ()=>{}, true, false)
            ]}
            noPadding
        >
            <FlexBox className="col gap" style={{overflow: "hidden"}}>
                <FlexBox style={{ minHeight: "40vh", minWidth: "70vw" }}>
                    <AutoSizer>
                        {({height, width})=>(
                        <DirektivEditor noBorderRadius value={eventData} setDValue={setEventData} dlang="json" 
                            options={{
                                autoLayout: true
                            }} 
                            width={width}
                            height={height}
                        />
                        )}
                    </AutoSizer>
                </FlexBox>
            </FlexBox>
        </Modal>
    </>)
}