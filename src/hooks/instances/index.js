import * as React from 'react'
import { HandleError, ExtractQueryString, StateReducer, STATE, useEventSourceCleaner, useQueryString, genericEventSourceErrorHandler, apiKeyHeaders } from '../util'

// For testing
// import fetch from "cross-fetch"
// In Production
const fetch = require('isomorphic-fetch')
const { EventSourcePolyfill } = require('event-source-polyfill')
/*
    useInstances is a react hook which returns a list of instances
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - apikey to provide authentication of an apikey
*/
export const useDirektivInstances = (url, stream, namespace, apikey, ...queryParameters) => {
    const [data, dispatchData] = React.useReducer(StateReducer, null)
    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)
    const { eventSourceRef } = useEventSourceCleaner(eventSource, "useInstances");

    // Store Query parameters
    const { queryString } = useQueryString(false, queryParameters)
    const [pathString, setPathString] = React.useState(null)

    // Stores PageInfo about instances list stream
    const [pageInfo, setPageInfo] = React.useState(null)

    // Stream Event Source Data Dispatch Handler
    React.useEffect(() => {
        const handler = setTimeout(() => {
            if (stream && pathString !== null) {
                // setup event listener 
                let listener = new EventSourcePolyfill(`${pathString}${queryString}`, {
                    headers: apiKeyHeaders(apikey)
                })

                listener.onerror = (e) => { genericEventSourceErrorHandler(e, setErr) }

                async function readData(e) {
                    if (e.data === "") {
                        return
                    }

                    let json = JSON.parse(e.data)
                    dispatchData({
                        type: STATE.UPDATE,
                        data: json.instances.results,
                    })

                    setPageInfo(json.instances.pageInfo)
                }

                listener.onmessage = e => readData(e)
                setEventSource(listener)
            } else {
                setEventSource(null)
            }
        }, 50);

        return () => {
            clearTimeout(handler);
        };
    }, [stream, queryString, pathString, apikey])

    // Non Stream Data Dispatch Handler
    React.useEffect(async () => {
        if (!stream && pathString !== null && !err) {
            setEventSource(null)
            try {
                const instancesData = await getInstances()
                dispatchData({ type: STATE.UPDATE, data: instancesData })
            } catch (e) {
                setErr(e)
            }
        }
    }, [stream, queryString, pathString, err])

    // Reset states when any prop that affects path is changed
    React.useEffect(() => {
        if (stream) {
            setPageInfo(null)
            setPathString(url && namespace ? `${url}namespaces/${namespace}/instances` : null)
        } else {
            dispatchData({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace ? `${url}namespaces/${namespace}/instances` : null)
        }
    }, [stream, namespace, url])

    // getInstances returns a list of instances
    async function getInstances(...queryParameters) {
        // fetch instance list by default
        let resp = await fetch(`${url}namespaces/${namespace}/instances${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error((await HandleError('list instances', resp, "listInstances")))
        }

        let json = await resp.json()
        setPageInfo(json.instances.pageInfo)
        return json.instances.results
    }

    return {
        data,
        err,
        pageInfo,
        getInstances
    }
}