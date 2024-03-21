import { Locator, expect, test } from "@playwright/test";
import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  jsonSchemaFormWorkflow,
  jsonSchemaWithRequiredEnum,
  testDiacriticsWorkflow,
} from "./utils";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/components/modals/CreateNew/Workflow/templates";
import { createFile } from "e2e/utils/files";
import { decode } from "js-base64";
import { faker } from "@faker-js/faker";
import { getInput } from "~/api/instances/query/input";
import { headers } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to open and use the run workflow modal from the editor and the header of the workflow page", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: basicWorkflow.data,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  // open modal via editor button
  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog from the editor button"
  ).toBeVisible();
  await page.getByTestId("run-workflow-cancel-btn").click();
  expect(await page.getByTestId("run-workflow-dialog")).not.toBeVisible();

  // open modal via header button
  await page.getByTestId("workflow-header-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog from the header button"
  ).toBeVisible();

  // use the tabs
  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected"),
    "the json tab is selected by default (since this workflow has no JSON schema)"
  ).toBe("true");

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected")
  ).toBe("false");

  await page.getByTestId("run-workflow-form-tab-btn").click();

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected"),
    "the form tab is now selected"
  ).toBe("true");

  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected")
  ).toBe("false");

  expect(
    await page.getByTestId("run-workflow-form-input-hint"),
    "it shows a hint that no form could be generated"
  ).toBeVisible();

  await page.getByTestId("run-workflow-cancel-btn").click();
  expect(await page.getByTestId("run-workflow-dialog")).not.toBeVisible();
});

test("it is possible to run the workflow by setting an input JSON via the editor", async ({
  page,
  browserName,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: basicWorkflow.data,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "the submit button is enabled by default"
  ).toEqual(true);

  await page.type("textarea", "some invalid json");

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "submit button is disabled when the json is invalid"
  ).toEqual(false);

  await page.getByTestId("run-workflow-editor").click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  const userInputString = `{"string": "1", "integer": 1, "boolean": true, "array": [1,2,3], "object": {"key": "value"}}`;
  await page.keyboard.type(userInputString);

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "submit is enabled when the json is valid"
  ).toEqual(true);

  // submit to run the workflow
  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  // check the server state of the input
  const res = await getInput({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      instanceId,
      namespace,
    },
    headers,
  });

  const inputResponseString = decode(res.data);

  expect(
    inputResponseString,
    "the server result is the same as the input that was sent"
  ).toBe(userInputString);
});

