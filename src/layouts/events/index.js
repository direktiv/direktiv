import React from 'react';
import { Config } from '../../util';
import { VscCloud, VscRedo, VscSymbolEvent } from 'react-icons/vsc';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {useEvents} from 'direktiv-react-hooks'

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

    let resp = useEvents(Config.url, true, namespace)
    console.log(resp);

    return(
        <>
            <FlexBox className="gap col" style={{paddingRight: "8px"}}>
                <FlexBox>
                    <ContentPanel style={{ width: "100%" }}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscCloud/>
                            </ContentPanelTitleIcon>
                            <div>
                                Cloud Events History
                            </div>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <div style={{maxHeight: "40vh", overflowY: "auto"}}>
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
                                            <div style={{display: "flex", alignItems: "flex-end", justifyContent: "right", paddingRight: "18px"}}>
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
                            <div>Active Event Listeners</div>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <div style={{maxHeight: "40vh", overflowY: "auto"}}>
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