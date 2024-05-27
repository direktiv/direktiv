import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";
import {
  expectedSnippetOutput,
  getCommonElements,
  getErrorContainer,
  scrolledToTheTop,
  userScrolledADecentAmount,
} from "./utils";

import { mockClipboardAPI } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async ({ page }) => {
  namespace = await createNamespace();
  await mockClipboardAPI(page);
  /**
   * networkidle is required to avoid flaky tests. The monaco
   * editor needs to be full loaded before we interact with it.
   */
  await page.goto(`/n/${namespace}/jq`, { waitUntil: "networkidle" });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("It will display the input as output, when the user clicks the run button without making any changes", async ({
  page,
}) => {
  const { btnRun, inputTextArea, outputTextArea, queryInput } =
    getCommonElements(page);

  const expectedDefaultInput = "{}";

  expect(await queryInput.inputValue(), "query input is . by default").toBe(
    "jq(.)"
  );

  expect(await outputTextArea.inputValue(), "output is empty by default").toBe(
    ""
  );

  expect(
    await inputTextArea.inputValue(),
    `input is ${expectedDefaultInput} by default`
  ).toBe(expectedDefaultInput);

  await btnRun.click();

  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      `output changes to ${expectedDefaultInput} after clicking run`
    )
    .toBe(expectedDefaultInput);
});

test("It will display an error when the query is not a JQ command", async ({
  page,
}) => {
  const { btnRun, queryInput } = await getCommonElements(page);

  await queryInput.fill(`"jq(\\)"`);
  await btnRun.click();

  const { errorContainer } = getErrorContainer(page);

  await expect(
    errorContainer,
    "an error message should be displayed"
  ).toBeVisible();

  expect(
    await errorContainer.textContent(),
    "the error message should inform about an invalid json"
  ).toContain("root : invalid 'jx': yaml: found unknown escape character");

  await queryInput.fill("changed the query");
  await expect(
    errorContainer,
    "the error message will disappear when the user changes the query"
  ).not.toBeVisible();
});

test("It will show an error in the logs, when the query is not a valid JQ command", async ({
  page,
}) => {
  const { btnRun, queryInput, logsTextArea } = await getCommonElements(page);
  await queryInput.fill("jq(\\.foo)");

  await expect(
    await logsTextArea.inputValue(),
    "the logs are empty by default"
  ).toBe("");

  await btnRun.click();
  await expect
    .poll(
      async () => await logsTextArea.inputValue(),
      "an error message should be displayed in the logs"
    )
    .toBe(
      'failure: error executing jq query \\.foo: unexpected token "\\\\"\n'
    );
});

test("It will display an error when the input is not a valid JSON", async ({
  page,
  browserName,
}) => {
  const { btnRun, inputTextContainer } = getCommonElements(page);
  await inputTextContainer.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("some invalid json");
  await btnRun.click();

  const errorContainer = page.getByTestId("form-errors");

  await expect(
    errorContainer,
    "an error message should be displayed"
  ).toBeVisible();

  expect(
    await errorContainer.textContent(),
    "the error message should inform about an invalid json"
  ).toContain(
    "root : invalid 'data': invalid character 's' looking for beginning of value"
  );

  await inputTextContainer.click();
  await page.keyboard.type("changed the input json");
  await expect(
    errorContainer,
    "the error message will disappear when the user changes the input"
  ).not.toBeVisible();
});

test("It will clear output when loading the result from the server", async ({
  page,
}) => {
  const { btnRun, outputTextArea } = getCommonElements(page);

  expect(
    await outputTextArea.inputValue(),
    `the initial ouput is an empty string`
  ).toBe("");
  await btnRun.click();
  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      "the output will change from the initial empty string to the result from the server"
    )
    .toBe("{}");

  await btnRun.click();
  await outputTextArea.inputValue(),
    await expect
      .poll(
        async () => await outputTextArea.inputValue(),
        `while loading the output from the server, the output will be cleared`
      )
      .toBe("{}");
  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      "the output will change back to the result from the server"
    )
    .toBe("{}");
});

test("It will persist the query to be available after a page reload", async ({
  page,
}) => {
  const { queryInput } = await getCommonElements(page);

  expect(await queryInput.inputValue(), 'the query is "." by default').toBe(
    "jq(.)"
  );

  const userQueryText = ".some .query .text";

  await queryInput.fill(userQueryText);

  expect(
    await queryInput.inputValue(),
    "the query has been changed by the user"
  ).toBe(userQueryText);

  await page.reload();

  expect(
    await queryInput.inputValue(),
    `after a page reload, the query has been restored to the last value`
  ).toBe(userQueryText);
});

