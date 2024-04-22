import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "../../utils/node";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";
const defaultDescription = "A simple 'no-op' state that returns 'Hello world!'";

test.beforeEach(async () => {
  namespace = await createNamespace();
  workflow = await createWorkflow(
    namespace,
    faker.internet.domainWord() + ".yaml"
  );
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to the code editor ", async ({ page }) => {
  await page.goto("/");
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();
  // at this point, any namespace may be loaded.
  // let's navigate to the test's namespace via breadcrumbs.
  await page.getByTestId("dropdown-trg-namespace").click();

  await page
    .getByRole("option", {
      name: namespace,
    })
    .click();

  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);

  await page.getByTestId(`explorer-item-link-${workflow}`).click();

  await expect(
    page.getByTestId("workflow-tabs-trg-editor"),
    "screen should have code editor tab"
  ).toBeVisible();

  await expect(page, "the workflow is reflected in the url").toHaveURL(
    `${namespace}/explorer/workflow/edit/${workflow}`
  );
});

test("it is possible to save the workflow", async ({ page }) => {
  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);

  const editorElement = page.getByText(defaultDescription);
  await editorElement.click();

  const testText = faker.random.alphaNumeric(9);
  await page.type("textarea", testText);

  // now click on Save
  const saveButton = page.getByTestId("workflow-editor-btn-save");
  await saveButton.click();

  // Commented out since this is not a critical step, but maybe we can enable
  // it again at some point after learning more about the following problem:
  // These steps fail locally, but work with a remote API. They works locally
  // with throttling enabled in devtools. This implies the request is completed
  // so fast there is not enough time to detect the inactive button.
  // await expect(
  //   saveButton,
  //   "save button should be disabled during the api call"
  // ).toBeDisabled();
  // await expect(
  //   saveButton,
  //   "save button should be enabled after the api call"
  // ).toBeEnabled();

  // after saving is completed screen should have those new changed text before/after the page reload
  await expect(
    page.getByText(testText),
    "after saving, screen should have the updated text"
  ).toBeVisible();
  await page.reload({ waitUntil: "networkidle" });
  await expect(
    page.getByText(testText),
    "after reloading, screen should have the updated text"
  ).toBeVisible();

  // check the text at the bottom left
  await expect(
    page.getByTestId("workflow-txt-updated"),
    "text should be Updated a few seconds ago"
  ).toHaveText("Updated a few seconds ago");
});

test("it renders response errors when saving an invalid workflow", async ({
  page,
}) => {
  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);

  const editor = page.locator(".lines-content");

  await editor.click();
  await editor.type("notvalidyaml");

  await expect(
    page.getByText("unsaved changes"),
    "it renders a hint that there are unsaved changes"
  ).toBeVisible();

  await page.getByTestId("workflow-editor-btn-save").click();

  await expect(
    page.getByText("There is an issue"),
    "after saving, it renders an error hint in the editor"
  ).toBeVisible();
  await expect(
    page.getByText("updated file data has invalid yaml string"),
    "it renders an error popup with the error message"
  ).toBeVisible();

  await expect(
    page.getByText("unsaved changes"),
    "it still renders a hint that there are unsaved changes"
  ).toBeVisible();
});

test("it is possible to navigate to another route from the editor", async ({
  page,
}) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    dialogTriggered = true;
    return dialog.dismiss();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription);

  await page.getByRole("link", { name: "Settings" }).click();
  await expect(dialogTriggered).toBe(false);

  await expect(page, "it navigates to the new route").toHaveURL(
    `${namespace}/settings`
  );
});

test("it prevents navigation to another route with unsaved changes", async ({
  page,
}) => {
  const expectedMsg =
    "You have unsaved changes that will be lost when leaving this route. Are you sure you want to leave?";
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    await expect(dialog.message()).toBe(expectedMsg);
    dialogTriggered = true;
    return dialog.dismiss();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription).click();

  const dirtyText = faker.random.alphaNumeric(9);
  await page.type("textarea", dirtyText);

  await page.getByRole("link", { name: "Settings" }).click();
  await expect(dialogTriggered).toBe(true);

  await expect(
    page,
    "after dismissing the dialog, it stays on the same route"
  ).toHaveURL(`${namespace}/explorer/workflow/edit/${workflow}`);

  await expect(
    page.getByText(dirtyText),
    "the edited text is still in the editor"
  ).toBeVisible();
});

test("with confirmation, it navigates to another route despite unsaved changes", async ({
  page,
}) => {
  const expectedMsg =
    "You have unsaved changes that will be lost when leaving this route. Are you sure you want to leave?";
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    await expect(dialog.message()).toBe(expectedMsg);
    dialogTriggered = true;
    return dialog.accept();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription).click();

  const dirtyText = faker.random.alphaNumeric(9);
  await page.type("textarea", dirtyText);

  await page.getByRole("link", { name: "Settings" }).click();
  await expect(dialogTriggered).toBe(true);

  await expect(
    page,
    "after confirming the dialog, it navigates to the new route"
  ).toHaveURL(`${namespace}/settings`);
});

test("it is possible to leave the app from the editor", async ({ page }) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    await expect(dialog.type()).toBe("beforeunload");
    dialogTriggered = true;
    await dialog.dismiss();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription);

  await page.goto("/api/v2/status");

  await expect(dialogTriggered).toBe(false);

  await expect(page, "it navigates to the new document").toHaveURL(
    "/api/v2/status"
  );
});

test("it prevents navigation away from the app with unsaved changes", async ({
  page,
}) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    await expect(dialog.type()).toBe("beforeunload");
    dialogTriggered = true;
    await dialog.dismiss();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription).click();

  const dirtyText = faker.random.alphaNumeric(9);
  await page.type("textarea", dirtyText);

  try {
    await page.goto("/api/v2/status").catch();
  } catch (error) {
    return;
  }

  await expect(dialogTriggered).toBe(true);

  await expect(
    page,
    "after dismissing the dialog, it stays on the same route"
  ).toHaveURL(`${namespace}/explorer/workflow/edit/${workflow}`);

  await expect(
    page.getByText(dirtyText),
    "the edited text is still in the editor"
  ).toBeVisible();
});

test("with confirmation, it allows navigation away from the app with unsaved changes", async ({
  page,
}) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    await expect(dialog.type()).toBe("beforeunload");
    dialogTriggered = true;
    await dialog.accept();
  });

  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);
  await page.getByText(defaultDescription).click();

  const dirtyText = faker.random.alphaNumeric(9);
  await page.type("textarea", dirtyText);

  await page.goto("/api/v2/status");

  await expect(dialogTriggered).toBe(true);
  await expect(page, "it has navigated to the new page").toHaveURL(
    "/api/v2/status"
  );
});
