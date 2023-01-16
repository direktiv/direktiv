import * as React from 'react'
import fetch from "cross-fetch"
import { CloseEventSource, HandleError, ExtractQueryString, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')

/*
    useWorkflowLogs is a react hook which returns data, err or getWorkflowLogs()
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace to call the api on
      - path for the workflow
      - apikey to provide authentication of an apikey
*/
export const useDirektivWorkflowLogs = (url, stream, namespace, path, apikey, ...queryParameters) => {

    const [data, setData] = React.useState(null)
    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)

    // Store Query parameters
    const [queryString, setQueryString] = React.useState(ExtractQueryString(true, ...queryParameters))

    // Stores PageInfo about workflow log stream
    const [pageInfo, setPageInfo] = React.useState(null)

    React.useEffect(() => {
        if (stream) {
            if (eventSource === null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${url}namespaces/${namespace}/tree/${path}?op=logs${queryString}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => {
                    if (e.status === 404) {
                        setErr(e.statusText)
                    } else if (e.status === 403) {
                        setErr("permission denied")
                    }
                }

                async function readData(e) {
                    if (e.data === "") {
                        return
                    }
                    let json = JSON.parse(e.data)
                    setData(json.results)
                    setPageInfo(json.pageInfo)
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            }
        } else {
            if (data === null) {
                getWorkflowLogs()
            }
        }
    }, [data])

    React.useEffect(() => {
        return () => {
            CloseEventSource(eventSource)
        }
    }, [eventSource])

    // If queryParameters change and streaming: update queryString, and reset sse connection
    React.useEffect(() => {
        if (stream) {
            let newQueryString = ExtractQueryString(true, ...queryParameters)
            if (newQueryString !== queryString) {
                setQueryString(newQueryString)
                CloseEventSource(eventSource)
                setEventSource(null)
            }
        }
    }, [eventSource, queryParameters, queryString, stream])

    // getWorkflowLogs returns a list of workflow logs
    async function getWorkflowLogs(...queryParameters) {
        // fetch namespace list by default
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=logs${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.results)
            setPageInfo(json.pageInfo)
            return json.results
        } else {
            throw new Error(await HandleError('list namespace logs', resp, 'namespaceLogs'))
        }
    }


    return {
        data,
        err,
        pageInfo,
        getWorkflowLogs
    }
}