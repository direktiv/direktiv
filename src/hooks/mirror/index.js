import * as React from 'react'
import { HandleError, ExtractQueryString, useEventSourceCleaner, StateReducer, STATE, useQueryString, genericEventSourceErrorHandler, SanitizePath, apiKeyHeaders } from '../util'
const { EventSourcePolyfill } = require('event-source-polyfill')
const fetch = require('isomorphic-fetch')

export const useDirektivMirror = (url, stream, namespace, path, apikey, ...queryParameters) => {
    const [info, dispatchInfo] = React.useReducer(StateReducer, null)
    const [activities, dispatchActivities] = React.useReducer(StateReducer, null)
    const [err, setErr] = React.useState(null)

    const [eventSource, setEventSource] = React.useState(null)
    const { } = useEventSourceCleaner(eventSource);

    // Store Query parameters
    const { queryString } = useQueryString(true, queryParameters)
    const [pathString, setPathString] = React.useState(null)

    // Stores PageInfo about node list stream
    const [pageInfo, setPageInfo] = React.useState(null)
    const pageInfoRef = React.useRef(pageInfo)

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
                if (json?.activities) {
                    dispatchActivities({
                        type: STATE.UPDATE,
                        data: json.activities.results,
                    })

                    setPageInfo(json.activities.pageInfo)
                }

                if (json?.info) {
                    dispatchInfo({
                        type: STATE.UPDATE,
                        data: json.info,
                    })
                }
            }

            listener.onmessage = e => readData(e)
            setEventSource(listener)
        } else {
            setEventSource(null)
        }
    }, [stream, queryString, pathString, apikey])

    // Non Stream Data Dispatch Handler
    React.useEffect(async () => {
        if (!stream && pathString !== null && !err) {
            setEventSource(null)
            try {
                const data = await getInfo()
                dispatchInfo({ type: STATE.UPDATE, data: data.info })
                dispatchActivities({ type: STATE.UPDATE, data: data.activities.results })
            } catch (e) {
                setErr(e.onmessage)
            }
        }
    }, [stream, queryString, pathString, err])

    // Update PageInfo Ref
    React.useEffect(() => {
        pageInfoRef.current = pageInfo
    }, [pageInfo])

    // Reset states when any prop that affects path is changed
    React.useEffect(() => {
        if (stream) {
            setPageInfo(null)
            setPathString(url && namespace && path ? `${url}namespaces/${namespace}/tree${SanitizePath(path)}?op=mirror-info` : null)
        } else {
            dispatchInfo({ type: STATE.UPDATE, data: null })
            dispatchActivities({ type: STATE.UPDATE, data: null })
            setPathString(url && namespace && path ? `${url}namespaces/${namespace}/tree${SanitizePath(path)}?op=mirror-info` : null)
        }
    }, [stream, path, namespace, url])

    async function getInfo(...queryParameters) {
        let uriPath = `${url}namespaces/${namespace}/tree`
        if (path !== "") {
            uriPath += `${SanitizePath(path)}`
        }
        let request = {
            method: "GET",
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${uriPath}?op=mirror-info${ExtractQueryString(true, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('get mirror info', resp, 'mirrorInfo'))
        }

        return await resp.json()
    }


    async function updateSettings(mirrorSettings, ...queryParameters) {
        let uriPath = `${url}namespaces/${namespace}/tree`
        if (path !== "") {
            uriPath += `${SanitizePath(path)}`
        }

        let request = {
            method: "POST",
            body: JSON.stringify(mirrorSettings),
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${uriPath}?op=update-mirror${ExtractQueryString(true, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('updated mirror', resp, 'updateMirror'))
        }

        return
    }

    async function sync(force, ...queryParameters) {
        let uriPath = `${url}namespaces/${namespace}/tree`
        if (path !== "") {
            uriPath += `${SanitizePath(path)}`
        }

        let request = {
            method: "POST",
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${uriPath}?op=sync-mirror${force ? "&force=true" : ""}${ExtractQueryString(true, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('sync mirror', resp, 'syncMirror'))
        }

        return
    }

    async function setLock(lock, ...queryParameters) {
        let uriPath = `${url}namespaces/${namespace}/tree`
        if (path !== "") {
            uriPath += `${SanitizePath(path)}`
        }

        let request = {
            method: "POST",
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${uriPath}?op=${lock ? "lock-mirror" : "unlock-mirror"}${ExtractQueryString(true, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('lock mirror', resp, 'lockMirror'))
        }

        return
    }

    async function cancelActivity(activity, ...queryParameters) {
        let uriPath = `${url}namespaces/${namespace}/activities/${activity}/cancel`

        let request = {
            method: "POST",
            headers: apiKeyHeaders(apikey)
        }

        let resp = await fetch(`${uriPath}${ExtractQueryString(false, ...queryParameters)}`, request)
        if (!resp.ok) {
            throw new Error(await HandleError('cancel mirror', resp, 'cancelMirror'))
        }

        return
    }

    return {
        info,
        activities,
        err,
        pageInfo,
        getInfo,
        updateSettings,
        cancelActivity,
        setLock,
        sync
    }
}