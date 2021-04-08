import React, {useEffect} from 'react'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'

import {TimeSinceUnix} from '../../util/utils'
import JSONPretty from 'react-json-pretty';


export default function InstanceDetails(props) {
    const {basicValue, fetchBasicWorkflow} = props
    // on mount start a poller to re-fetch workflow
    useEffect(() => {
        let timer = setInterval(() => {
            fetchBasicWorkflow()
        }, 2000)
        if (basicValue.status === "complete" || basicValue.status === "cancelled") {
            clearInterval(timer)
        }
        return function cleanup() {
            clearInterval(timer)
        }
    }, [fetchBasicWorkflow, basicValue.status])

    let timeSinceStarted = ""
    if (basicValue.beginTime && basicValue.beginTime.seconds) {
        timeSinceStarted = TimeSinceUnix(basicValue.beginTime.seconds)
    }

    function OutputComponent() {
        if (!basicValue.output) {
            return (<pre>pending...</pre>)
        }

        if (basicValue.output.length > 0) {
            return (<JSONPretty data={atob(basicValue.output)} onJSONPrettyError={e => console.error(e)}></JSONPretty>)
        }

        return (<pre>loading...</pre>)

    }

    function InputComponent() {
        if (!basicValue.input) {
            return (<pre>pending...</pre>)
        }

        if (basicValue.input.length > 0) {
            return (<JSONPretty data={atob(basicValue.input)} onJSONPrettyError={e => console.error(e)}></JSONPretty>)
        }

        return (<pre>loading...</pre>)

    }

    return (
        <Row style={{padding: "5px", marginTop: "10px"}}>
            <Col xs={12} style={{display: "inline-flex"}}>
                <div style={{display: "flex", justifyContent: "center", flexDirection: "column"}}>
                    <div>
                        <span style={{color: "#2396d8"}}>Instance:</span> {basicValue.id}
                    </div>
                    <div>
                        <span style={{color: "#2396d8"}}>Status:</span> {basicValue.status}
                    </div>
                    <div>
                        <span style={{color: "#2396d8"}}>Started:</span> {timeSinceStarted}
                    </div>
                </div>
            </Col>
            <Col xs={12} style={{marginTop: "20px"}}>
                <div>
                    <h4 style={{marginBottom: "0px", fontWeight: 300}}>Input:</h4> <br/>
                    <InputComponent/>
                </div>
                {!basicValue.output ?
                    "" :
                    <div>
                        <h4 style={{marginBottom: "0px", fontWeight: 300}}>Output:</h4> <br/>
                        <OutputComponent/>

                    </div>}
            </Col>
        </Row>
    )
}