test("it is possible to run a workflow with input data containing special characters", async ({
  page,
}) => {
  const name = "test-diacritics.yaml";

  await createFile({
    name,
    namespace,
    type: "workflow",
    yaml: testDiacriticsWorkflow,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${name}`);

  await expect(
    page.locator(".view-lines"),
    "The editor renders special characters correctly"
  ).toContainText("A workflow for testing characters like îèüñÆ");

  await page.getByTestId("workflow-editor-btn-run").click();
  await page.getByLabel("Name").fill("Kateřina Horáčková");
  await page.getByTestId("run-workflow-submit-btn").click();

  await expect(
    page.locator(".lines-content"),
    "The text from the input is rendered correctly in the workflow output"
  ).toContainText(
    `{    
    "result": "Hello Kateřina Horáčková"
}`,
    { useInnerText: true }
  );
});

test("it is not possible to run the workflow when the editor has unsaved changes", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: basicWorkflow.data,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await expect(page.getByTestId("workflow-editor-btn-run")).not.toBeDisabled();

  await page.type("textarea", faker.random.alphaNumeric(9));

  await expect(page.getByTestId("workflow-editor-btn-run")).toBeDisabled();
});

test("it is possible to provide the input via generated form", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: jsonSchemaFormWorkflow,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected"),
    "it detects the validate step and makes the form tab active by default"
  ).toBe("true");

  // it generated a form (first and last name are required)
  await expect(page.getByLabel("First Name")).toBeVisible();
  await expect(page.getByLabel("Last Name")).toBeVisible();
  await expect(page.getByLabel("Age")).toBeVisible();
  await expect(page.getByRole("combobox", { name: "role" })).toBeVisible();

  await expect(
    await page.getByRole("combobox", { name: "role" }).innerText(),
    "the select input shows a fallback text when it has no value"
  ).toBe("Select role");

  await expect(page.getByTestId("json-schema-form-add-button")).toBeVisible();
  await expect(page.getByLabel("Age")).toBeVisible();
  await expect(page.getByLabel("File")).toBeVisible();

  // interact with the select input
  await page.getByRole("combobox", { name: "role" }).click();
  await page.getByRole("option", { name: "guest" }).click();
  await expect(
    await page.getByRole("combobox", { name: "role" }).innerText(),
    "the select input now shows the selected value"
  ).toBe("guest");

  // interact with the file input
  await page
    .getByLabel("File")
    .setInputFiles("./e2e/utils/fixtures/upload-testfile.txt");

  // interact with the array input
  await page.getByTestId("json-schema-form-add-button").click();
  await page.getByTestId("json-schema-form-add-button").click();
  await page.getByTestId("json-schema-form-add-button").click();
  await page.getByLabel("array-0*").fill("array item 2");
  await page.getByLabel("array-1*").fill("array item 1");
  await page.getByTestId("json-schema-form-down-button-0").click(); // switch 1 and 2
  await page
    .getByLabel("array-2*")
    .fill("this will be deleted in the next step");
  await page.getByTestId("json-schema-form-remove-button-2").click();

  // interact with the number input
  await page.getByLabel("Age").fill("2");

  // submit this form via enter:
  // we have an array form on this page, which also has some buttons
  // using enter here makes sure that we will submit the form and
  // not trigger the buttons from the array form
  await page.keyboard.press("Enter");

  // last name is required and we just tried to send the form without filling it
  await expect(page.getByLabel("First Name")).toBeFocused();
  await page.getByLabel("First Name").fill("Marty");
  await page.getByTestId("run-workflow-submit-btn").click();

  // first name is also required and will now be focused
  await expect(page.getByLabel("Last Name")).toBeFocused();
  await page.getByLabel("Last Name").fill("McFly");
  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  // check the server state of the input
  const res = await getInput({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      instanceId,
      namespace,
    },
    headers,
  });

  const expectedJson = {
    age: 2,
    array: ["array item 1", "array item 2"],
    firstName: "Marty",
    lastName: "McFly",
    select: "guest",
    file: `data:text/plain;base64,SSBhbSBqdXN0IGEgdGVzdGZpbGUgdGhhdCBjYW4gYmUgdXNlZCB0byB0ZXN0IGFuIHVwbG9hZCBmb3JtIHdpdGhpbiBhIHBsYXl3cmlnaHQgdGVzdA==`,
  };
  const inputResponseAsJson = JSON.parse(decode(res.data));
  expect(inputResponseAsJson).toEqual(expectedJson);
});

test("it is possible to provide the input via generated form and resolve form errors", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: jsonSchemaWithRequiredEnum,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected"),
    "it detects the validate step and makes the form tab active by default"
  ).toBe("true");

  // it generated a form (first and last name are required)
  await expect(page.getByLabel("First Name")).toBeVisible();
  await expect(page.getByLabel("Last Name")).toBeVisible();
  await expect(page.getByRole("combobox", { name: "role" })).toBeVisible();

  //click on submit
  await page.getByTestId("run-workflow-submit-btn").click();

  // last name is required and we just tried to send the form without filling it
  await expect(page.getByLabel("First Name")).toBeFocused();
  await page.getByLabel("First Name").fill("Marty");
  await page.getByTestId("run-workflow-submit-btn").click();

  // first name is also required and will now be focused
  await expect(page.getByLabel("Last Name")).toBeFocused();
  await page.getByLabel("Last Name").fill("McFly");
  await page.getByTestId("run-workflow-submit-btn").click();

  // shows the error to select option
  await expect(
    page.getByTestId("jsonschema-form-error"),
    "an error should be visible"
  ).toBeVisible();

  await expect(
    page.getByTestId("jsonschema-form-error"),
    "error message should be \"must have required property 'role'\""
  ).toContainText("must have required property 'role'");

  // interact with the select input
  await page.getByRole("combobox", { name: "role" }).click();
  await page.getByRole("option", { name: "guest" }).click();
  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  // check the server state of the input
  const res = await getInput({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      instanceId,
      namespace,
    },
    headers,
  });

  const expectedJson = {
    firstName: "Marty",
    lastName: "McFly",
    select: "guest",
  };
  const inputResponseAsJson = JSON.parse(decode(res.data));
  expect(inputResponseAsJson).toEqual(expectedJson);
});

test("it is possible to provide the input via Form Input and see the same data in the tab JSON Input", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: jsonSchemaWithRequiredEnum,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected"),
    "it detects the validate step and makes the form tab active by default"
  ).toBe("true");

  // it generated a form (first and last name are required)
  await expect(page.getByLabel("First Name")).toBeVisible();
  await expect(page.getByLabel("Last Name")).toBeVisible();
  await expect(page.getByRole("combobox", { name: "role" })).toBeVisible();

  await page.getByLabel("First Name").fill("Marty");
  await page.getByLabel("Last Name").fill("McFly");

  // interact with the select input
  await page.getByRole("combobox", { name: "role" }).click();
  await page.getByRole("option", { name: "guest" }).click();

  // switch to tab json input
  await page.getByTestId("run-workflow-json-tab-btn").click();

  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected"),
    "the json tab is selected"
  ).toBe("true");

  const expectedJson = {
    firstName: "Marty",
    lastName: "McFly",
    select: "guest",
  };

  // TODO: check if the json input is correct - loose test - is working
  expect(
    await page.locator(".view-line > span"),
    "the Input was set as expected"
  ).toContainText(["Marty", "McFly"]);

  // TODO: check if the json input is correct - strict Test - is NOT working
  // ***
  const visibleJsonArray = await page.locator(".view-line > span")
    .allInnerTexts;

  const visibleJsonObject = Object.fromEntries(visibleJsonArray);

  expect(visibleJsonObject, "the Input was set as expected").toMatchObject(
    expectedJson
  );
  // ***

  // run the workflow in the json tab
  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  // check the server state of the input
  const res = await getInput({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      instanceId,
      namespace,
    },
    headers,
  });

  const inputResponseAsJson = JSON.parse(decode(res.data));

  // the data in the json input is the same
  await expect(
    inputResponseAsJson,
    "workflow was triggered and the input is the same as initially set in the other tab"
  ).toEqual(expectedJson);

  // the input in the workflow is the same like entered in the form
  expect(
    await page.locator(".view-line > span"),
    "the Input was set as expected"
  ).toContainText(["Marty", "McFly"]);
});

test("it is possible to provide the input via JSON Input and see the same data in the tab Form Input", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: jsonSchemaWithRequiredEnum,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  // switch to tab JSON input
  await page.getByTestId("run-workflow-json-tab-btn").click();

  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected"),
    "the json tab is selected"
  ).toBe("true");

  // clear editor, to prevent invalid JSON due to auto completion
  await page.getByRole("textbox").fill("");

  // give valid JSON data
  await page
    .getByRole("textbox")
    .fill('{"firstName":"Marty","lastName":"McFly","select":"guest"}');

  // switch to tab Form input
  await page.getByTestId("run-workflow-form-tab-btn").click();

  // the generated form is visible
  await expect(page.getByLabel("First Name")).toBeVisible();
  await expect(page.getByLabel("Last Name")).toBeVisible();
  await expect(page.getByRole("combobox", { name: "role" })).toBeVisible();

  expect(
    await page.getByLabel("First Name"),
    "the value for first name was set automatically"
  ).toHaveValue("Marty");

  expect(
    await page.getByLabel("Last Name"),
    "the value for last name was set automatically"
  ).toHaveValue("McFly");

  expect(
    await page.getByRole("combobox").locator("span"),
    "the value for role was set automatically"
  ).toContainText("guest");

  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  const expectedJson = {
    firstName: "Marty",
    lastName: "McFly",
    select: "guest",
  };

  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);

  await expect(
    page.locator("div").getByText("complete", { exact: true }),
    "wait until workflow is complete"
  ).toBeVisible();

  await expect(
    page.getByRole("tab", { name: "Input" }),
    "tab for input is visible"
  ).toBeVisible();

  // resize window to see the whole json
  await page
    .locator(
      ".flex > div > .\\[\\&\\>\\*\\]\\:rounded-none > div:nth-child(2) > .inline-flex"
    )
    .click();

  await page.getByRole("tab", { name: "Input" }).click();

  // TODO: Loose Test - is working
  expect(
    await page.locator(".view-line > span"),
    "the Input was set as expected"
  ).toContainText(["Marty", "McFly"]);

  // TODO: Strict Test 1 - is NOT working
  // ***
  expect(
    await page.locator(".view-line > span").allInnerTexts,
    "the Input was set as expected"
  ).toEqual(expect.objectContaining(expectedJson));
  // ***

  // TODO: Strict Test 2 - is NOT working
  // ***
  async function getTextFromLocator(locator: Locator) {
    const elements = await locator.all();
    let combinedText = "";

    for (const element of elements) {
      const text = await element.innerText();
      const cleantext = removeSpacesAndBackslashes(text);
      combinedText += cleantext;
    }
    return combinedText;
  }

  function removeSpacesAndBackslashes(inputString: string) {
    return inputString.replace(/([\\ ])(?=(?:[^"]|"[^"]*")*$)/g, "");
  }

  const elementsLocator = page.locator(".view-line > span");
  const combinedText = await getTextFromLocator(elementsLocator);

  const currentJson = removeSpacesAndBackslashes(combinedText);

  expect(await currentJson, "OBJ EQUAL the Input was set as expected").toEqual(
    expectedJson
  );
  // ***
});

test("the input is synchronized between tabs, but the data that is currently in the view will be sent", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    yaml: jsonSchemaWithRequiredEnum,
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  // switch to tab JSON input
  await page.getByTestId("run-workflow-json-tab-btn").click();

  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected"),
    "the json tab is selected"
  ).toBe("true");

  // clear editor, to prevent invalid JSON due to auto completion
  await page.getByRole("textbox").fill("");

  // give valid JSON data
  await page
    .getByRole("textbox")
    .fill('{"firstName":"Marty","lastName":"McFly","select":"guest"}');

  // switch to tab Form input
  await page.getByTestId("run-workflow-form-tab-btn").click();

  // the generated form is visible
  await expect(page.getByLabel("First Name")).toBeVisible();
  await expect(page.getByLabel("Last Name")).toBeVisible();
  await expect(page.getByRole("combobox", { name: "role" })).toBeVisible();

  expect(
    await page.getByLabel("First Name"),
    "the value for first name was set automatically"
  ).toHaveValue("Marty");

  expect(
    await page.getByLabel("Last Name"),
    "the value for last name was set automatically"
  ).toHaveValue("McFly");

  expect(
    await page.getByRole("combobox").locator("span"),
    "the value for role was set automatically"
  ).toContainText("guest");

  // change data again
  await page.getByLabel("Last Name").fill("McDonald");

  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  const expectedJson = {
    firstName: "Marty",
    lastName: "McDonald",
    select: "guest",
  };

  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);

  await expect(
    page.locator("div").getByText("complete", { exact: true }),
    "wait until workflow is complete"
  ).toBeVisible();

  await expect(
    page.getByRole("tab", { name: "Input" }),
    "tab for input is visible"
  ).toBeVisible();

  await page.getByRole("tab", { name: "Input" }).click();

  // resize window to see the whole json
  await page
    .locator(
      ".flex > div > .\\[\\&\\>\\*\\]\\:rounded-none > div:nth-child(2) > .inline-flex"
    )
    .click();

  expect(
    await page.locator(".view-line > span"),
    "the Input was set as expected"
  ).toContainText(["Marty", "McDonald"]);
});
