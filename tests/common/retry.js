import { test } from '@jest/globals'

import helpers from './helpers'

function runTest (handler) {
	return new Promise((resolve, reject) => {
		const result = handler(err => err ? reject(err) : resolve())

		if (result && result.then)
			result.catch(reject).then(resolve)
		else
			resolve()
	})
}


async function _retry (description, retries, handler, delay) {
	if (!description || typeof description !== 'string')
		throw new Error('Invalid argument, description must be a string')


	if (typeof retries === 'function' && !handler) {
		handler = retries
		retries = 1
	}

	if (!retries || typeof retries !== 'number' || retries < 1)
		throw new Error('Invalid argument, retries must be a greather than 0')


	test(description, async () => {
		let latestError
		for (let tries = 0; tries < retries; tries++)
			try {
				await helpers.sleep(delay)
				await runTest(handler)
				return
			} catch (error) {
				latestError = error
			}


		throw latestError
	})
}

export function retry10 (description, handler) {
	return _retry(description, 10, handler, 500)
}
export function retry50 (description, handler) {
	return _retry(description, 50, handler, 500)
}
