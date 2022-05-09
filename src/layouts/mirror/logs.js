import React, { useEffect, useState } from 'react';
import './style.css';
import { AutoSizer, List, CellMeasurer, CellMeasurerCache } from 'react-virtualized';

import FlexBox from '../../components/flexbox';
import { VscCopy, VscEye, VscEyeClosed, } from 'react-icons/vsc';
import { Config, copyTextToClipboard, } from '../../util';
import { useMirrorLogs, } from 'direktiv-react-hooks';

import * as dayjs from "dayjs"

import { TerminalButton } from '../instance';

export default function ActivityLogs(props) {
    const { activity, namespace, setErrorMsg } = props

    const [clipData, setClipData] = useState(null)
    const [follow, setFollow] = useState(true)
    const [width,] = useState(window.innerWidth);




    return (
        <>
            <FlexBox className="col">
                <FlexBox style={{ backgroundColor: "#002240", color: "white", borderRadius: "8px 8px 0px 0px", overflow: "hidden", padding: "8px" }}>
                    <Logs clipData={clipData} setClipData={setClipData} follow={true} activityID={activity} namespace={namespace} setErrorMsg={setErrorMsg} />
                </FlexBox>
                <div style={{ height: "40px", backgroundColor: "#223848", color: "white", maxHeight: "40px", minHeight: "40px", padding: "0px 10px 0px 10px", boxShadow: "0px 0px 3px 0px #fcfdfe", alignItems: 'center', borderRadius: " 0px 0px 8px 8px", overflow: "hidden" }}>
                    <FlexBox className="gap" style={{ width: "100%", flexDirection: "row-reverse", height: "100%", alignItems: "center" }}>
                        <TerminalButton className={`${activity ? "" : "terminal-disabled"}`} onClick={() => {
                            copyTextToClipboard(clipData)
                        }}>
                            <VscCopy /> Copy {width > 999 ? <span>to Clipboard</span> : ""}
                        </TerminalButton>
                        {follow ?
                            <TerminalButton className={`${activity ? "" : "terminal-disabled"}`} onClick={(e) => setFollow(!follow)}>
                                <VscEyeClosed /> Stop {width > 999 ? <span>watching</span> : ""}
                            </TerminalButton>
                            :
                            <TerminalButton className={`${activity ? "" : "terminal-disabled"}`} onClick={(e) => setFollow(!follow)} >
                                <VscEye /> <div>Follow {width > 999 ? <span>logs</span> : ""}</div>
                            </TerminalButton>
                        }
                    </FlexBox>
                </div>
            </FlexBox>
        </>
    )
}

function Logs(props) {
    const { namespace, activityID, follow, setClipData, clipData, setErrorMsg } = props;

    const cache = new CellMeasurerCache({
        fixedWidth: true,
        fixedHeight: false
    })

    const [logLength, setLogLength] = useState(0)
    const { data, err } = useMirrorLogs(Config.url, true, namespace, activityID, localStorage.getItem("apikey"))

    useEffect(() => {
        if (err) {
            setErrorMsg(`Could not get logs: ${err}`)
        }
    }, [err, setErrorMsg])

    useEffect(() => {
        if (!setClipData) {
            // Skip ClipData if unset
            return
        }

        if (data !== null) {
            if (clipData === null || logLength === 0) {
                let cd = ""
                for (let i = 0; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`
                }
                setClipData(cd)
                setLogLength(data.length)
            } else if (data.length !== logLength) {
                let cd = clipData
                for (let i = logLength - 1; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`
                }
                setClipData(cd)
                setLogLength(data.length)
            }
        }
    }, [data, clipData, setClipData, logLength])

    if (!activityID) {
        return <div>No Activity Selected</div>
    }


    if (!data) {
        return <div>Loading...</div>
    }

    if (err) {
        return <></> // TODO 
    }

    function rowRenderer({ index, parent, key, style }) {
        if (!data[index]) {
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
                <div style={{ ...style, minWidth: "800px", width: "800px" }}>
                    <div style={{ display: "inline-block", minWidth: "112px", color: "#b5b5b5" }}>
                        <div className="log-timestamp">
                            <div>[</div>
                            <div style={{ display: "flex", flex: "auto", justifyContent: "center" }}>{dayjs.utc(data[index].node.t).local().format("HH:mm:ss.SSS")}</div>
                            <div>]</div>
                        </div>
                    </div>
                    <span style={{ marginLeft: "5px", whiteSpace: "pre-wrap" }}>
                        {data[index].node.msg}
                    </span>
                    <div style={{ height: `fit-content` }}></div>
                </div>
            </CellMeasurer>
        );
    }


    return (
        <div className="activity-logger" style={{ flex: "1 1 auto", lineHeight: "20px" }}>
            <AutoSizer>
                {({ height, width }) => (
                    <div style={{ height: "100%", minHeight: "100%" }}>
                        <List
                            width={width}
                            height={height}
                            rowRenderer={rowRenderer}
                            deferredMeasurementCache={cache}
                            scrollToIndex={follow ? data.length - 1 : 0}
                            rowCount={data.length}
                            rowHeight={cache.rowHeight}
                            scrollToAlignment={"start"}
                        />
                    </div>
                )}
            </AutoSizer>
        </div>
    )
}
