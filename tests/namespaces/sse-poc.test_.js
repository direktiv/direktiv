import { describe, expect, it } from '@jest/globals'

import common from '../common'

const API_HOST = 'http://ec2-3-231-218-167.compute-1.amazonaws.com'
const IS_ENTERPRISE = false
const AUTH_TOKEN = 'password'
const NAMESPACE = 'my-namespace'

describe('Test Server Sent Events', () => {
	it(`can handle Server Sent Events`, async () => {
		// configure the sse request and store the dispatch function and the message and error mock functions
		const { dispatch, onErrorMock, onMessageMock } = common.utils.direktivSSE({
			path: `${ API_HOST }/api/namespaces/${ NAMESPACE }/logs`,
			auth: {
				token: AUTH_TOKEN,
				isEnterprice: IS_ENTERPRISE,
			},
		})

		// call the dispatch function and wait for the first message or error
		await dispatch()

		// at this point the connection is already closed
		// we can now check the onErrorMock, onMessageMock
		// and test if and how they have been called
		expect(onErrorMock).not.toHaveBeenCalled()
		expect(onMessageMock).toHaveBeenCalledTimes(1)
		expect(onMessageMock).toHaveBeenCalledWith(
			expect.objectContaining({
				namespace: NAMESPACE,
			}),
		)

		// we can also retreive the exact data with which the mock function was called
		// we get the fist call and the fist argument with [0][0]
		const logResultsArr = onMessageMock.mock.calls[0][0]?.results
		expect(onMessageMock).toHaveBeenCalledWith(
			expect.objectContaining({
				namespace: NAMESPACE,
				pageInfo: null,
			}),
		)
	})

	it(`returns an unauthorized error when called with no credentials`, async () => {
		const { dispatch, onErrorMock, onMessageMock } = common.utils.direktivSSE({
			path: `${ API_HOST }/api/namespaces/${ NAMESPACE }/logs`,
		})
		await dispatch()
		expect(onErrorMock).toHaveBeenCalledWith(
			expect.objectContaining(common.structs.unauthorizedResponse),
		)
		expect(onMessageMock).not.toHaveBeenCalled()
	})
})
