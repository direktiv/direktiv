import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import exampleSnippets, {
  KeyVal,
} from "~/pages/namespace/JqPlayground/Examples/exampleSnippets";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

function deepObjectsAreEqual(objA: KeyVal, objB: KeyVal): boolean {
  if (objA === objB) return true;

  if (typeof objA !== "object" || typeof objB !== "object") return false;

  const keysA = Object.keys(objA);
  const keysB = Object.keys(objB);

  if (keysA.length !== keysB.length) return false;

  for (const key of keysA) {
    if (!keysB.includes(key)) return false;

    if (!deepObjectsAreEqual(objA[key], objB[key])) return false;
  }

  return true;
}

const xCompare = (output: string, expectation: KeyVal) => {
  if (typeof expectation === "number") {
    return parseFloat(output) == expectation;
  } else if (typeof expectation === "string") {
    return output === expectation;
  } else {
    //this is the case of the object
    return deepObjectsAreEqual(JSON.parse(output), expectation);
  }
};

test("it is possible to run simple query against a input json", async ({
  page,
  browserName,
}) => {
  await page.goto(`/${namespace}/jq`);
  const btnRun = page.getByTestId("jq-run-btn");
  await btnRun.click();
  const defaultOutput = "{}";
  const inputTextArea = page.getByTestId("jq-input-editor");

  const outputTextArea = page
    .getByTestId("jq-output-editor")
    .getByRole("textbox");
  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      "the variable's content is loaded into the editor"
    )
    .toBe(defaultOutput);

  const invalidJson = "invalid json";
  await inputTextArea.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type(invalidJson);
  await btnRun.click();

  await expect(
    page.getByTestId("form-errors"),
    "error alert should popup due to invalid json"
  ).toBeVisible();
  await expect
    .poll(
      async () => await outputTextArea.inputValue(),
      "output should be empty on error"
    )
    .toBe("");

  const snippets = page.getByTestId(/jq-run-snippet-/);
  await snippets.first().click();

  await page.waitForTimeout(1000); // scroll has animation to go to top, so should wait for it
  const isScrolledUp = await page.evaluate(() => window.scrollY === 0);
  expect(
    isScrolledUp,
    "scroll should be at the top when click on a snippet run"
  ).toBe(true);

  const invalidQuery = "invalid-query";
  await page.getByTestId("jq-query-input").fill(invalidQuery);
  await btnRun.click();
  await expect(
    page.getByTestId("form-errors"),
    "error alert should popup due to invalid json"
  ).toBeVisible();
});

test("it is possible to run snippets", async ({ page }) => {
  test.setTimeout(50000);
  await page.goto(`/${namespace}/jq`, { waitUntil: "networkidle" });
  // await page.waitForLoadState('networkidle');
  const btnRun = page.getByTestId("jq-run-btn");
  const outputTextArea = page
    .getByTestId("jq-output-editor")
    .getByRole("textbox");

  for (let i = 0; i < exampleSnippets.length; i++) {
    const item = exampleSnippets[i];
    const snippet = page.getByTestId(`jq-run-snippet-${item?.key}-btn`);
    await snippet.click();
    await btnRun.click();
    await expect(
      btnRun,
      "this should get enabled again after query completion"
    ).toBeEnabled();
    await expect
      .poll(async () => {
        const out = await outputTextArea.inputValue();
        return xCompare(out, item?.output);
      }, "output should be the expected output")
      .toBe(true);
    if (i === 0) {
      // test copy button
      await page.getByTestId("copy-output").click({ force: true });
      await page.waitForTimeout(1000);
      const clipboardText = await page.evaluate(() =>
        navigator.clipboard.readText()
      );
      const comparedClipboard = xCompare(clipboardText, item?.output);
      expect(comparedClipboard, "").toBe(true);
    }
  }
});

test("it is possible to save query and input data in local storage", async ({
  page,
  browserName,
}) => {
  await page.goto(`/${namespace}/jq`, { waitUntil: "networkidle" });

  const inputTextArea = page.getByTestId("jq-input-editor");
  const queryInput = page.getByTestId("jq-query-input");
  await queryInput.fill("test-query");

  await inputTextArea.click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type("some-test-input-data");

  await expect
    .poll(
      async () => await inputTextArea.getByRole("textbox").inputValue(),
      "input data should be saved and restored"
    )
    .toBe("some-test-input-data");

  await expect
    .poll(
      async () => await queryInput.inputValue(),
      "query should be saved and restored"
    )
    .toBe("test-query");
});
