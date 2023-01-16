import * as React from 'react'
import { CloseEventSource, ExtractQueryString, HandleError, TrimPathSlashes, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require('isomorphic-fetch')

/*
    useInstanceLogs is a react hook which returns details for an instance
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - instance the id used for the instance
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstanceLogs = (url, stream, namespace, instance, apikey, ...queryParameters) => {
    const [data, setData] = React.useState(null)
    const logsRef = React.useRef([])

    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)

    // Store Query parameters
    const [queryString, setQueryString] = React.useState(ExtractQueryString(false, ...queryParameters))

    // Stores PageInfo about instance log stream
    const [pageInfo, setPageInfo] = React.useState(null)

    React.useEffect(() => {
        if (stream) {
            let log = logsRef.current
            if (eventSource === null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${url}namespaces/${namespace}/instances/${instance}/logs${queryString}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => {
                    if (e.status === 403) {
                        setErr("permission denied")
                    } else if (e.status === 404) {
                        setErr(e.statusText)
                    }
                }

                async function readData(e) {
                    if (e.data === "") {
                        return
                    }
                    let json = JSON.parse(e.data)
                    for (let i = 0; i < json.results.length; i++) {
                        log.push(json.results[i])
                    }
                    logsRef.current = log
                    setData(JSON.parse(JSON.stringify(logsRef.current)))
                    setPageInfo(json.pageInfo)
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            }
        } else {
            if (data === null) {
                getInstanceLogs()
            }
        }
    }, [data])

    React.useEffect(() => {
        return () => CloseEventSource(eventSource)
    }, [eventSource])


    // If queryParameters change and streaming: update queryString, and reset sse connection
    React.useEffect(() => {
        if (stream) {
            let newQueryString = ExtractQueryString(false, ...queryParameters)
            if (newQueryString !== queryString) {
                setQueryString(newQueryString)
                CloseEventSource(eventSource)
                setEventSource(null)
            }
        }
    }, [eventSource, queryParameters, queryString, stream])

    // getInstanceLogs returns a list of logs
    async function getInstanceLogs(...queryParameters) {
        // fetch instance list by default
        let resp = await fetch(`${url}namespaces/${namespace}/instances/${instance}/logs${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.results)
            setPageInfo(json.pageInfo)
            return json.results
        }

        throw new Error(await HandleError('get instance logs', resp, "instanceLogs"))

    }

    return {
        data,
        err,
        pageInfo,
        getInstanceLogs
    }
}
/*
    useInstance is a react hook which returns details for an instance
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - instance the id used for the instance
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstance = (url, stream, namespace, instance, apikey) => {
    const [data, setData] = React.useState(null)
    const [latestRevision, setLatestRevision] = React.useState(null)
    const [workflow, setWorkflow] = React.useState(null)
    const [err, setErr] = React.useState(null)
    const [instanceID, setInstanceID] = React.useState(instance)
    const [eventSource, setEventSource] = React.useState(null)


    React.useEffect(() => {
        if (stream) {
            if (eventSource === null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${url}namespaces/${namespace}/instances/${instanceID}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => {
                    if (e.status === 403) {
                        setErr("permission denied")
                    } else if (e.status === 404) {
                        setErr(e.statusText)
                    } else {
                        try {
                            let json = JSON.parse(e.data)
                            setErr(json.Message)
                        } catch (e) {
                            // TODO
                        }
                    }
                }

                async function readData(e) {
                    if (e.data === "") {
                        return
                    }
                    let json = JSON.parse(e.data)
                    json.instance["flow"] = json.flow
                    setData(json.instance)
                    setWorkflow(json.workflow)
                    getLatestRevision(json.workflow.path)
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            }
        } else {
            if (data === null) {
                getInstance()
            }
        }
    }, [data, eventSource])

    // If instance changes reset eventSource
    React.useEffect(() => {
        if (stream) {
            if (instance !== instanceID) {
                setInstanceID(instance)
                CloseEventSource(eventSource)
                setEventSource(null)
                setData(null)
            }
        }
    }, [eventSource, instanceID, instance, stream])

    React.useEffect(() => {
        return () => CloseEventSource(eventSource)
    }, [eventSource])

    // getInstance returns a list of instances
    async function getInstance(...queryParameters) {
        // fetch instance list by default
        let resp = await fetch(`${url}namespaces/${namespace}/instances/${instanceID}${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.instance)
            setWorkflow(json.workflow)
            getLatestRevision(json.workflow.path)
            return json.instance
        }
        throw new Error(await HandleError('get instance', resp, "getInstance"))

    }

    async function getLatestRevision(workflowPath, ...queryParameters) {
        // workflow doesnt exist anymore
        if (workflowPath === "") {
            setLatestRevision("")
        }

        let path = TrimPathSlashes(workflowPath)
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=validate-ref&ref=latest${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setLatestRevision(json.revision.name)
            return json.revision.name
        }
        throw new Error(await HandleError('get instance wf details', resp, "getInstance"))

    }

    async function getInput(...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/instances/${instanceID}/input${ExtractQueryString(false, ...queryParameters)}`, {
            method: "GET",
            headers: apiKeyHeaders(apikey)

        })
        if (resp.ok) {
            let json = await resp.json()
            return atob(json.data)
        }
        throw new Error(await HandleError('get instance input', resp, 'getInstance'))
    }

    async function getOutput(...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/instances/${instanceID}/output${ExtractQueryString(false, ...queryParameters)}`, {
            method: "GET",
            headers: apiKeyHeaders(apikey)

        })
        if (resp.ok) {
            let json = await resp.json()
            return atob(json.data)
        }
        throw new Error(await HandleError('get instance output', resp, 'getInstance'))

    }

    async function cancelInstance(...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/instances/${instanceID}/cancel${ExtractQueryString(false, ...queryParameters)}`, {
            method: "POST",
            headers: apiKeyHeaders(apikey)

        })
        if (!resp.ok) {
            throw new Error(await HandleError('cancelling instance', resp, "cancelInstance"))
        }
    }

    return {
        data,
        workflow,
        latestRevision,
        err,
        getInstance,
        cancelInstance,
        getInput,
        getOutput
    }
}