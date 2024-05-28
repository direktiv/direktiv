import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { createRouteYaml, removeLines } from "./utils";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create a basic route file", async ({ page }) => {
  /* prepare data */
  const filename = "myroute.yaml";

  const expectedYaml = createRouteYaml({
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
    plugins: {
      target: `
    type: instant-response
    configuration:
        status_code: 200`,
    },
  });

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it navigates to the test namespace in the explorer"
  ).toHaveText(namespace);

  /* create route */
  await page.getByRole("button", { name: "New" }).first().click();
  await page.getByRole("menuitem", { name: "Gateway" }).click();
  await page.getByRole("button", { name: "New Route" }).click();

  await expect(page.getByRole("button", { name: "Create" })).toBeDisabled();
  await page.getByPlaceholder("route-name.yaml").fill(filename);
  await page.getByRole("button", { name: "Create" }).click();

  /**
   * close the toast, which covers the save button and prevents
   * us from clicking it (makes this test 4 seconds faster)
   */
  await page.getByTestId("toast-close").click();

  await expect(
    page,
    "it creates the route file and opens it in the explorer"
  ).toHaveURL(`/n/${namespace}/explorer/endpoint/${filename}`);

  /* fill out form */
  await page.getByLabel("path").fill("path");
  await page.getByLabel("timeout").fill("3000");
  await page.getByLabel("GET").click();
  await page.getByLabel("POST").click();

  /* try to save incomplete form */
  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByText("plugins : this field is invalid"),
    "it can not save the route without a valid target plugin"
  ).toBeVisible();

  await page.getByRole("button", { name: "set target plugin" }).click();

  /* add an empty instant response plugin */
  await page.getByRole("combobox").click();
  await page.getByLabel("Instant Response").click();
  await page.getByRole("button", { name: "Save" }).click();

  /* check editor content */
  const editor = page.locator(".lines-content");
  await expect(
    editor,
    "all entered data is represented in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });

  /* note: saving the plugin should have saved the whole file. */
  await expect(
    page.getByText("unsaved changes"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  /* reload */
  await page.reload({ waitUntil: "networkidle" });

  await expect(
    editor,
    "after reloading, the entered data is still in the editor preview"
  ).toContainText(expectedYaml, { useInnerText: true });

  page.getByRole("link", { name: "Open Logs" }).click();

  await expect(
    page,
    "when the open logs link is clicked, page should navigate to the route detail page"
  ).toHaveURL(`/n/${namespace}/gateway/routes/${filename}`);
});

test("it is possible to add plugins to a route file", async ({ page }) => {
  /* prepare data */
  const filename = "myroute.yaml";
  const editor = page.locator(".lines-content");

  type CreateRouteYamlParam = Parameters<typeof createRouteYaml>[0];
  const minimalRouteConfig: Omit<CreateRouteYamlParam, "plugins"> = {
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
  };

  const basicTargetPlugin = `
    type: instant-response
    configuration:
      status_code: 200`;

  const initialRouteYaml = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await createFile({
    namespace,
    name: filename,
    type: "endpoint",
    yaml: initialRouteYaml,
  });

  await page.goto(`/n/${namespace}/explorer/endpoint/${filename}`, {
    waitUntil: "networkidle",
  });

  /* configure inbound plugin: ACL */
  await page.getByRole("button", { name: "add inbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Access control list (acl)").click();

  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .fill("allow this group 1");

  /* submit via enter */
  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .press("Enter");

  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByPlaceholder("Enter a group")
    .nth(1)
    .fill("allow this group 2");

  /* submit via button */
  await page
    .locator("fieldset")
    .filter({ hasText: "Allow Groups (optional)" })
    .getByRole("button")
    .nth(1)
    .click();

  await page.getByRole("button", { name: "Save" }).click();

  /* configure inbound plugin: Request Convert */
  await page.getByRole("button", { name: "add inbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Request Convert").click();
  await page.getByText("Omit Queries").click();
  await page.getByText("Omit Consumer").click();
  await page.getByRole("button", { name: "Save" }).click();

  /* check editor content */
  const inboundPluginsBeforeSorting = `
    - type: acl
      configuration:
        allow_groups:
          - allow this group 1
          - allow this group 2
        deny_groups: []
        allow_tags: []
        deny_tags: []
    - type: request-convert
      configuration:
        omit_headers: false
        omit_queries: true
        omit_body: false
        omit_consumer: true`;

  const inboundPluginsAfterSorting = `
    - type: request-convert
      configuration:
        omit_headers: false
        omit_queries: true
        omit_body: false
        omit_consumer: true
    - type: acl
      configuration:
        allow_groups:
          - allow this group 1
          - allow this group 2
        deny_groups: []
        allow_tags: []
        deny_tags: []`;

  let expectedEditorContent = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
      inbound: inboundPluginsBeforeSorting,
    },
  });

  /**
   * Note: the editor only shows a limited amount of lines, The Editor uses a
   * virtualized list to render the content. This means that the invisible content
   * is not even rendered in the DOM. So from now on we have to crop some lines
   * in our assertions to make them pass. This is not a big problem, because we
   * already tested the upper part of the file in the previous test.
   *
   * We will scroll the editor to the very bottom, now. The editor will automatically
   * keep that scroll position when we change the content.
   */
  await page.evaluate(() => {
    document
      .querySelector(".monaco-editor .monaco-scrollable-element")
      ?.scrollBy(0, 100000000);
  });

  await expect(
    editor,
    "the inbound plugins are represented in the editor preview"
  ).toContainText(removeLines(expectedEditorContent, 4, "top"), {
    useInnerText: true,
  });

  /* change sorting of inbound plugins */
  await page
    .getByRole("row", { name: "Access control list (acl)" })
    .getByRole("button")
    .click();
  await page.getByRole("button", { name: "Move down" }).click();

  expectedEditorContent = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
      inbound: inboundPluginsAfterSorting,
    },
  });

  await expect(
    editor,
    "the new inbound plugin order is represented in the editor preview"
  ).toContainText(removeLines(expectedEditorContent, 4, "top"), {
    useInnerText: true,
  });

  /* configure outbound plugin: JavaScript */
  await page.getByRole("button", { name: "add outbound plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("JavaScript").click();
  await page.getByRole("textbox").fill("// execute some JavaScript here");
  await page.getByRole("button", { name: "Save" }).click();

  /* check editor content */
  const outboundPlugins = `
    - type: js-outbound
      configuration:
        script: // execute some JavaScript here`;

  expectedEditorContent = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
      inbound: inboundPluginsAfterSorting,
      outbound: outboundPlugins,
    },
  });

  await expect(
    editor,
    "the outbound plugin is represented in the editor preview"
  ).toContainText(removeLines(expectedEditorContent, 7, "top"), {
    useInnerText: true,
  });

  /* configure auth plugin: Github Webhook */
  await page.getByRole("button", { name: "add auth plugin" }).click();
  await page.getByRole("combobox").click();
  await page.getByLabel("Github Webhook").click();
  await page.getByLabel("secret").fill("my github secret");
  await page.getByRole("button", { name: "Save" }).click();

  /* check editor content */
  const authPlugins = `
    - type: github-webhook-auth
      configuration:
        secret: my github secret`;

  expectedEditorContent = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
      inbound: inboundPluginsAfterSorting,
      outbound: outboundPlugins,
      auth: authPlugins,
    },
  });

  await expect(
    editor,
    "the auth plugin is represented in the editor preview"
  ).toContainText(removeLines(expectedEditorContent, 10, "top"), {
    useInnerText: true,
  });

  /* note: saving the plugin should have saved the whole file. */
  await expect(
    page.getByText("unsaved changes"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  /* reload */
  await page.reload({ waitUntil: "networkidle" });
  await expect(
    editor,
    "after reloading, the entered data is still in the editor preview"
  ).toContainText(removeLines(expectedEditorContent, 9, "bottom"), {
    useInnerText: true,
  });

  /* delete all optional plugins */
  await page
    .getByRole("row", { name: "Access control list (acl)" })
    .getByRole("button")
    .click();
  await page.getByRole("button", { name: "Delete" }).click();

  await page
    .getByRole("row", { name: "Request convert" })
    .getByRole("button")
    .click();
  await page.getByRole("button", { name: "Delete" }).click();

  await page
    .getByRole("row", { name: "Javascript" })
    .getByRole("button")
    .click();
  await page.getByRole("button", { name: "Delete" }).click();

  await page
    .getByRole("row", { name: "Github Webhook" })
    .getByRole("button")
    .click();
  await page.getByRole("button", { name: "Delete" }).click();

  expectedEditorContent = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await expect(
    editor,
    "the deleted plugins are also represented in the editor preview"
  ).toContainText(expectedEditorContent, {
    useInnerText: true,
  });
});

