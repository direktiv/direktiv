export default {
	getDirektivHost () {
		if (process.env.DIREKTIV_HOST)
			return process.env.DIREKTIV_HOST

		return 'http://localhost:80'
	},
	getAuthHeader: (authToken, isEnterprice = false) => {
		if (!authToken)
			return {}

		return isEnterprice
			? { Authorization: `Bearer ${ authToken }` }
			: { 'direktiv-token': authToken }
	},
}
