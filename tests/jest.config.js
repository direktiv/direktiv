const isGithubActions = process.env.GITHUB_ACTIONS === "true";
if (isGithubActions) {
  console.log("Running in GitHub Actions environment.");
  console.log("DIREKTIV_HOST:", process.env.DIREKTIV_HOST);
}
module.exports = {
  testTimeout: isGithubActions ? 40000 : 10000,
  watchPlugins: [
    "jest-watch-typeahead/filename",
    "jest-watch-typeahead/testname",
  ],
};
