import { BlockType as Block } from "../../../../schema/blocks";
import { FormEvent } from "react";
import { LocalVariablesContent } from "../../../primitives/Variable/VariableContext";

const keySeparator = "::";

export const encodeBlockKey = (
  blockType: Block["type"],
  elementId: string,
  optional: boolean
) =>
  [blockType, elementId, optional ? "optional" : "required"].join(keySeparator);

const decodeBlockKey = (blockKey: string) => {
  const [blockType, elementId, optional] = blockKey.split(keySeparator, 3);
  if (!blockType || !elementId || !optional) {
    return null;
  }
  return [
    blockType as Block["type"],
    elementId,
    optional === "optional" ? true : false,
  ] as const;
};

const resolveFormValue = (
  blockType: Block["type"],
  value: FormDataEntryValue
) => {
  switch (blockType) {
    case "form-checkbox":
      return value === "true";
    case "form-number-input":
      return parseFloat(String(value));
    case "form-string-input":
    case "form-date-input":
    case "form-select":
    case "form-textarea":
      return value;
    default:
      throw new Error("block type is not a valid form element");
  }
};

const isFormFieldMissing = (
  blockType: Block["type"],
  value: FormDataEntryValue,
  optional: boolean
) => {
  if (optional) {
    return false;
  }

  if (blockType === "form-checkbox") {
    return value === "false";
  }

  return !value;
};

type createLocalFormVariablesReturnType = {
  formVariables: LocalVariablesContent;
  missingRequiredFields: string[];
};

/**
 * Transforms a form submission event into local variables accessible within the page.
 *
 * Extracts form data and converts it into a structured object with proper type resolution
 * based on the form element types. The returned object can be referenced in templates
 * using the `this` namespace.
 *
 * Returns an object structure like:
 *
 * {
 *   "username": "john_doe",      // from the "form-string-input::username" element
 *   "age": 25,                   // from the "form-number-input::age" element
 *   "isActive": true             // from thw "form-checkbox::isActive" element
 * }
 *
 * To eventually be used as template string: {{this.username}}
 */

export const createLocalFormVariables = (
  formEvent: FormEvent<HTMLFormElement>
): createLocalFormVariablesReturnType => {
  const formData = new FormData(formEvent.currentTarget);
  const formValues = Object.fromEntries(formData.entries());
  const missingRequiredFields: string[] = [];

  const transformedEntries = Object.entries(formValues).map(
    ([serializedKey, value]) => {
      const decodedKey = decodeBlockKey(serializedKey);
      if (!decodedKey) {
        throw new Error(`could not decode key: ${serializedKey || "empty"}`);
      }
      const [blockType, elementId, optional] = decodedKey;
      const resolvedValue = resolveFormValue(blockType, value);
      if (isFormFieldMissing(blockType, value, optional)) {
        missingRequiredFields.push(elementId);
      }
      return [elementId, resolvedValue];
    }
  );

  const formVariables = Object.fromEntries(transformedEntries);

  return { formVariables, missingRequiredFields };
};

const isFormElement = (element: Element) =>
  element instanceof HTMLInputElement ||
  element instanceof HTMLSelectElement ||
  element instanceof HTMLTextAreaElement;

/**
 * Extracts formKeys from a collection of HTML form control elements.
 *
 * We rely on createLocalFormVariables to parse the filled out form's keys and values
 * on submit. ExtractFormKeys returns all element names from the form's child elements
 * that match our field encoding syntax.
 */

export const extractFormKeys = (elements: HTMLFormControlsCollection) => {
  const formElementNames = Array.from(elements)
    .filter((element) => isFormElement(element))
    .map((element) => element.name);

  const formKeys = formElementNames.reduce<Record<string, unknown>>(
    (acc, field) => {
      const decodedKey = decodeBlockKey(field);
      if (decodedKey) {
        const [, elementId] = decodedKey;
        acc[elementId] = "";
      }

      return acc;
    },
    {}
  );

  return formKeys;
};
