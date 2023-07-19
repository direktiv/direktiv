import { Page, expect, test } from "@playwright/test";
import {
  noop as basicWorkflow,
  validate as complexWorkflow,
} from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createNamespace, deleteNamespace } from "../../utils/namespace";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";

const getCodeLayoutButtons = async (page: Page) => {
  const codeBtn = await page.getByTestId("editor-layout-btn-code");
  const diagramBtn = await page.getByTestId("editor-layout-btn-diagram");
  const splitVertBtn = await page.getByTestId(
    "editor-layout-btn-splitVertically"
  );
  const splitHorBtn = await page.getByTestId(
    "editor-layout-btn-splitHorizontally"
  );

  return {
    codeBtn,
    diagramBtn,
    splitVertBtn,
    splitHorBtn,
  };
};

test.beforeEach(async () => {
  namespace = await createNamespace();

  workflow = `${faker.system.commonFileName("yaml")}`;
  await createWorkflow({
    payload: basicWorkflow.data,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflow,
    },
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to switch between Code View, Diagram View, Split Vertically and Split Horizontally", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);

  const { codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCodeLayoutButtons(page);

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

test("it will persist the prefered layout selection in local storage", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);

  const { codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCodeLayoutButtons(page);

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

  // still diagram layout after reload
  await page.reload();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(page.getByTestId("workflow-editor")).not.toBeVisible();
  await expect(page.getByTestId("workflow-diagram")).toBeVisible();
});

test("it will update the diagram when the workflow is saved", async ({
  page,
  browserName,
  context,
}) => {
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.exposeFunction("writeToClipboard", async (text: string) => {
    await navigator.clipboard.writeText(text);
  });

  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const { splitVertBtn } = await getCodeLayoutButtons(page);
  await splitVertBtn.click();

  await page.evaluate(
    async () => await navigator.clipboard.writeText("write to clipboard")
  );

  await page.getByTestId("workflow-editor").click();
  await page.keyboard.press("Meta+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.press("Meta+V");

  // TODO: fix meta key

  // await page.keyboard.type(complexWorkflow.data);

  await expect(page.getByTestId("workflow-dedede")).toBeVisible();
});
