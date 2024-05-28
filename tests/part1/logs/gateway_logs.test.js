import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(__filename)

describe('Test gateway api calls', () => {

	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

    it(`create endpoint`, async () => {
        const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/files/`)
        .set('Content-Type', 'application/json')
        .send(`{"name":"route.yaml","type":"endpoint","mimeType":"application/yaml","data":"ZGlyZWt0aXZfYXBpOiBlbmRwb2ludC92MQo="}`)
        expect(res.statusCode).toEqual(200)

        const resPatch = await request(common.config.getDirektivHost()).patch(`/api/v2/namespaces/${ namespace }/files/route.yaml`)
        .set('Content-Type', 'application/json')
        .send(`{"data":"ZGlyZWt0aXZfYXBpOiBlbmRwb2ludC92MQphbGxvd19hbm9ueW1vdXM6IHRydWUKcGF0aDogL3Rlc3QKbWV0aG9kczoKICAtIEdFVApwbHVnaW5zOgogIHRhcmdldDoKICAgIHR5cGU6IGluc3RhbnQtcmVzcG9uc2UKICAgIGNvbmZpZ3VyYXRpb246CiAgICAgIHN0YXR1c19jb2RlOiAyMDAKICAgICAgc3RhdHVzX21lc3NhZ2U6IEhlbGxvCiAgaW5ib3VuZDogW10KICBvdXRib3VuZDogW10KICBhdXRoOiBbXQo="}`)
        expect(resPatch.statusCode).toEqual(200)

    })

    retry50(`create namespace`, async () => {
        const gwRes = await request(common.config.getDirektivHost()).get(`/ns/${ namespace }/test`)
        expect(gwRes.statusCode).toEqual(200)

        const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?route=%2Ftest`)
        expect(logRes.statusCode).toEqual(200)

        console.log(logRes.body.data)
        // TODO: check logs, we just disabled the info logging for the plugins

    })
    
})