test("It will persist the input to be available after a page reload", async ({
  page,
  browserName,
}) => {
  const { inputTextArea, inputTextContainer } = getCommonElements(page);

  expect(await inputTextArea.inputValue(), `the input is {} by default`).toBe(
    "{}"
  );

  const userInputText = `{"foo": 42,"bar": "less interesting data"}`;

  await inputTextContainer.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type(userInputText);

  expect(
    await inputTextArea.inputValue(),
    `the input was changed by the user`
  ).toBe(userInputText);

  // wait a second to make sure the input is persisted (this was flaky without the wait)
  await page.waitForTimeout(1000);

  await page.reload({});

  expect(
    await inputTextArea.inputValue(),
    `after a page reload, the input has been restored to the last value`
  ).toBe(userInputText);
});

test("the user can copy the input to the clipboard when there is one", async ({
  page,
  browserName,
}) => {
  const { inputTextContainer, copyInputBtn } = getCommonElements(page);

  await inputTextContainer.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("");
  await expect(
    copyInputBtn,
    "an empty input will disable the copy button"
  ).toBeDisabled();

  const userInputText = `{"this": "will be copied into the clipboard"}`;
  await inputTextContainer.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type(userInputText);
  await copyInputBtn.click();
  const clipboardText = await page.evaluate(() =>
    navigator.clipboard.readText()
  );

  expect(clipboardText, "the input was copied into the clipboard").toBe(
    userInputText
  );
});

test("the user can copy the output to the clipboard when there is one", async ({
  page,
}) => {
  const { outputTextArea, copyOutputBtn } = getCommonElements(page);

  expect(
    await outputTextArea.inputValue(),
    `the initial ouput is an empty string`
  ).toBe("");
  await expect(
    copyOutputBtn,
    "an empty output will disable the copy button"
  ).toBeDisabled();

  const snippetToRun = "feedInput";
  const expectedOutput = expectedSnippetOutput[snippetToRun];

  const snippetButton = page.getByTestId(`jq-run-snippet-${snippetToRun}-btn`);
  await snippetButton.click();

  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      "running the snippet should change the output"
    )
    .toBe(expectedOutput);

  await copyOutputBtn.click();
  const clipboardText = await page.evaluate(() =>
    navigator.clipboard.readText()
  );

  expect(clipboardText, "the output was copied into the clipboard").toBe(
    expectedOutput
  );
});

test("the user can copy the logs to the clipboard when there are some", async ({
  page,
}) => {
  const { logsTextArea, queryInput, btnRun, copyLogsBtn } =
    getCommonElements(page);

  expect(
    await logsTextArea.inputValue(),
    `the initial logs are an empty string`
  ).toBe("");

  await expect(
    copyLogsBtn,
    "empty logs will disable the copy button"
  ).toBeDisabled();

  await queryInput.fill("jq(\\.)");
  await btnRun.click();

  const expectedLogs =
    'failure: error executing jq query \\.: unexpected token "\\\\"\n';

  await expect
    .poll(
      async () => await logsTextArea.inputValue(),
      "an error message should be displayed in the logs"
    )
    .toBe(expectedLogs);

  await copyLogsBtn.click();
  const clipboardText = await page.evaluate(() =>
    navigator.clipboard.readText()
  );

  expect(clipboardText, "the logs were copied into the clipboard").toBe(
    expectedLogs
  );
});

test("It will run every snippet succefully", async ({ page }) => {
  const { outputTextArea } = getCommonElements(page);

  for (const [snippetKey, expectedOutput] of Object.entries(
    expectedSnippetOutput
  )) {
    const snippetButton = page.getByTestId(`jq-run-snippet-${snippetKey}-btn`);
    await snippetButton.click();

    await expect
      .poll(
        async () => await outputTextArea.inputValue(),
        "running the snippet should match the expected output"
      )
      .toBe(expectedOutput);
  }
});

test("running a snippet will automatically scroll the page to the top", async ({
  page,
}) => {
  const snippetToRun = "stringInterpolation";
  const snippetButton = page.getByTestId(`jq-run-snippet-${snippetToRun}-btn`);

  await snippetButton.click();
  expect(
    await userScrolledADecentAmount(page),
    `the user has scrolled a decent amount to reach the button`
  ).toBe(true);

  await expect
    .poll(
      async () => await scrolledToTheTop(page),
      `the page automatically scrolled to the top after running the snippet`
    )
    .toBe(true);
});

test("running a snippet will store the jqx and input in local storage", async ({
  page,
}) => {
  const { inputTextArea, queryInput } = getCommonElements(page);
  const snippetButton = page.getByTestId(`jq-run-snippet-valueAtKey-btn`);

  await snippetButton.click();

  const expectedInput = `{
    "foo": 42,
    "bar": "less interesting data"
}`;

  const expectedJqx = "jq(.foo)";

  expect(
    await queryInput.inputValue(),
    "the jqx from the example snippet is used"
  ).toBe(expectedJqx);
  await expect
    .poll(
      async () => await inputTextArea.inputValue(),
      "the input from the example snippet is used"
    )
    .toBe(expectedInput);

  await page.reload();

  expect(
    await queryInput.inputValue(),
    "the jqx will be restored from local storage"
  ).toBe(expectedJqx);
  await expect
    .poll(
      async () => await inputTextArea.inputValue(),
      "the input will be restored from local storage"
    )
    .toBe(expectedInput);
});
