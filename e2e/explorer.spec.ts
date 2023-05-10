import {
  checkIfNamespaceExists,
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "./utils/namespace";
import {
  checkIfNodeExists,
  createDirectory,
  createWorkflow,
  workflowExamples,
} from "./utils/node";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to a namespace via breadcrumbs", async ({
  page,
}) => {
  // visit page
  await page.goto("/");

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  // at this point, any namespace may be loaded.
  // let's navigate to the test's namespace via breadcrumbs.

  await page.getByTestId("dropdown-trg-namespace").click();
  await page.getByRole("menuitemradio", { name: namespace }).click();

  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
});

test("it is possible to navigate to a namespace via URL", async ({ page }) => {
  // visit url
  await page.goto(`/${namespace}/explorer/tree`);

  // make sure breadcrumb and url are correct after loading
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );
});

test("it is possible to create a namespace via breadcrumbs", async ({
  page,
}) => {
  // visit page and make sure explorer is loaded
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a testing namespace is loaded in the explorer"
  ).toHaveText(namespace);

  // create new namespace
  const newNamespace = createNamespaceName();
  await page.getByTestId("dropdown-trg-namespace").click();
  await page.getByTestId("new-namespace").click();
  await page.getByTestId("new-namespace-name").fill(newNamespace);
  await page.getByTestId("new-namespace-submit").click();

  // make sure it has navigated to new namespace
  await expect(page, "it redirects to the new namespace's url").toHaveURL(
    `/${newNamespace}/explorer/tree`
  );

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the new namespace is reflected in the breadcrumbs"
  ).toHaveText(newNamespace);

  // make sure namespace exists in backend
  const namespaceCreated = await checkIfNamespaceExists(newNamespace);
  await expect(namespaceCreated).toBeTruthy;

  // cleanup
  await deleteNamespace(newNamespace);
});

test("it is possible to create a folder", async ({ page }) => {
  // visit page and make sure explorer is loaded
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a testing namespace is loaded in the explorer"
  ).toHaveText(namespace);

  const folderName = "awesome-folder";

  // create folder
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-dir").click();
  await page.getByPlaceholder("folder-name").fill(folderName);
  await page.getByRole("button", { name: "Create" }).click();

  // it automatically navigates to the folder
  await expect(
    page.getByTestId("breadcrumb-segment").getByText(folderName).first()
  ).toBeVisible();
  await expect(
    page,
    "it creates a new folder and navigates to it automatically"
  ).toHaveURL(`/${namespace}/explorer/tree/${folderName}`);

  // navigate back to tree root
  await page.getByTestId("tree-root").click();
  await expect(
    page,
    "when clicking the tree icon, it navigates back to the tree root"
  ).toHaveURL(`/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${folderName}`),
    "it renders the node in the explorer"
  ).toBeVisible();

  // navigate to folder by clicking on it
  await page.getByTestId(`explorer-item-${folderName}`).click();
  await expect(
    page,
    "when clicking on the folder, it navigates to it"
  ).toHaveURL(`/${namespace}/explorer/tree/${folderName}`);

  // navigate back by clicking on .. "folder"
  await page.getByRole("link", { name: ".." }).click();
  await expect(
    page,
    "when clicking .. it navigates back to the tree root"
  ).toHaveURL(`/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${folderName}`),
    "it renders the node in the explorer"
  ).toBeVisible();
});

test("it is possible to create a workflow", async ({ page }) => {
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a testing namespace is loaded in the explorer"
  ).toHaveText(namespace);

  const filename = "awesome-workflow.yaml";

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(filename);
  await page.getByTestId("new-workflow-editor").fill(workflowExamples.noop);
  await page.getByTestId("new-workflow-submit").click();

  // assert it has created and navigated to workflow
  await expect(
    page,
    "it creates the workflow and loads the active revision page"
  ).toHaveURL(`${namespace}/explorer/workflow/active/${filename}`);

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "breadcrumbs reflect the correct namespace"
  ).toHaveText(namespace);

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "breadcrumbs reflect the file name"
  ).toHaveText(filename);

  await expect(
    page.getByTestId("workflow-header"),
    "the page heading contains the file name"
  ).toHaveText(filename);

  const nodeCreated = await checkIfNodeExists(namespace, filename);
  await expect(nodeCreated).toBeTruthy();

  // TODO: test editor functions in separate test once editor is implemented
  await expect(
    page.getByText(
      "description: A simple 'no-op' state that returns 'Hello world!'"
    )
  ).toBeVisible();

  // navigate back by clicking on the namespace breadcrumb"
  await page.getByTestId("breadcrumb-namespace").getByText(namespace).click(),
    await expect(
      page,
      "when clicking the namespace breadcrumb it navigates to tree root"
    ).toHaveURL(`/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${filename}`),
    "it renders the node in the explorer"
  ).toBeVisible();
});

test(`it is possible to delete a worfklow`, async ({ page }) => {
  const name = "workflow.yaml";
  await createWorkflow(namespace, name);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it renders the node in the explorer"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${name}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-delete").click();
  await page.getByTestId("node-delete-confirm").click();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const nodeExists = await checkIfNodeExists(namespace, name);
  await expect(nodeExists).toBeFalsy();
});

test(`it is possible to rename a workflow`, async ({ page }) => {
  const oldname = "old-name.yaml";
  const newname = "new-name.yaml";
  await createWorkflow(namespace, oldname);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldname}`),
    "it renders the folder"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldname}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(newname);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newname}`),
    "it renders the new folder name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldname}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const originalExists = await checkIfNodeExists(namespace, oldname);
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfNodeExists(namespace, newname);
  await expect(isRenamed).toBeTruthy();
});

test(`it is possible to delete a directory`, async ({ page }) => {
  const name = "directory";
  await createDirectory(namespace, name);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it renders the node in the explorer"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${name}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-delete").click();
  await page.getByTestId("node-delete-confirm").click();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const nodeExists = await checkIfNodeExists(namespace, name);
  await expect(nodeExists).toBeFalsy();
});

// API currently returns a 500 error when trying to rename directory
test(`it is possible to rename a directory`, async ({ page }) => {
  const oldname = "old-name";
  const newname = "new-name";
  await createDirectory(namespace, oldname);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldname}`),
    "it renders the folder"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldname}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(newname);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newname}`),
    "it renders the new folder name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldname}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const originalExists = await checkIfNodeExists(namespace, oldname);
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfNodeExists(namespace, newname);
  await expect(isRenamed).toBeTruthy();
});
