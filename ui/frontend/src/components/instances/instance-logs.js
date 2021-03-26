import ServerContext from "components/app/context"
import {Alert, Col, Container, Row} from "react-bootstrap"
import {XCircle} from "react-bootstrap-icons"
import * as dayjs from "dayjs"
import {useCallback, useContext, useEffect, useState} from "react"

export function InstanceLogs(props) {
    const {params} = props
    const context = useContext(ServerContext)
    const [logs, setLogs] = useState([])
    const [logsOffset, setLogsOffset] = useState(0)
    const [err, setErr] = useState("")
    const [scrolled, setScrolled] = useState(false)
    const [init, setInit] = useState(false)

    // Log limits
    const [limit, setLimit] = useState(300)

    const clearError = () => {
        setErr("")
    }

    function checkIfScrollAtBottom(event) {
        if (event.target.offsetHeight + event.target.scrollTop === event.target.scrollHeight) {
            return true
        }
    }

    let checkScrollDirection = useCallback((event) => {
        if (checkScrollDirectionIsUp(event)) {
            setScrolled(true)
        } else if (checkIfScrollAtBottom(event)) {
            setScrolled(false)
        }
    }, [])

    function checkScrollDirectionIsUp(event) {
        if (event.wheelDelta) {
            return event.wheelDelta > 0;
        }
        return event.deltaY < 0;
    }


    let fetchLogs = useCallback(() => {
        async function fetchLogs() {
            try {
                let resp = await context.Fetch(`/instances/${params.namespace}/${params.workflow}/${params.id}/logs?offset=${logsOffset}&limit=${limit}`, {
                    method: "GET",
                })
                if (!resp.ok) {
                    let text = await resp.text()
                    throw (new Error(`Error fetching logs: ${text}`))
                } else {
                    let json = await resp.json()
                    if (json.workflowInstanceLogs && json.workflowInstanceLogs.length > 0) {
                        if (limit > 100) {
                            setLimit(100)
                        }

                        setLogsOffset(offset => {
                            offset += json.workflowInstanceLogs.length
                            return offset
                        })

                        setLogs([...logs, ...json.workflowInstanceLogs])
                    } else if (limit < 300) {
                        setLimit((l) => {
                            if (l === 10) {
                                return l
                            }

                            l -= 10
                            if (l < 10) {
                                l = 10
                            }

                            return l
                        })
                    }

                    if (!scrolled) {
                        if (document.getElementById('logs')) {
                            document.getElementById('logs').scrollTop = document.getElementById('logs').scrollHeight
                        }
                    }
                }
                setErr("")
            } catch (e) {
                setErr(e.message)
            }
        }

        return fetchLogs()

    }, [params, context.Fetch, scrolled, logs, logsOffset, limit])

    useEffect(() => {
        if (!init) {
            fetchLogs()
            setInit(true)
        } else {
            let timer = setInterval(fetchLogs, 2500)
            return function cleanup() {
                clearInterval(timer)
            }
        }

    }, [fetchLogs, init])

    useEffect(() => {
        let scrollableElement = document.getElementById('logs')
        scrollableElement.addEventListener('wheel', checkScrollDirection);
        return function cleanup() {
            scrollableElement.removeEventListener("wheel", checkScrollDirection)
        }
    }, [checkScrollDirection])


    return (
        <>
            {err !== "" ?
                <Row style={{margin: "0px"}}>
                    <Col style={{marginBottom: "15px"}} xs={12} md={12}>
                        <Alert variant="danger">
                            <Container>
                                <Row>
                                    <Col sm={11}>
                                        Fetching Logs: {err}
                                    </Col>
                                    <Col sm={1} style={{textAlign: "right", paddingRight: "0"}}>
                                        <XCircle style={{cursor: "pointer", fontSize: "large"}} onClick={() => {
                                            clearError()
                                        }}/>
                                    </Col>
                                </Row>
                            </Container>
                        </Alert>
                    </Col>
                </Row>
                :
                <div style={{
                    width: "100%",
                    minHeight: "600px",
                    background: 'black',
                    color: "#b5b5b5",
                    overflowY: "auto"
                }}>
                    <pre id="logs" style={{
                        backgroundColor: "transparent",
                        height: "600px",
                        color: "white",
                        fontFamily: "monospace",
                        fontSize: "14px",
                        overflow: 'auto',
                        margin: "0px",
                        inlineSize: "max-content",
                        width: "100%"
                    }}>
                        {logs.map((obj, i) => {
                            let time = dayjs.unix(obj.timestamp.seconds).format("h:mm:ss")
                            return (
                                <div key={obj.timestamp.seconds + i}>
                                    <span style={{color: "#b5b5b5"}}>
                                        [{time}]
                                        </span>
                                    {" "}
                                    {/* LOG LEVEL DISABLED, as it currently not supported */}
                                    {/* <span style={{ color: "#b5b5b5" }}>
                                        [<span style={{ color: COLORARR[obj.level] }}>
                                            {LVLARR[obj.level]}
                                        </span>]</span> */}
                                    {obj.message}
                                    {obj.context && obj.context.constructor === Object && Object.keys(obj.context).length > 0 ?
                                        <span style={{color: "#b5b5b5"}}>
                                            {"  ( "}
                                            <span style={{color: "#b5b5b5"}}>
                                                {Object.keys(obj.context).map((k) => {
                                                    return (
                                                        `${k}=${obj.context[k]} `
                                                    )
                                                })}
                                            </span>
                                            {")"}
                                        </span>
                                        :
                                        <></>
                                    }
                                </div>
                            )
                        })}


                    </pre>
                </div>
            }
        </>

    )
}

export default InstanceLogs