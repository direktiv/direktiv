const isGithubActions = process.env.GITHUB_ACTIONS === "true";

module.exports = {
  testTimeout: isGithubActions ? 40000 : 10000,
  watchPlugins: [
    "jest-watch-typeahead/filename",
    "jest-watch-typeahead/testname",
  ],
};
