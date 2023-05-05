import { createNamespace, deleteNamespace } from "./utils/namespace";
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
  await page.goto(`http://localhost:3000/`);

  await expect(page).toHaveTitle("direktiv.io");

  // at this point, any namespace may be loaded.
  // let's navigate to the test's namespace via breadcrumbs.

  await page.getByRole("main").getByTestId("dropdown-trg-namespace").click();

  await page.getByRole("menuitemradio", { name: namespace }).click();

  await expect(page).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree`
  );

  await expect(page.getByRole("link", { name: namespace })).toBeVisible();
});

test("it is possible to navigate to a namespace directly", async ({ page }) => {
  await page.goto(`http://localhost:3000/${namespace}/explorer/tree`);

  await expect(page.getByRole("link", { name: namespace })).toBeVisible();

  await expect(page).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree`
  );
});

test("it is possible to create and delete a directory", async ({ page }) => {
  await page.goto(`http://localhost:3000/${namespace}/explorer/tree`);

  await expect(page.getByRole("link", { name: namespace })).toBeVisible();

  await expect(page).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree`
  );

  // create folder
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-dir").click();
  await page.getByPlaceholder("folder-name").fill("awesome-folder");
  await page.getByRole("button", { name: "Create" }).click();

  // it automatically navigates to the folder
  await page.getByText("/ awesome-folder").isVisible();
  await expect(
    page,
    "it creates a new folder and navigates to it automatically"
  ).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree/awesome-folder`
  );

  // navigate back to tree root
  await page.getByTestId("tree-root").click();
  await expect(
    page,
    "when clicking the tree icon, it navigates back to the tree root"
  ).toHaveURL(`http://localhost:3000/${namespace}/explorer/tree`);
  await expect(page.getByRole("link", { name: namespace })).toBeVisible();

  // navigate to folder by clicking on it
  await page.getByRole("link", { name: "awesome-folder" }).click();
  await expect(
    page,
    "when clicking on the folder, it navigates to it"
  ).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree/awesome-folder`
  );

  // navigate back by clicking on .. "folder"
  await page.getByRole("link", { name: ".." }).click();
  await expect(
    page,
    "when clicking .. it navigates back to the tree root"
  ).toHaveURL(`http://localhost:3000/${namespace}/explorer/tree`);

  // click delete and confirm
  await page.getByTestId("dropdown-trg-dir-actions").click();
  await page.getByTestId("dir-actions-delete").click();
  await page.getByRole("button", { name: "Delete" }).click();

  await expect(page.getByRole("dialog", { name: "Delete" })).toHaveCount(0);
  await expect(page).toHaveURL(
    `http://localhost:3000/${namespace}/explorer/tree`
  );
  await expect(
    page.getByRole("link", { name: "awesome-folder" }),
    "it deletes the folder"
  ).toHaveCount(0);
});
