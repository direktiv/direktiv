import { checkIfFileExists, createDirectory } from "e2e/utils/files";
import {
  checkIfNamespaceExists,
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createService } from "./service/utils";
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
    `/n/${namespace}/explorer/tree`
  );
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
});

test("it is possible to navigate to a namespace via URL", async ({ page }) => {
  // visit url
  await page.goto(`/n/${namespace}/explorer/tree`);

  // make sure breadcrumb and url are correct after loading
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/n/${namespace}/explorer/tree`
  );
});

test("it is possible to create a namespace via breadcrumbs", async ({
  page,
}) => {
  // visit page and make sure explorer is loaded
  await page.goto(`/n/${namespace}/explorer/tree`);
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
    `/n/${newNamespace}/explorer/tree`
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
  await page.goto(`/n/${namespace}/explorer/tree`);
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
  ).toHaveURL(`/n/${namespace}/explorer/tree/${folderName}`);

  // navigate back to tree root
  await page.getByTestId("tree-root").click();
  await expect(
    page,
    "when clicking the tree icon, it navigates back to the tree root"
  ).toHaveURL(`/n/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${folderName}`),
    "it renders the node in the explorer"
  ).toBeVisible();

  // navigate to folder by clicking on it
  await page.getByTestId(`explorer-item-${folderName}`).click();
  await expect(
    page,
    "when clicking on the folder, it navigates to it"
  ).toHaveURL(`/n/${namespace}/explorer/tree/${folderName}`);

  // navigate back by clicking on .. "folder"
  await page.getByRole("link", { name: ".." }).click();
  await expect(
    page,
    "when clicking .. it navigates back to the tree root"
  ).toHaveURL(`/n/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${folderName}`),
    "it renders the node in the explorer"
  ).toBeVisible();
});

test("it is possible to create a workflow", async ({ page }) => {
  await page.goto(`/n/${namespace}/explorer/tree`);
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
  ).toHaveURL(`/n/${namespace}/explorer/workflow/edit/${filename}`);

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
    ).toHaveURL(`/n/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId(`explorer-item-${filename}`),
    "it renders the node in the explorer"
  ).toBeVisible();
});

