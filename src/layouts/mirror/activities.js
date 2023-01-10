import React, { useEffect, useState } from 'react';
import './style.css';

import FlexBox from '../../components/flexbox';
import { GenerateRandomKey } from '../../util';
import Button from '../../components/button';
import Loader from '../../components/loader';

import * as dayjs from "dayjs"
import { SuccessState, CancelledState, FailState, RunningState } from '../instances';


export default function ActivityTable(props) {
    const { activities, err, placeholder, namespace, setActivity, cancelActivity, setErrorMsg } = props
    const [load, setLoad] = useState(true)

    useEffect(() => {
        if (activities !== null || err !== null) {
            setLoad(false)
        }
    }, [activities, err])

    return (
        <Loader load={load} timer={3000}>
            {activities ? <>{
                activities !== null && activities.length === 0 ?
                    <div style={{ paddingLeft: "10px", fontSize: "10pt" }}>{`${placeholder ? placeholder : "No activities have been recently executed. Recent activities will appear here."}`}</div>
                    :
                    <table className="instances-table" style={{ width: "100%" }}>
                        <thead>
                            <tr>
                                <th className="center-align" style={{ maxWidth: "120px", minWidth: "120px", width: "120px" }}>State</th>
                                <th className="center-align">Type</th>
                                <th className="center-align">Started <span className="hide-1000">at</span></th>
                                <th className="center-align" style={{ maxWidth: "120px", minWidth: "120px", width: "120px" }}></th>
                            </tr>
                        </thead>
                        <tbody>
                            {activities !== null ?
                                <>
                                    <>
                                        {activities.map((obj) => {
                                            return (
                                                <ActivityRow
                                                    key={GenerateRandomKey()}
                                                    namespace={namespace}
                                                    state={obj.status}
                                                    id={obj.id}
                                                    type={obj.type}
                                                    startedDate={dayjs.utc(obj.createdAt).local().format("DD MMM YY")}
                                                    startedTime={dayjs.utc(obj.createdAt).local().format("HH:mm a")}
                                                    finishedDate={dayjs.utc(obj.updatedAt).local().format("DD MMM YY")}
                                                    finishedTime={dayjs.utc(obj.updatedAt).local().format("HH:mm a")}
                                                    setActivity={setActivity}
                                                    cancelActivity={cancelActivity}
                                                    setErrorMsg={setErrorMsg}
                                                />
                                            )
                                        })}</>
                                </>
                                : ""}
                        </tbody>
                    </table>
            }</> : <></>}
        </Loader>
    );
}

const success = "complete";
const fail = "failed";
const crashed = "crashed";
const cancelled = "cancelled";
const running = "pending";

export function ActivityRow(props) {
    let { state, startedDate, startedTime, id, type, setActivity, cancelActivity, setErrorMsg } = props;

    let label;
    if (state === success) {
        label = <SuccessState />
    } else if (state === cancelled) {
        label = <CancelledState />
    } else if (state === fail || state === crashed) {
        label = <FailState />
    } else if (state === running) {
        label = <RunningState />
    }

    return (

        <tr className="activity-row" style={{ minHeight: "48px", maxHeight: "48px" }}>
            <td className="label-cell">
                {label}
            </td>
            <td className="center-align">
                {type}
            </td>
            <td className="center-align">
                <span className="hide-864">{startedDate}, </span>
                {startedTime}
            </td>
            <td className="center-align">
                <FlexBox className="center gap">
                    <Button color="info" variant="outlined" className={`small light`} style={state !== "pending" ? { visibility: "hidden" } : {}} onClick={async () => {
                        try {
                            await cancelActivity(id)
                        } catch (e) {
                            setErrorMsg(`Failed to cancel: ${e.message}`)
                        }

                    }}>
                        Cancel
                    </Button>
                    <Button onClick={async () => {
                        setActivity(id)
                    }}>
                        Logs
                    </Button>
                </FlexBox>
            </td>
        </tr>
    )
}