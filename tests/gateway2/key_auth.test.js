import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test gateway key-auth plugin', () => {
    beforeAll(helpers.deleteAllNamespaces)
    helpers.itShouldCreateNamespace(it, expect, namespace)

    helpers.itShouldCreateYamlFile(it, expect, namespace,
        '/', 'c1.yaml', 'consumer', `
direktiv_api: "consumer/v1"
username: user1
password: pwd1
api_key: key1
tags:
- tag1
groups:
- group1
`)


helpers.itShouldCreateYamlFile(it, expect, namespace,
    '/', 'ep1.yaml', 'endpoint', `
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/foo"
    allow_anonymous: false
    plugins:
        target:
            type: debug-target
        auth:
        - type: key-auth   
          configuration:
                add_username_header: true
                add_tags_header: true
                add_groups_header: true
post:
    responses:
        "200":
        description: works
`
)

    retry10(`should denied ep1.yaml endpoint`, async () => {
        const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
            .set('API-Token', 'wrong_key')
            .send({})
        expect(res.statusCode).toEqual(403)
        expect(res.body).toEqual({
            error: {
                endpointFile: '/ep1.yaml',
                message: 'authentication failed',
            },
        })
    })

    retry10(`should access ep1.yaml endpoint`, async () => {
        const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
            .set('API-Token', 'key1')
            .send({ foo: 'bar' })
        expect(res.statusCode).toEqual(200)
        expect(res.body.data.headers).toMatchObject({
            'Direktiv-Consumer-Groups': [ 'group1' ],
            'Direktiv-Consumer-Tags': [ 'tag1' ],
            'Direktiv-Consumer-User': [ 'user1' ],
            'Accept-Encoding': [ 'gzip, deflate' ],
            'Api-Token': [ 'key1' ],
            'Content-Length': [ '13' ],
            'Content-Type': [ 'application/json' ],
        })
        expect(res.body.data.text).toEqual('from debug plugin')
        expect(res.body.data.body).toEqual('{"foo":"bar"}')
    })
})