test("it is possible to create a workflow without providing the .yaml file extension", async ({
  page,
}) => {
  await page.goto(`/n/${namespace}/explorer/tree`);
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
    `/n/${namespace}/explorer/workflow/edit/${filenameWithoutExtension}.yaml`
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
  await page.goto(`/n/${namespace}/explorer/tree`);

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(directoryName);
  await page.getByTestId("new-workflow-submit").click();

  // assert it has created and navigated to workflow
  await expect(
    page,
    "it creates the workflow and loads the edit page"
  ).toHaveURL(`/n/${namespace}/explorer/workflow/edit/${directoryName}.yaml`);

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

test("it is not possible to create a workflow when the name already exists", async ({
  page,
}) => {
  const alreadyExists = "workflow.yaml";
  await createWorkflow(namespace, alreadyExists);

  // go to tree root
  await page.goto(`/n/${namespace}/explorer/tree`);

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
  await page.goto(`/n/${namespace}/explorer/tree`);

  // create workflow
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-workflow").click();
  await page.getByTestId("new-workflow-name").fill(typedInName);
  await page.getByTestId("new-workflow-submit").click();

  await expect(page.getByTestId("form-errors")).toContainText(
    "The name already exists"
  );
});

test(`it is possible to rename a workflow`, async ({ page }) => {
  const oldName = "old-name.yaml";
  const newName = "new-name.yaml";
  await createWorkflow(namespace, oldName);

  await page.goto(`/n/${namespace}/explorer/tree/`);
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

  await page.goto(`/n/${namespace}/explorer/tree/`);
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

  await page.goto(`/n/${namespace}/explorer/tree/`);
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

  await page.goto(`/n/${namespace}/explorer/tree/`);

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

  await page.goto(`/n/${namespace}/explorer/tree/`);

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

  await page.goto(`/n/${namespace}/explorer/tree/`);
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

  await page.goto(`/n/${namespace}/explorer/tree/`);
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

test(`it is possible to delete a file (and it will be removed from cache)`, async ({
  page,
}) => {
  /* prepare data */
  const service = {
    name: "mynewservice.yaml",
    image: "bash",
    scale: 2,
    size: "medium",
    cmd: "hello",
  };

  await createService(namespace, service);

  /* visit explorer */
  await page.goto(`/n/${namespace}/explorer/tree/`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();

  await expect(
    page.getByTestId(`explorer-item-${service.name}`),
    "it renders the node in the explorer"
  ).toBeVisible();

  /* open file to load it into useQuery cache */
  await page.getByTestId(`explorer-item-${service.name}`).click();

  await expect(page.getByLabel("Image")).toHaveValue(service.image);
  await expect(page.getByLabel("Cmd")).toHaveValue(service.cmd);

  /* delete file */
  await page.getByTestId("breadcrumb-namespace").click();
  await page
    .getByTestId(`explorer-item-${service.name}`)
    .getByTestId("dropdown-trg-node-actions")
    .click();

  await page.getByTestId("node-actions-delete").click();

  await expect(
    page.getByText(
      `Are you sure you want to delete ${service.name}? This cannot be undone.`,
      {
        exact: true,
      }
    )
  ).toBeVisible();

  await page.getByTestId("node-delete-confirm").click();

  /* assert file is deleted */
  await expect(
    page.getByTestId(`explorer-item-${service.name}`),
    "it does not render the old file"
  ).toHaveCount(0);

  const nodeExists = await checkIfFileExists({
    namespace,
    path: `/${service.name}`,
  });

  await expect(nodeExists).toBeFalsy();

  /* create new file with the same name */
  await page.getByTestId("dropdown-trg-new").first().click();
  await page.getByRole("button", { name: "New Service" }).click();
  await page.getByPlaceholder("service-name.yaml").fill(service.name);
  await page.getByTestId("new-workflow-submit").click();

  /* assert form is empty (cache was cleared) */
  await expect(page.getByLabel("Image")).toHaveValue("");
  await expect(page.getByLabel("Cmd")).toHaveValue("");
});

test("it is not possible to navigate to a workflow that does not exist", async ({
  page,
}) => {
  await page.goto(
    `/n/${namespace}/explorer/workflow/edit/this-file-does-not-exists.yaml`
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
  await page.goto(`/n/${namespace}/explorer/tree/this-folder-does-not-exist`);

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
  await page.goto(`/n/this-namespace-does-not-exist/explorer/tree`);

  await expect(page.getByTestId("error-title")).toContainText("404");
  await expect(page.getByTestId("error-message")).toContainText(
    "The resource you are trying to access does not exist. This might be due to a typo in the  URL or the resource might have been deleted or renamed."
  );

  await expect(
    page.getByRole("link", { name: "Explorer" }),
    "it does not show the main navigation"
  ).not.toBeVisible();
});

// the test for filters works locally, but has issues in CI,
// instead of filtered items, it seems to detect every item or none
// we will address this in the ticket: https://linear.app/direktiv/issue/DIR-1696/fix-for-e2e-test-which-is-failing-in-ci

/*
test("it is possible to filter the file list by name", async ({ page }) => {
  // mock namespace with a list of files
  await page.route(`/api/v2/namespaces/${namespace}/files/`, async (route) => {
    if (route.request().method() === "GET") {
      const json = {
        data: {
          path: "/",
          type: "directory",
          createdAt: "2024-06-03T09:13:12.404617Z",
          updatedAt: "2024-06-03T09:13:12.404617Z",
          children: [
            {
              path: "/important-directory",
              type: "directory",
              createdAt: "2024-06-04T10:29:31.446876Z",
              updatedAt: "2024-06-04T10:29:31.446876Z",
            },
            {
              path: "/other-directory",
              type: "directory",
              createdAt: "2024-06-04T10:29:12.849234Z",
              updatedAt: "2024-06-12T11:46:57.524557Z",
            },
            {
              path: "/important-workflow.yaml",
              type: "workflow",
              size: 377,
              mimeType: "application/yaml",
              createdAt: "2024-06-03T09:14:20.838079Z",
              updatedAt: "2024-06-03T09:14:20.838079Z",
            },
            {
              path: "/other-workflow.yaml",
              type: "workflow",
              size: 377,
              mimeType: "application/yaml",
              createdAt: "2024-06-03T09:45:10.452797Z",
              updatedAt: "2024-06-03T11:27:05.535781Z",
            },
            {
              path: "/test.yaml",
              type: "workflow",
              size: 378,
              mimeType: "application/yaml",
              createdAt: "2024-06-03T09:13:29.215518Z",
              updatedAt: "2024-06-12T09:23:23.208635Z",
            },
          ],
        },
      };
      await route.fulfill({ json });
    } else route.continue();
  });
*/
/* 
     Note for future uses: 
     The route for files needs a '/' at the end
     because the '/' is actually the beginning of the path
     see also: src/api/files/query/file.ts
  */
/*
  const filter = page.getByTestId("queryField");

  // visit page and make sure explorer is loaded
  await page.goto(`/n/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a testing namespace is loaded in the explorer"
  ).toHaveText(namespace);

  await expect(
    page.locator("tr"),
    "it renders all the elements in the list"
  ).toHaveCount(5);

  filter.click();
  await filter.fill("important");

  await expect(
    page.locator("tr"),
    "it renders two elements for this query"
  ).toHaveCount(2);

  filter.click();
  page.keyboard.press("Delete");
  await filter.fill("yaml");

  await expect(
    page.locator("tr"),
    "it renders three elements for this query"
  ).toHaveCount(3);

  filter.click();
  page.keyboard.press("Delete");
  await filter.fill("test");

  await expect(
    page.locator("tr"),
    "it renders one element for this query"
  ).toHaveCount(1);

  await page.reload({
    waitUntil: "networkidle",
  });

  await expect(
    page.locator("tr"),
    "it renders all the elements in the list"
  ).toHaveCount(5);
});
*/