test("it blocks navigation when there are unsaved changes", async ({
  page,
}) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    dialogTriggered = true;
    return dialog.dismiss();
  });

  /* prepare data */
  const filename = "formatroute.yaml";

  type CreateRouteYamlParam = Parameters<typeof createRouteYaml>[0];
  const minimalRouteConfig: Omit<CreateRouteYamlParam, "plugins"> = {
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
  };

  const basicTargetPlugin = `
    type: instant-response
    configuration:
      status_code: 200`;

  const initialRouteYaml = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await createFile({
    namespace,
    name: filename,
    type: "endpoint",
    yaml: initialRouteYaml,
  });

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/endpoint/${filename}`, {
    waitUntil: "networkidle",
  });

  await page.getByLabel("path").fill("new_path");

  await expect(
    page.getByText("unsaved changes"),
    "it renders a hint that there are unsaved changes"
  ).toBeVisible();

  await page.getByRole("link", { name: "Monitoring" }).click();

  await expect(
    dialogTriggered,
    "it triggers a confirmation dialogue"
  ).toBeTruthy();

  await expect(page, "it does not navigate away from the page").toHaveURL(
    `/n/${namespace}/explorer/endpoint/${filename}`
  );
});

test("it does not block navigation when only formatting has changed", async ({
  page,
}) => {
  let dialogTriggered = false;

  page.on("dialog", async (dialog) => {
    dialogTriggered = true;
    return dialog.dismiss();
  });

  /* prepare data */
  const filename = "formatroute.yaml";

  type CreateRouteYamlParam = Parameters<typeof createRouteYaml>[0];
  const minimalRouteConfig: Omit<CreateRouteYamlParam, "plugins"> = {
    path: "path",
    timeout: 3000,
    methods: ["GET", "POST"],
  };

  const basicTargetPlugin = `
    type: "instant-response"
    configuration:
      status_code: 200`;

  const initialRouteYaml = createRouteYaml({
    ...minimalRouteConfig,
    plugins: {
      target: basicTargetPlugin,
    },
  });

  await createFile({
    namespace,
    name: filename,
    type: "endpoint",
    yaml: initialRouteYaml,
  });

  /* visit page */
  await page.goto(`/n/${namespace}/explorer/endpoint/${filename}`, {
    waitUntil: "networkidle",
  });

  const formattedText = page
    .getByRole("code")
    .locator("div")
    .filter({ hasText: "type: instant-response" })
    .nth(4);

  await expect(
    formattedText,
    "it has updated the formatting (removed quotes)"
  ).toBeVisible();

  await expect(
    page.getByText("unsaved changes"),
    "it does not render a hint that there are unsaved changes"
  ).not.toBeVisible();

  await page.getByRole("link", { name: "Monitoring" }).click();

  await expect(
    dialogTriggered,
    "it does not trigger a warning dialogue"
  ).toBeFalsy();

  await expect(
    page.getByRole("heading", { name: "Monitoring", exact: true }),
    "it is possible to leave the route"
  ).toBeVisible();
});
