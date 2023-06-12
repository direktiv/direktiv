import { Page, expect } from "@playwright/test";

import { faker } from "@faker-js/faker";

const defaultDescription = "A simple 'no-op' state that returns 'Hello world!'";

export const actionEditAndSaveWorkflow = async (
  page: Page,
  updatedText?: string
): Promise<[string, string]> => {
  const description = page.getByText(
    updatedText ? updatedText : defaultDescription
  );
  await description.click();
  // type random text in that textarea which is for description
  const testText = faker.random.alphaNumeric(9);
  await page.type("textarea", testText);

  const textArea = page.getByRole("textbox");

  //now click on Save
  const saveButton = page.getByTestId("workflow-editor-btn-save");
  await saveButton.click();

  // save button should be disabled/enabled while/after the api call
  await expect(
    saveButton,
    "save button should be disabled during the api call"
  ).toBeDisabled();
  await expect(
    saveButton,
    "save button should be enabled after the api call"
  ).toBeEnabled();

  // after saving is completed screen should have those new changed text before/after the page reload
  await expect(
    page.getByText(testText),
    "after saving, screen should have the updated text"
  ).toBeVisible();

  const updatedWorkflow = await textArea.inputValue();
  return [updatedWorkflow, testText];
};

export const actionWaitForSuccessToast = async (page: Page) => {
  const successToast = page.getByTestId("toast-success");
  await expect(
    successToast,
    "success toast should appear after revert action button click"
  ).toBeVisible();
  await page.getByTestId("toast-close").click();
  await expect(
    successToast,
    "success toast should disappear after click toast-close"
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
