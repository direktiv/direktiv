export default {
    getDirektivHost: function () {
        if (process.env.DIREKTIV_HOST) {
            return process.env.DIREKTIV_HOST
        } else {
            return "http://localhost:80"
        }
    },
    getAuthHeader: (authToken, isEnterprice = false) => {
        if (!authToken) {
            return {};
        }
        return isEnterprice
            ? {Authorization: `Bearer ${authToken}`}
            : {"direktiv-token": authToken};
    }
}