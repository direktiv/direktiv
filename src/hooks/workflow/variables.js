import * as React from 'react'
import { useEventSourceCleaner, HandleError, ExtractQueryString, StateReducer, useQueryString, STATE, genericEventSourceErrorHandler, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require("isomorphic-fetch")

/*
    useWorkflowVariables is a react hook
    takes:
      - url to direktiv api http://x/api/
      - stream to use sse or a normal fetch
      - namespace the namespace to send the requests to
      - path to the workflow you want to change
      - apikey to provide authentication of an apikey
*/
export const useDirektivWorkflowVariables = (url, stream, namespace, path, apikey, ...queryParameters) => {

    const [data, dispatchData] = React.useReducer(StateReducer, null)
    const [err, setErr] = React.useState(null)
    const [eventSource, setEventSource] = React.useState(null)
    const { eventSourceRef } = useEventSourceCleaner(eventSource);

    // Store Query parameters
    const { queryString } = useQueryString(true, queryParameters)
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
                    dispatchData({
                        type: STATE.UPDATE,
                        data: json.variables.results,
                    })

                    setPageInfo(json.variables.pageInfo)
                }
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
        } else {
            setEventSource(null)
        }
    }, [stream, apikey, queryString, pathString])


    // Non Stream Data Dispatch Handler
    React.useEffect(() => {
        if (!stream && pathString !== null && !err) {
            setEventSource(null)

            fetch(`${pathString}${queryString}`, {
                headers: apiKeyHeaders(apikey)
            }).then((resp)=>{
                resp.json().then((data) =>{
                    dispatchData({ type: STATE.UPDATE, data: data })
                })
            }).catch((e) =>{
                setErr(e.onmessage)
            })
        }
    }, [stream, queryString, pathString, err, apikey])

    // Reset states when any prop that affects path is changed
    React.useEffect(() => {
        if (stream) {
            setPageInfo(null)
            dispatchData({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace ? `${url}namespaces/${namespace}/tree/${path}/?op=vars` : null)
        } else {
            dispatchData({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace ? `${url}namespaces/${namespace}/tree/${path}/?op=vars` : null)
        }
    }, [stream, namespace, path, url])


    async function getWorkflowVariables(...queryParameters) {
        let uri = `${url}namespaces/${namespace}/tree/${path}?op=vars${ExtractQueryString(true, ...queryParameters)}`
        let resp = await fetch(`${uri}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            return json.variables.results
        } else {
            throw new Error(await HandleError('get node', resp, 'listNodes'))
        }
    }

    async function setWorkflowVariable(name, val, mimeType, ...queryParameters) {
        if (mimeType === undefined) {
            mimeType = "application/json"
        }
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=set-var&var=${name}${ExtractQueryString(true, ...queryParameters)}`, {
            method: "PUT",
            body: val,
            headers: {
                "Content-type": mimeType,
                ...apiKeyHeaders(apikey)
            },
        })
        if (!resp.ok) {
            throw new Error(await HandleError('set variable', resp, 'setWorkflowVariable'))
        }

        return await resp.json()
    }

    async function getWorkflowVariable(name, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=var&var=${name}${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            return { data: await resp.text(), contentType: resp.headers.get("Content-Type") }
        } else {
            throw new Error(await HandleError('get variable', resp, 'getWorkflowVariable'))
        }
    }

    async function getWorkflowVariableBuffer(name, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=var&var=${name}${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            return { data: await resp.arrayBuffer(), contentType: resp.headers.get("Content-Type") }
        } else {
            throw new Error(await HandleError('get variable', resp, 'getWorkflowVariable'))
        }
    }

    async function getWorkflowVariableBlob(name, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=var&var=${name}${ExtractQueryString(true, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            return { data: await resp.blob(), contentType: resp.headers.get("Content-Type") }
        } else {
            throw new Error(await HandleError('get variable', resp, 'getWorkflowVariable'))
        }
    }

    async function deleteWorkflowVariable(name, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/tree/${path}?op=delete-var&var=${name}${ExtractQueryString(true, ...queryParameters)}`, {
            method: "DELETE",
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('delete variable', resp, 'deleteWorkflowVariable'))
        }

    }

    return {
        data,
        err,
        pageInfo,
        getWorkflowVariables,
        setWorkflowVariable,
        deleteWorkflowVariable,
        getWorkflowVariable,
        getWorkflowVariableBuffer,
        getWorkflowVariableBlob
    }
}