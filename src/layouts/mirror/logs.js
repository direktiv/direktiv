import React, { useEffect, useState } from 'react';
import './style.css';
import { AutoSizer, List, CellMeasurer, CellMeasurerCache } from 'react-virtualized';

import FlexBox from '../../components/flexbox';
import { VscCopy, VscEye, VscEyeClosed, } from 'react-icons/vsc';
import { Config, copyTextToClipboard, } from '../../util';
import { useMirrorLogs, } from 'direktiv-react-hooks';

import * as dayjs from "dayjs"
import Button from '../../components/button';
import Logs, { createClipboardData, LogFooterButtons } from '../../components/logs/logs';

export default function ActivityLogs(props) {
    const { activity, namespace, setErrorMsg } = props

    const { data, err } = useMirrorLogs(Config.url, true, namespace, activity, localStorage.getItem("apikey"))
    const [follow, setFollow] = useState(true)

    return (
        <>
            <FlexBox className="col">
                <FlexBox style={{ backgroundColor: "#002240", color: "white", borderRadius: "8px 8px 0px 0px", overflow: "hidden", padding: "8px" }}>
                    <Logs logItems={data} wordWrap={true} autoScroll={follow} setAutoScroll={setFollow} overrideLoadingMsg={activity === null ? "No Activity Selected" : null}/>
                </FlexBox>
                <div style={{ height: "40px", backgroundColor: "#223848", color: "white", maxHeight: "40px", minHeight: "40px", padding: "0px 10px 0px 10px", boxShadow: "0px 0px 3px 0px #fcfdfe", alignItems: 'center', borderRadius: " 0px 0px 8px 8px", overflow: "hidden" }}>
                    <FlexBox className="gap" style={{ width: "100%", flexDirection: "row-reverse", height: "100%", alignItems: "center" }}>
                        <LogFooterButtons setFollow={setFollow} follow={follow} data={data} />
                    </FlexBox>
                </div>
            </FlexBox>
        </>
    )
}
