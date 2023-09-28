import { Page } from "@playwright/test";

export const getCommonElements = (page: Page) => {
  const queryInput = page.getByTestId("jq-query-input");
  const btnRun = page.getByTestId("jq-run-btn");
  const outputTextArea = page
    .getByTestId("jq-output-editor")
    .getByRole("textbox");

  const inputTextContainer = page.getByTestId("jq-input-editor");
  const inputTextArea = inputTextContainer.getByRole("textbox");

  return {
    queryInput,
    btnRun,
    outputTextArea,
    inputTextContainer,
    inputTextArea,
  };
};

export const getErrorContainer = (page: Page) => {
  const errorContainer = page.getByTestId("form-errors");
  return { errorContainer };
};
