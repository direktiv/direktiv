import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(__filename)

describe('Test sync api calls', () => {

	beforeAll(helpers.deleteAllNamespaces)


    it(`create namespace`, async () => {
        const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces`)
        .set('Content-Type', 'application/json')
        .send(`{"name":"${ namespace }","mirror":{"authType":"public","gitRef":"main","url":"https://github.com/direktiv/direktiv","insecure":false}}`)
        expect(res.statusCode).toEqual(200)

        const resSync = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/syncs`)
        expect(resSync.statusCode).toEqual(200)
    })

    retry50(`get sync logs`, async () => {
        const res = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/syncs`)
        expect(res.statusCode).toEqual(200)
        
        const resLog = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?activity=${ res.body.data[0].id }`)
        expect(resLog.statusCode).toEqual(200)

        console.log(resLog.body)
        expect(resLog.body.data.length).toBeGreaterThan(1)
    })
    
})
