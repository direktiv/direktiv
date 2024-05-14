import { Page } from "@playwright/test";
import exampleSnippets from "~/pages/namespace/JqPlayground/Examples/exampleSnippets";

export const getCommonElements = (page: Page) => {
  const queryInput = page.getByTestId("jq-query-input");
  const btnRun = page.getByTestId("jq-run-btn");
  const outputTextArea = page
    .getByTestId("jq-output-editor")
    .getByRole("textbox");

  const copyInputBtn = page.getByTestId("copy-input-btn");
  const copyOutputBtn = page.getByTestId("copy-output-btn");
  const copyLogsBtn = page.getByTestId("copy-logs-btn");

  const inputTextContainer = page.getByTestId("jq-input-editor");
  const inputTextArea = inputTextContainer.getByRole("textbox");

  const logsTextArea = page.getByTestId("jq-logs-editor").getByRole("textbox");

  return {
    copyInputBtn,
    copyOutputBtn,
    copyLogsBtn,
    queryInput,
    btnRun,
    outputTextArea,
    inputTextContainer,
    inputTextArea,
    logsTextArea,
  };
};

export const userScrolledADecentAmount = async (page: Page) =>
  await page.evaluate(
    /**
     * the value 300 is just a threshhold to make sure the needed to scroll
     * a decent amount to reach the button. Just testing it against 0 would
     * no garantee that there is a fair amount of scrolling involved
     */
    () => window.scrollY > 150
  );

export const scrolledToTheTop = async (page: Page) =>
  await page.evaluate(() => window.scrollY === 0);

type SnippetKeys = (typeof exampleSnippets)[number]["key"];

const objectToPrettifiedString = (object: unknown) =>
  JSON.stringify(object, null, 2);

export const expectedSnippetOutput: Record<SnippetKeys, string> = {
  unchangedInput: objectToPrettifiedString({
    foo: {
      bar: {
        baz: 123,
      },
    },
  }),
  valueAtKey: "42",
  arrayOperation: objectToPrettifiedString({
    good: false,
    name: "XML",
  }),
  arrayObjectConstruction: objectToPrettifiedString([
    {
      title: "JQ Primer",
      user: "stedolan",
    },
    {
      title: "More JQ",
      user: "stedolan",
    },
  ]),
  lengthOfValue: objectToPrettifiedString([2, 6, 1, 0]),
  keysInArray: objectToPrettifiedString(["Foo", "abc", "abcd"]),
  feedInput: objectToPrettifiedString([42, "something else"]),
  pipeOutput: objectToPrettifiedString(["JSON", "XML"]),
  inputUnchanged: objectToPrettifiedString([2, 4, 7]),
  invokeFilter: objectToPrettifiedString([2, 3, 4]),
  conditionals: '"many"',
  stringInterpolation: '"The input was 42, which is one less than 43"',
};
