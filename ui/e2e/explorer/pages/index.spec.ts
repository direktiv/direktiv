import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create a page and view the result in the page editor", async ({
  page,
}) => {
  /* prepare data */
  const filename = "mypage.yaml";
  const headlinevalue = "my-headline";
  const textvalue = "my-text";

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create page */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Page" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("page-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the page and opens the file in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/page/${filename}`);

  const dragarea = page.getByTestId("editor-dragArea");

  /* add 'headline' block per drag and drop */
  await page
    .getByRole("button", { name: "headline" })
    .dragTo(page.getByTestId("dropzone"));

  /* fill in values for headline block */
  await page.getByPlaceholder("Headline").fill("my-headline");
  await page.getByRole("combobox").click();
  page.getByLabel("h3").click();
  await dragarea.getByRole("button", { name: "save" }).click();

  /* add 'text' block per drag and drop */
  await page
    .getByRole("button", { name: "Text", exact: true })
    .dragTo(page.getByTestId("dropzone").nth(1));

  /* fill in values for text block */
  await page.getByPlaceholder("Enter text here...").fill("my-text");
  await dragarea.getByRole("button", { name: "save" }).click();

  const liveMode = page.getByTestId("live");
  const headline = page.locator("h3", { hasText: "my-headline" });
  const text = page.locator("p", { hasText: "my-text" });

  /* inspect page in editor in 'live' mode */
  await liveMode.click();
  await expect(headline, "the headline is visible on the page").toBeVisible();
  await expect(text, "the text is visible on the page").toBeVisible();
  await expect(headline, "the headline contains the edited value").toHaveText(
    headlinevalue
  );
  await expect(text, "the text contains the edited value").toHaveText(
    textvalue
  );
});

test("it is possible to create a page and view the result in a gateway route", async ({
  page,
}) => {
  /* prepare data */
  const filename = "mypage.yaml";
  const routename = "myroute.yaml";
  const pathname = "mypath";
  const headlinevalue = "my-headline";
  const textvalue = "my-text";

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create page */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Page" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("page-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  await expect(
    page,
    "it creates the page and opens the file in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/page/${filename}`);

  const dragarea = page.getByTestId("editor-dragArea");

  /* add 'headline' block per drag and drop */
  await page
    .getByRole("button", { name: "headline" })
    .dragTo(page.getByTestId("dropzone"));

  /* fill in values for headline block */
  await page.getByPlaceholder("Headline").fill("my-headline");
  await page.getByRole("combobox").click();
  page.getByLabel("h3").click();
  await dragarea.getByRole("button", { name: "save" }).click();

  /* add 'text' block per drag and drop */
  await page
    .getByRole("button", { name: "Text", exact: true })
    .dragTo(page.getByTestId("dropzone").nth(1));

  /* fill in values for text block */
  await page.getByPlaceholder("Enter text here...").fill("my-text");
  await dragarea.getByRole("button", { name: "save" }).click();

  const headline = page.locator("h3");
  const text = page.locator("p");

  await page.getByRole("button", { name: "save" }).click();

  /* create a gateway route */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Gateway" }).click();
  await page.getByRole("button", { name: "Route" }).click();

  await page.getByPlaceholder("route-name.yaml").fill(routename);
  await page.getByRole("button", { name: "Create" }).click();
  await expect(
    page,
    "it creates the route and opens the file in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/endpoint/${routename}`);

  /* set up gateway route */
  await page.getByRole("textbox", { name: "path" }).fill(pathname);
  await page.getByRole("checkbox", { name: "get" }).click();
  await page.getByRole("switch", { name: "allow anonymous" }).click();
  await page.getByRole("button", { name: "set target plugin" }).click();
  await expect(page.getByText("Configure target plugin")).toBeVisible();
  await page.getByRole("combobox").click();
  await page.getByRole("option", { name: "Page" }).click();
  await page.getByRole("button", { name: "Browse Files" }).click();
  await expect(
    page.getByRole("button", { name: "mypage.yaml" }),
    "The created page is selectable"
  ).toBeVisible;
  await page.getByRole("button", { name: "mypage.yaml" }).click();
  await page.getByRole("button", { name: "save" }).click();

  const routeURL = `${process.env.PLAYWRIGHT_UI_BASE_URL}/ns/${namespace}/${pathname}`;

  /* go to gateway route */
  await page.goto(routeURL, {
    waitUntil: "networkidle",
  });

  /* inspect page in the gateway route */
  await expect(headline, "the headline is visible on the page").toBeVisible();
  await expect(text, "the text is visible on the page").toBeVisible();
  await expect(headline, "the headline contains the edited value").toHaveText(
    headlinevalue
  );
  await expect(text, "the text contains the edited value").toHaveText(
    textvalue
  );
});
