import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to switch between Code View, Diagram View, Split Vertically and Split Horizontally", async ({
  page,
}) => {
  const workflow = `${faker.system.commonFileName("yaml")}`;

  await createWorkflow({
    payload: basicWorkflow.data,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflow,
    },
  });

  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`, {});

  const codeBtn = await page.getByTestId("editor-layout-btn-code");
  const diagramBtn = await page.getByTestId("editor-layout-btn-diagram");
  const splitVertBtn = await page.getByTestId(
    "editor-layout-btn-splitVertically"
  );
  const splitHorBtn = await page.getByTestId(
    "editor-layout-btn-splitHorizontally"
  );

  // code is the default view
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(page.getByTestId("workflow-editor")).toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).not.toBeVisible();

  // diagram view
  await diagramBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(page.getByTestId("workflow-editor")).not.toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).toBeVisible();

  // split vertically
  await splitVertBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(page.getByTestId("workflow-editor")).toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).toBeVisible();

  // split horizontally
  await splitHorBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("true");
  await expect(page.getByTestId("workflow-editor")).toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).toBeVisible();

  // back to default
  await codeBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(page.getByTestId("workflow-editor")).toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).not.toBeVisible();
});
