import { checkIfFileExists, createDirectory } from "e2e/utils/files";
import {
  checkIfNamespaceExists,
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "../utils/workflow";

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
  await page.goto("/", { waitUntil: "networkidle" });

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
  await page.getByText("New").first().click();
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

  const filename = "awesomeworkflow.yaml";

  // create workflow
  await page.getByText("New").first().click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(filename);
  await page.getByTestId("new-workflow-submit").click();

  // assert it has created and navigated to workflow
  await expect(
    page,
    "it creates the workflow and loads the edit page"
  ).toHaveURL(`${namespace}/explorer/workflow/edit/${filename}`);

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

  const nodeCreated = await checkIfFileExists({
    namespace,
    path: `/${filename}`,
  });

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

test("it is possible to create a workflow without providing the .yaml file extension", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a testing namespace is loaded in the explorer"
  ).toHaveText(namespace);

  const filenameWithoutExtension = "awesome-workflow";

  // create workflow
  await page.getByText("New").first().click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(filenameWithoutExtension);
  await page.getByTestId("new-workflow-submit").click();

  // assert it has created and navigated to workflow
  await expect(
    page,
    "it creates the workflow and loads the edit page"
  ).toHaveURL(
    `${namespace}/explorer/workflow/edit/${filenameWithoutExtension}.yaml`
  );

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "breadcrumbs reflect the correct namespace"
  ).toHaveText(namespace);

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "breadcrumbs reflect the file name"
  ).toHaveText(`${filenameWithoutExtension}.yaml`);

  await expect(
    page.getByTestId("workflow-header"),
    "the page heading contains the file name"
  ).toHaveText(`${filenameWithoutExtension}.yaml`);

  const nodeCreated = checkIfFileExists({
    namespace,
    path: `/${filenameWithoutExtension}.yaml`,
  });
  await expect(nodeCreated).toBeTruthy();
});

test("when creating a workflow, the name (before extension) may be the same as a directory name at the same level", async ({
  page,
}) => {
  const directoryName = "directory";
  await createDirectory({ namespace, name: directoryName });

  // go to tree root
  await page.goto(`/${namespace}/explorer/tree`);

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(directoryName);
  await page.getByTestId("new-workflow-submit").click();

  // assert it has created and navigated to workflow
  await expect(
    page,
    "it creates the workflow and loads the edit page"
  ).toHaveURL(`${namespace}/explorer/workflow/edit/${directoryName}.yaml`);

  await expect(
    page.getByTestId("workflow-header"),
    "the page heading contains the file name"
  ).toHaveText(`${directoryName}.yaml`);

  const nodeCreated = await checkIfFileExists({
    namespace,
    path: `/${directoryName}.yaml`,
  });
  await expect(nodeCreated).toBeTruthy();
});

test("it is not possible to create a workflow when the name already exixts", async ({
  page,
}) => {
  const alreadyExists = "workflow.yaml";
  await createWorkflow(namespace, alreadyExists);

  // go to tree root
  await page.goto(`/${namespace}/explorer/tree`);

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(alreadyExists);
  await page.getByTestId("new-workflow-submit").click();

  await expect(page.getByTestId("form-errors")).toContainText(
    "The name already exists"
  );
});

test("it is not possible to create a workflow when the name already exists and the file extension is added automatically", async ({
  page,
}) => {
  const alreadyExists = "workflow.yaml";
  const typedInName = "workflow";
  await createWorkflow(namespace, alreadyExists);

  // go to tree root
  await page.goto(`/${namespace}/explorer/tree`);

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(typedInName);
  await page.getByTestId("new-workflow-submit").click();

  await expect(page.getByTestId("form-errors")).toContainText(
    "The name already exists"
  );
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

  await expect(
    page.getByText(
      `Are you sure you want to delete ${name}? This cannot be undone.`,
      {
        exact: true,
      }
    )
  ).toBeVisible();
  await page.getByTestId("node-delete-confirm").click();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it does not render the old workflow name"
  ).toHaveCount(0);

  const nodeExists = await checkIfFileExists({ namespace, path: `/${name}` });

  await expect(nodeExists).toBeFalsy();
});

test(`it is possible to rename a workflow`, async ({ page }) => {
  const oldName = "old-name.yaml";
  const newName = "new-name.yaml";
  await createWorkflow(namespace, oldName);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it renders the workflow"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldName}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(newName);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newName}`),
    "it renders the new workflow name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it does not render the old workflow name"
  ).toHaveCount(0);

  const originalExists = await checkIfFileExists({
    namespace,
    path: `/${oldName}`,
  });
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfFileExists({ namespace, path: `/${newName}` });
  await expect(isRenamed).toBeTruthy();
});

test(`when renaming a workflow, the name (before extension) may be the same as a directory name at the same level`, async ({
  page,
}) => {
  const oldName = "old-name.yaml";
  const directoryName = "directory";
  const newName = `${directoryName}.yaml`;
  await createDirectory({ namespace, name: directoryName });
  await createWorkflow(namespace, oldName);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it renders the workflow"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldName}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(directoryName);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newName}`),
    "it renders the new workflow name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it does not render the old workflow name"
  ).toHaveCount(0);

  const originalExists = await checkIfFileExists({
    namespace,
    path: `/${oldName}`,
  });
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfFileExists({ namespace, path: `/${newName}` });
  await expect(isRenamed).toBeTruthy();
});

