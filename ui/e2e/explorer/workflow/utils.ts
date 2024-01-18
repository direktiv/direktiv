import { Page, expect } from "@playwright/test";

export const waitForSuccessToast = async (page: Page) => {
  const successToast = page.getByTestId("toast-success");
  await expect(successToast, "a success toast appears").toBeVisible();
  await page.getByTestId("toast-close").click();
  await expect(
    successToast,
    "success toast disappears after clicking toast-close"
  ).toBeHidden();
};

export const actionRevertRevision = async (page: Page) => {
  const revisionTrigger = page.getByTestId("workflow-editor-btn-revision-drop");
  await revisionTrigger.click();
  const revertRevisionButton = page.getByTestId(
    "workflow-editor-btn-revert-revision"
  );
  await revertRevisionButton.click();
};

export const actionMakeRevision = async (page: Page) => {
  const makeRevisionButton = page.getByTestId(
    "workflow-editor-btn-make-revision"
  );
  await makeRevisionButton.click();
};

export const actionDeleteRevision = async (page: Page, revision: string) => {
  const menuTrg = page.getByTestId(
    `workflow-revisions-item-menu-trg-${revision}`
  );
  await menuTrg.click();
  const deleteTrg = page.getByTestId(
    `workflow-revisions-trg-delete-dlg-${revision}`
  );
  await deleteTrg.click();

  const deleteDialog = page.getByTestId("dialog-delete-revision");
  await expect(
    deleteDialog,
    "after click delete menu, it should show the delete confirm dialog"
  ).toBeVisible();
  const submitButton = page.getByTestId("dialog-delete-revision-btn-submit");
  await submitButton.click();
};

export const jsonSchemaFormWorkflow = `description: A workflow with a complex json schema form'
states:
- id: input
  type: validate
  schema:
    title: some test
    type: object
    required:
    - firstName
    - lastName
    properties:
      firstName:
        type: string
        title: First name
      lastName:
        type: string
        title: Last name
      select:
        title: role
        type: string
        enum: 
          - admin
          - guest
      array:
        title: A list of strings
        type: array
        items:
          type: string
      age:
        type: integer
        title: Age
      file:
        type: string
        title: file upload
        format: data-url`;

export const jsonSchemaWithRequiredEnum = `description: A workflow with a complex json schema form'
states:
- id: input
  type: validate
  schema:
    title: some test
    type: object
    required:
    - firstName
    - lastName
    - select
    properties:
      firstName:
        type: string
        title: First name
      lastName:
        type: string
        title: Last name
      select:
        title: role
        type: string
        enum: 
          - admin
          - guest
      `;
