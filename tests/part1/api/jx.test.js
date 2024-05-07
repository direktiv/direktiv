import { describe, expect, it } from '@jest/globals'
import { encode, decode } from 'js-base64'

import config from '../../common/config'
import request from '../../common/request'

describe('Test the jx API.', () => {
	it(`should perform a simple jx query with a valid body`, async () => {
        var pl = {
            jx: encode('{}'),
            data: encode('{}'),
        }

		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode('{}')],
		})
	})

	it(`should perform a simple jx string query`, async () => {
        var pl = {
            jx: encode('"jq(5)"'),
            data: encode('{}'),
        }

		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode('5')],
		})
	})

	it(`should perform a broken jx string query`, async () => {
        var pl = {
            jx: encode('"jq("'),
            data: encode('{}'),
        }

		const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

		expect(r.statusCode).toEqual(200)
		expect(r.body.data).toEqual({
			jx: pl.jx,
            data: pl.data,
            logs: 'ZmFpbHVyZToganEvanMgc2NyaXB0IG1pc3NpbmcgYnJhY2tldApxdWVyeSBwcm9kdWNlZCB6ZXJvIHJlc3VsdHMK',
            output: null,
		})
	})

    it(`should perform a broken jx javascript query`, async () => {
        var pl = {
            jx: encode('"js(5)"'),
            data: encode('{}'),
        }

        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: pl.jx,
            data: pl.data,
            logs: 'ZmFpbHVyZTogZXJyb3IgaW4ganMgcXVlcnkgNTogbm8gcmVzdWx0cwpxdWVyeSBwcm9kdWNlZCB6ZXJvIHJlc3VsdHMK',
            output: null,
        })
    })

    it(`should perform a simple jx javascript query`, async () => {
        var pl = {
            jx: encode('"js(return 5)"'),
            data: encode('{}'),
        }

        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode('5')],
        })
    })

    it(`should perform a structured jx query`, async () => {
        var pl = {
            jx: encode('{"x": "jq(5)"}'),
            data: encode('{}'),
        }

        const r = await request(config.getDirektivHost()).post(`/api/v2/jx`).send(pl)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode(`{
  "x": 5
}`)],
        })
    })

    it(`should perform a simple jx query with passing assertions`, async () => {
        var pl = {
            jx: encode('{"x": "jq(5)"}'),
            data: encode('{}'),
        }

        const r = await request(config.getDirektivHost()).post(`/api/v2/jx?assert=object&assert=success`).send(pl)

        expect(r.statusCode).toEqual(200)
        expect(r.body.data).toEqual({
            jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode(`{
  "x": 5
}`)],
        })
    })

    it(`should perform a simple jx query with failing assertions`, async () => {
        var pl = {
            jx: encode('"jq(5)"'),
            data: encode('{}'),
        }

		const r = await request(config.getDirektivHost()).post(`/api/v2/jx?assert=object`).send(pl)

		expect(r.statusCode).toEqual(400)
        expect(r.body.error).toEqual({
			code: "assert_object",
            message: "result is not an object"
		})
		expect(r.body.data).toEqual({
			jx: pl.jx,
            data: pl.data,
            logs: '',
            output: [encode('5')],
		})
    })
})
