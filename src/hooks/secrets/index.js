import * as React from 'react'
import { HandleError, ExtractQueryString, apiKeyHeaders } from '../util'
const fetch = require('isomorphic-fetch')

/*
    useSecrets is a react hook which returns create registry, delete registry and data
    takes:
      - url to direktiv api http://x/api/
      - namespace the namespace to query on
      - apikey to provide authentication of an apikey
*/
export const useDirektivSecrets = (url, namespace, apikey) => {
    const [data, setData] = React.useState(null)

    React.useEffect(() => {
        const getData = async () => getSecrets();
        if (data === null) {
          getData().catch(() => {
          });
        }
    }, [data])

    // getSecrets returns a list of registries
    async function getSecrets(...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/secrets${ExtractQueryString(false, ...queryParameters)}`, {
            headers: apiKeyHeaders(apikey)
        })
        if (resp.ok) {
            let json = await resp.json()
            setData(json.secrets.results)
            return json.secrets.results
        } else {
            throw new Error(await HandleError('list secrets', resp, 'listSecrets'))
        }
    }

    async function createSecret(name, value, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/secrets/${name}${ExtractQueryString(false, ...queryParameters)}`, {
            method: "PUT",
            body: value,
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('create secret', resp, 'createSecret'))
        }
    }

    async function deleteSecret(name, ...queryParameters) {
        let resp = await fetch(`${url}namespaces/${namespace}/secrets/${name}${ExtractQueryString(false, ...queryParameters)}`, {
            method: "DELETE",
            headers: apiKeyHeaders(apikey)
        })
        if (!resp.ok) {
            throw new Error(await HandleError('delete secret', resp, 'deleteSecret'))
        }
    }


    return {
        data,
        createSecret,
        deleteSecret,
        getSecrets
    }

}