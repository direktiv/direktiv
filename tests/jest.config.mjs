export default {
  testEnvironment: "node",
  testTimeout: 300000,
  transform: {}, // disable babel unless you add one
  watchPlugins: [
    "jest-watch-typeahead/filename",
    "jest-watch-typeahead/testname",
  ],
};
