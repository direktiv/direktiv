/**
 * Playwright suggests using "codegen" to generate a browser state that
 * is then saved as a .json file and referenced in playwright.config.ts.
 *
 * See here: https://playwright.dev/docs/auth
 *
 * Since the .json would contain sensitive information, it should not be
 * committed to the repo. Accordingly, we would have to repeat the steps
 * to generate it before running any tests.
 *
 * Since there are only two variables and it is more convenient to set
 * them via the environment, the code below is based on such a .json file,
 * but just fills these values programatically.
 *
 * If breaking changes are made to our local storage, see the following
 * docs on how to generate a new .json:
 * https://playwright.dev/docs/codegen#preserve-authenticated-state
 *
 * (It is not necessary to write a test to generate the data, but it
 * is possible to perform the required steps manually in the test browser
 * and then save the generated json.)
 *
 * Irrelevant parts contained in the generated json have been removed
 * (e.g., theme and initial workflow). See below for original json.
 */

const port = process.env.VITE_E2E_UI_PORT;
const token = process.env.VITE_E2E_API_TOKEN || "";

// if token is "", no token is added to the request.
export const storageState = {
  cookies: [],
  origins: [
    {
      origin: `http://localhost:${port}`,
      localStorage: [
        {
          name: "direktiv-store-api-key",
          value: `{"state":{"apiKey":"${token}"},"version":0}`,
        },
      ],
    },
  ],
};

/*
 * original JSON as generated with codegen:
{
  "cookies": [],
  "origins": [
    {
      "origin": "http://localhost:3333",
      "localStorage": [
        {
          "name": "direktiv-store-api-key",
          "value": "{\"state\":{\"apiKey\":\"SECRET-KEY"},\"version\":0}"
        },
        {
          "name": "direktiv-store-theme",
          "value": "{\"state\":{\"storedTheme\":null},\"version\":0}"
        },
        {
          "name": "direktiv-store-editor",
          "value": "{\"state\":{\"layout\":\"code\"},\"version\":0}"
        },
        {
          "name": "direktiv-store-namespace",
          "value": "{\"state\":{\"namespace\":\"foobar\"},\"version\":0}"
        }
      ]
    }
  ]
}
 */
