import { jest } from '@jest/globals'
import EventSource from 'eventsource'

import config from './config'

export default {
	/**
     * This function returns a an object with a the following properties:
     *
     * dispatch: an async function to dispatch the sse request. It will return
     * a promise that resolves when the first message or error is received
     *
     *
     * onMessageMock: a jest mock function that will be called when a message
     * is received. The mock function will be called with the JSON parsed message
     *
     * onErrorMock: a jest mock function that will be called when an error is
     * received
     *
     * CURRENT LIMITATION:
     * this sse will only call the mock function when the first message or error is
     * received this can be changed in the future if needed. The general problem is,
     * that we don't want to block the jest process very long. We might add some
     * parameter to the direktivSSE configures how many messages we want to receive
     * before we close the connection again. We might also do something in between
     * two messages. We can extend this feature when we have the first very specific
     * use case.
     */
	direktivSSE: ({
		path, // f.e. http://localhost:8080/api/namespaces/my-namespace/logs
		auth = { token: null,
			isEnterprice: false }, // optional auth token
		headers = {}, // optional additional headers
	}) => {
		const { token = null, isEnterprice = false } = auth
		const sseListener = new EventSource(path, {
			headers: {
				...config.getAuthHeader(token, isEnterprice),
				...headers,
			},
		})

		const onMessageMock = jest.fn()
		const onErrorMock = jest.fn()

		const dispatch = async () =>
			await new Promise(resolve => {
				sseListener.onmessage = e => {
					onMessageMock(JSON.parse(e.data))
					sseListener.close()
					resolve()
				}

				sseListener.onerror = e => {
					onErrorMock(e)
					sseListener.close()
					resolve()
				}
			})

		return { dispatch,
			onErrorMock,
			onMessageMock }
	},
}
