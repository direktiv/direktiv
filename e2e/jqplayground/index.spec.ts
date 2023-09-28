import { Page, expect, test } from "@playwright/test";
import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { getCommonElements, getErrorContainer } from "./utils";

import exampleSnippets from "~/pages/namespace/JqPlayground/Examples/exampleSnippets";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("It will display the input as output, when the user clicks the run button without making any changes", async ({
  page,
}) => {
  await page.goto(`/${namespace}/jq`);
  const { btnRun, inputTextArea, outputTextArea, queryInput } =
    getCommonElements(page);

  const expectedDefaultInput = "{}";

  expect(await queryInput.inputValue(), "query input is . by default").toBe(
    "."
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
  await page.goto(`/${namespace}/jq`);
  const { btnRun, queryInput } = await getCommonElements(page);
  await queryInput.fill("some invalid jq command");
  await btnRun.click();

  const { errorContainer } = getErrorContainer(page);

  await expect(
    errorContainer,
    "an error message should be displayed"
  ).toBeVisible();

  expect(
    await errorContainer.textContent(),
    "the error message should inform about an invalid json"
  ).toContain(
    'error : error executing JQ command: failed to evaluate jq/js: error executing jq query some invalid jq command: unexpected token "invalid"'
  );

  await queryInput.fill("changed the query");
  await expect(
    errorContainer,
    "the error message will disappear when the user changes the query"
  ).not.toBeVisible();
});

test("It will display an error when the input is not a valid JSON", async ({
  page,
  browserName,
}) => {
  await page.goto(`/${namespace}/jq`);
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
    "error : invalid json data: invalid character 's' looking for beginning of value"
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
  await page.goto(`/${namespace}/jq`);
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
  expect(
    await outputTextArea.inputValue(),
    `while loading the output from the server, the output will be cleared`
  ).toBe("");
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
  await page.goto(`/${namespace}/jq`);
  const { queryInput } = await getCommonElements(page);

  expect(await queryInput.inputValue(), 'the query is "." by default').toBe(
    "."
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

