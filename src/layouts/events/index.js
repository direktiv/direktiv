import React, { useState } from 'react';
import { Config } from '../../util';
import {  VscCloud, VscRedo, VscSymbolEvent } from 'react-icons/vsc';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {useEvents} from 'direktiv-react-hooks'
import Modal, { ButtonDefinition } from '../../components/modal';
import DirektivEditor from '../../components/editor';
import { AutoSizer } from 'react-virtualized';
import HelpIcon from "../../components/help";


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

function EventsPage(props) {

    let {namespace} = props;
    console.log(useEvents);

    let {getEventListeners, sendEvent} = useEvents(Config.url, true, namespace)
    console.log(getEventListeners);
    console.log(sendEvent);


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
                            <div style={{maxHeight: "40vh", overflowY: "auto", fontSize: "12px"}}>
                                <table>
                                    <thead>
                                        <tr>
                                            <th>
                                                Type
                                            </th>
                                            <th>
                                                Source
                                            </th>
                                            <th>
                                                Time
                                            </th>
                                            <th>
                                                <div style={{display: "flex", alignItems: "flex-end", justifyContent: "right"}}>
                                                    Retrigger  
                                                </div>
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <td>
                                            azure.example
                                        </td>
                                        <td>
                                            Azure
                                        </td>
                                        <td>
                                            2.30pm (a while ago)
                                        </td>
                                        <td>
                                            <div style={{display: "flex", alignItems: "flex-end", justifyContent: "right", paddingRight: "10px"}}>
                                                <Button className="small light">
                                                    <VscRedo/>
                                                </Button>
                                            </div>
                                        </td>
                                    </tbody>
                                </table>
                            </div>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox>
                    <ContentPanel style={{ width: "100%" }}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscSymbolEvent/>
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
                                <table>
                                    <thead>
                                        <tr>
                                            <th>
                                                Workflow
                                            </th>
                                            <th>
                                                Type
                                            </th>
                                            <th>
                                                Source
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        <td>
                                            /workflows/my-test-workflow
                                        </td>
                                        <td>
                                            azure.example
                                        </td>
                                        <td>
                                            Azure
                                        </td>
                                    </tbody>
                                </table>
                            </div>
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
                    let err = await sendEvent(eventData)
                    if (err) return err
                }, "small", true, false),
                ButtonDefinition("Cancel", () => {}, "small light", true, false)
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