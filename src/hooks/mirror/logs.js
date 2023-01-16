import * as React from 'react'
import { genericEventSourceErrorHandler, HandleError, STATE, StateReducer, useEventSourceCleaner, useQueryString, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require('isomorphic-fetch')

export const useDirektivMirrorLogs = (url, stream, namespace, activity, apikey, ...queryParameters) => {
    const [data, dataDispatch] = React.useReducer(StateReducer, null)
    const [err, setErr] = React.useState(null)

    const [eventSource, setEventSource] = React.useState(null)
    const { } = useEventSourceCleaner(eventSource);

    const { queryString } = useQueryString(false, queryParameters)
    const [pathString, setPathString] = React.useState(null)

    // Stores PageInfo about node list stream
    const [pageInfo, setPageInfo] = React.useState(null)

    // Stream Event Source Data Dispatch Handler
    React.useEffect(() => {
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
                if (json) {
                    dataDispatch({
                        type: STATE.APPENDLIST,
                        data: json.results,
                    })

                    setPageInfo(json.pageInfo)
                }
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
        } else {
            setEventSource(null)
        }
    }, [stream, queryString, pathString])

    // Non Stream Data Dispatch Handler
    React.useEffect(() => {
        if (!stream && pathString !== null) {
            setEventSource(null)
            getActivityLogs().then((data) => {
                dataDispatch({ type: STATE.UPDATE, data: data })
            })
        }
    }, [stream, queryString, pathString])

    // Reset states when any prop that affects path is changed
    React.useEffect(() => {
        if (stream) {
            setPageInfo(null)
            dataDispatch({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace && activity ? `${url}namespaces/${namespace}/activities/${activity}/logs` : null)
        } else {
            dataDispatch({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace && activity ? `${url}namespaces/${namespace}/activities/${activity}/logs` : null)
        }
    }, [stream, activity, namespace, url])


    async function getActivityLogs() {
        let request = {
            method: "GET",
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${pathString}${queryString}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('mirror activity logs', resp, 'mirrorActivityLogs'))
        }

        return await resp.json()
    }

    return {
        data,
        err,
        pageInfo,
        getActivityLogs,
    }
}