test(`it will automatically add a yaml extension when renaming a workflow`, async ({
  page,
}) => {
  const oldName = "old-name.yaml";
  const newNameWithoutYamlExtension = "new-name";
  const newNameWithYamlExtension = `${newNameWithoutYamlExtension}.yaml`;
  await createWorkflow(namespace, oldName);

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it renders the workflow"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldName}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(newNameWithoutYamlExtension);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newNameWithYamlExtension}`),
    "it renders the new workflow name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it does not render the old workflow name"
  ).toHaveCount(0);

  const originalExists = await checkIfFileExists({
    namespace,
    path: `/${oldName}`,
  });
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfFileExists({
    namespace,
    path: `/${newNameWithYamlExtension}`,
  });
  await expect(isRenamed).toBeTruthy();
});

test(`it is not possible to rename a workflow when the name already exists`, async ({
  page,
}) => {
  const tobeRenamed = "workflow-a.yaml";
  const alreadyExists = "workflow-b.yaml";
  await createWorkflow(namespace, tobeRenamed);
  await createWorkflow(namespace, alreadyExists);

  await page.goto(`/${namespace}/explorer/tree/`);

  await page
    .getByTestId(`explorer-item-${tobeRenamed}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(alreadyExists);
  await page.getByTestId("node-rename-submit").click();

  await expect(page.getByTestId("form-errors")).toContainText(
    "The name already exists"
  );
});

test(`it is not possible to rename a workflow when the name already exists and extension is added automatically`, async ({
  page,
}) => {
  const tobeRenamed = "workflow-a.yaml";
  const alreadyExists = "workflow-b.yaml";
  const alreadyExistsWithoutExtension = "workflow-b";
  await createWorkflow(namespace, tobeRenamed);
  await createWorkflow(namespace, alreadyExists);

  await page.goto(`/${namespace}/explorer/tree/`);

  await page
    .getByTestId(`explorer-item-${tobeRenamed}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page
    .getByTestId("node-rename-input")
    .fill(alreadyExistsWithoutExtension);
  await page.getByTestId("node-rename-submit").click();

  await expect(page.getByTestId("form-errors")).toContainText(
    "The name already exists"
  );
});

test(`it is possible to delete a directory`, async ({ page }) => {
  const name = "directory";
  await createDirectory({ namespace, name });

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

  await expect(
    page.getByText(
      `Are you sure you want to delete ${name}? All content of this directory will be deleted as well.`,
      {
        exact: true,
      }
    )
  ).toBeVisible();

  await page.getByTestId("node-delete-confirm").click();

  await expect(
    page.getByTestId(`explorer-item-${name}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const nodeExists = await checkIfFileExists({ namespace, path: `/${name}` });
  await expect(nodeExists).toBeFalsy();
});

test(`it is possible to rename a directory`, async ({ page }) => {
  const oldName = "old-name";
  const newName = "new-name";
  await createDirectory({ namespace, name: oldName });

  await page.goto(`/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it renders the folder"
  ).toBeVisible();

  await page
    .getByTestId(`explorer-item-${oldName}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();
  await page.getByTestId("node-actions-rename").click();
  await page.getByTestId("node-rename-input").fill(newName);
  await page.getByTestId("node-rename-submit").click();

  await expect(
    page.getByTestId(`explorer-item-${newName}`),
    "it renders the new folder name"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${oldName}`),
    "it does not render the old folder name"
  ).toHaveCount(0);

  const originalExists = await checkIfFileExists({
    namespace,
    path: `/${oldName}`,
  });
  await expect(originalExists).toBeFalsy();

  const isRenamed = await checkIfFileExists({ namespace, path: `/${newName}` });
  await expect(isRenamed).toBeTruthy();
});

test("it is not possible to navigate to a workflow that does not exist", async ({
  page,
}) => {
  await page.goto(
    `${namespace}/explorer/workflow/edit/this-file-does-not-exists.yaml`
  );

  await expect(page.getByTestId("error-title")).toContainText("404");
  await expect(page.getByTestId("error-message")).toContainText(
    "The resource you are trying to access does not exist. This might be due to a typo in the  URL or the resource might have been deleted or renamed."
  );

  await expect(
    page.getByRole("link", { name: "Explorer" }),
    "it still shows the main navigation"
  ).toBeVisible();
});

test("it is not possible to navigate to a folder that does not exist", async ({
  page,
}) => {
  await page.goto(`${namespace}/explorer/tree/this-folder-does-not-exist`);

  await expect(page.getByTestId("error-title")).toContainText("404");
  await expect(page.getByTestId("error-message")).toContainText(
    "The resource you are trying to access does not exist. This might be due to a typo in the  URL or the resource might have been deleted or renamed."
  );

  await expect(
    page.getByRole("link", { name: "Explorer" }),
    "it still shows the main navigation"
  ).toBeVisible();
});

test("it is not possible to navigate to a namespace that does not exist", async ({
  page,
}) => {
  await page.goto(`this-namespace-does-not-exist/explorer/tree`);

  await expect(page.getByTestId("error-title")).toContainText("404");
  await expect(page.getByTestId("error-message")).toContainText(
    "The resource you are trying to access does not exist. This might be due to a typo in the  URL or the resource might have been deleted or renamed."
  );

  await expect(
    page.getByRole("link", { name: "Explorer" }),
    "it does not show the main navigation"
  ).not.toBeVisible();
});
