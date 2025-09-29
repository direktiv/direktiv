export default {
	getDirektivBaseUrl () {
		if (process.env.DIREKTIV_BASE_URL)
			return process.env.DIREKTIV_BASE_URL

		return 'http://127.0.0.1:8080'
	},
	getAuthHeader: (authToken, isEnterprice = false) => {
		if (!authToken)
			return {}

		return isEnterprice
			? { Authorization: `Bearer ${ authToken }` }
			: { 'Direktiv-Api-Key': authToken }
	},
}
