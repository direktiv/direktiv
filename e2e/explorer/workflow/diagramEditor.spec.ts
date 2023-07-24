import { Page, expect, test } from "@playwright/test";
import { createNamespace, deleteNamespace } from "../../utils/namespace";

import { consumeEvent as consumeEventWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
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
    payload: consumeEventWorkflow.data,
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
  const editor = page.getByTestId("workflow-editor");
  const diagram = page.getByTestId("workflow-diagram");

  const { codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCodeLayoutButtons(page);

  // code is the default view
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).not.toBeVisible();

  // diagram view
  await diagramBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).not.toBeVisible();
  await expect(diagram).toBeVisible();

  // split vertically
  await splitVertBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).toBeVisible();

  // split horizontally
  await splitHorBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("true");
  await expect(editor).toBeVisible();
  await expect(diagram).toBeVisible();

  // back to default
  await codeBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).not.toBeVisible();
});

test("it will change the direction of the diagram when the layout is set to Split Vertically", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const startNode = page.getByTestId("rf__node-startNode");
  const endNode = page.getByTestId("rf__node-endNode");

  const { splitVertBtn, splitHorBtn } = await getCodeLayoutButtons(page);

  // use split horizontally layout
  await splitHorBtn.click();

  await expect(startNode).toBeVisible();
  await expect(endNode).toBeVisible();

  const { x: startNodeXHor, y: startNodeYHor } =
    (await startNode.boundingBox()) || {};
  const { x: endNodeXHor, y: endNodeYHor } =
    (await endNode.boundingBox()) || {};

  if (!startNodeXHor || !startNodeYHor || !endNodeXHor || !endNodeYHor) {
    throw new Error("one of the nodes is not visible");
  }

  expect(
    startNodeYHor,
    "start- and end node are on the same vertical line"
  ).toBe(endNodeYHor);

  expect(
    endNodeXHor,
    "the end node is on the right of the start node"
  ).toBeGreaterThan(startNodeXHor);

  await splitVertBtn.click();
  const { x: startNodeXVert, y: startNodeYVert } =
    (await startNode.boundingBox()) || {};
  const { x: endNodeXVert, y: endNodeYVert } =
    (await endNode.boundingBox()) || {};

  if (!startNodeXHor || !startNodeYVert || !endNodeXVert || !endNodeYVert) {
    throw new Error("one of the nodes is not visible");
  }

  expect(
    startNodeXVert,
    "start- and end node are on the same horizontal line"
  ).toBe(endNodeXVert);

  expect(
    endNodeYVert,
    "the end node is on the right of the start node"
  ).toBeGreaterThan(startNodeYVert);
});

test("it will persist the prefered layout selection in local storage", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const editor = page.getByTestId("workflow-editor");
  const diagram = page.getByTestId("workflow-diagram");

  const { codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCodeLayoutButtons(page);

  // code is the default view
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).not.toBeVisible();

  // diagram view
  await diagramBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).not.toBeVisible();
  await expect(diagram).toBeVisible();

  // still diagram layout after reload
  await page.reload();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).not.toBeVisible();
  await expect(diagram).toBeVisible();
});

test("it will update the diagram when the workflow is saved", async ({
  page,
}) => {
  const editor = page.getByTestId("workflow-editor");
  const diagram = page.getByTestId("workflow-diagram");


  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const { splitVertBtn } = await getCodeLayoutButtons(page);
  await splitVertBtn.click();

  await expect(
    editor.getByText("A simple 'consumeEvent' state that")
  ).toBeVisible();

  await expect(diagram.getByTestId("rf__node-ce")).toBeVisible();
  await expect(diagram.getByTestId("rf__node-greet")).toBeVisible();

  await page.getByTestId("workflow-editor").click();
  await page.keyboard.press("ArrowRight");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowLeft");
  await page.keyboard.press("#");

  await page.getByTestId("workflow-editor-btn-save").click();
  await expect(diagram.getByTestId("rf__node-ce")).toBeVisible();
  await expect(diagram.getByTestId("rf__node-greet")).not.toBeVisible();
});
