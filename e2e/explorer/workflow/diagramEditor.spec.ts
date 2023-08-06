import { Page, expect, test } from "@playwright/test";
import { createNamespace, deleteNamespace } from "../../utils/namespace";

import { consumeEvent as consumeEventWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createRevision } from "~/api/tree/mutate/createRevision";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { headers } from "e2e/utils/testutils";

let namespace = "";
let workflow = "";

const getCommonPageElements = (page: Page) => {
  const isActiveWorkflowPage = page.url().includes("workflow/active/");
  const editor = page.getByTestId(
    isActiveWorkflowPage ? "workflow-editor" : "revisions-detail-editor"
  );
  const diagram = page.getByTestId("workflow-diagram");
  const codeBtn = page.getByTestId("editor-layout-btn-code");
  const diagramBtn = page.getByTestId("editor-layout-btn-diagram");
  const splitVertBtn = page.getByTestId("editor-layout-btn-splitVertically");
  const splitHorBtn = page.getByTestId("editor-layout-btn-splitHorizontally");

  return {
    editor,
    diagram,
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
    headers,
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

  const { editor, diagram, codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCommonPageElements(page);

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

test("it will change the direction of the diagram, when the layout is set to Split Vertically", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const startNode = page.getByTestId("rf__node-startNode");
  const endNode = page.getByTestId("rf__node-endNode");

  const { splitVertBtn, splitHorBtn } = await getCommonPageElements(page);

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

test("it will persist the preferred layout selection in local storage", async ({
  page,
}) => {
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const { editor, diagram, codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCommonPageElements(page);

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
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  const { editor, diagram, splitVertBtn } = await getCommonPageElements(page);

  await splitVertBtn.click();

  await expect(
    editor.getByText(
      "A simple 'consumeEvent' state that listens for the greetingcloudevent generated from the template 'generate-event'."
    ),
    "the description of the workflow is visible in the code editor"
  ).toBeVisible();

  await expect(
    diagram.getByTestId("rf__node-ce"),
    "the first state 'id' is shown in the diagram"
  ).toBeVisible();
  await expect(
    diagram.getByTestId("rf__node-greet"),
    "the second state 'greet' is shown in the diagram"
  ).toBeVisible();

  const commentOutNextLine = async () => {
    await page.keyboard.press("ArrowDown");
    await page.keyboard.press("ArrowLeft");
    await page.keyboard.press("#");
  };

  // comment out all states expect the first one to force the diagram to update
  await page.getByTestId("workflow-editor").click();
  // cursor is at the end of line 8, use right arrow to go to the first column of line 9
  await page.keyboard.press("ArrowRight");
  await page.keyboard.press("ArrowDown"); // line 10, column 1
  await page.keyboard.press("ArrowDown"); // line 11, column 1
  await page.keyboard.press("ArrowDown"); // line 12, column 1
  await page.keyboard.press("#");
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();
  await commentOutNextLine();

  // save changes
  await page.getByTestId("workflow-editor-btn-save").click();

  await expect(
    diagram.getByTestId("rf__node-ce"),
    "the first state 'id' is still shown in the diagram"
  ).toBeVisible();
  await expect(
    diagram.getByTestId("rf__node-greet"),
    "the second state 'greet' was commented out and the diagram updated to not show it anymore"
  ).not.toBeVisible();
});

test("it is possible to use the diagram view on the revisions detail page as well", async ({
  page,
}) => {
  const {
    revision: { name: revisionName },
  } = await createRevision({
    payload: consumeEventWorkflow,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: workflow,
    },
    headers,
  });

  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=${revisionName}`
  );

  const { editor, diagram, codeBtn, diagramBtn, splitVertBtn, splitHorBtn } =
    await getCommonPageElements(page);

  // code is the default view
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).not.toBeVisible();

  // split vertically
  await splitVertBtn.click();
  expect(await codeBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await diagramBtn.getAttribute("aria-pressed")).toBe("false");
  expect(await splitVertBtn.getAttribute("aria-pressed")).toBe("true");
  expect(await splitHorBtn.getAttribute("aria-pressed")).toBe("false");
  await expect(editor).toBeVisible();
  await expect(diagram).toBeVisible();
});
