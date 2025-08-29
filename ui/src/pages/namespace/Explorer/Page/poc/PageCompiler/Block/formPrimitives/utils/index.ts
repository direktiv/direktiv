import { BlockType as Block } from "../../../../schema/blocks";
import { FormEvent } from "react";
import { LocalVariablesContent } from "../../../primitives/Variable/LocalVariables";

const keySeparator = "::";

export const encodeBlockKey = (blockType: Block["type"], elementId: string) =>
  [blockType, elementId].join(keySeparator);

const decodeBlockKey = (blockKey: string) => {
  const [blockType, elementId] = blockKey.split(keySeparator, 2);
  if (!blockType || !elementId) throw new Error("invalid form element name");
  return [blockType as Block["type"], elementId] as const;
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

/**
 * Transforms a form submission event into local variables accessible within the page.
 *
 * Extracts form data and converts it into a structured object with proper type resolution
 * based on the form element types. The returned object can be referenced in templates
 * using the `this.form` namespace.
 *
 * Returns an object structure like:
 *
 * {
 *   "form": {
 *     "username": "john_doe",      // from the "form-string-input::username" element
 *     "age": 25,                   // from the "form-number-input::age" element
 *     "isActive": true             // from thw "form-checkbox::isActive" element
 *   }
 * }
 *
 * To eventually be used as template string: {{this.form.username}}
 */
export const createLocalFormVariables = (
  formEvent: FormEvent<HTMLFormElement>
): LocalVariablesContent => {
  const formData = new FormData(formEvent.currentTarget);
  const formValues = Object.fromEntries(formData.entries());

  const transformedEntries = Object.entries(formValues).map(
    ([serializedKey, value]) => {
      const [blockType, elementId] = decodeBlockKey(serializedKey);
      const resolvedValue = resolveFormValue(blockType, value);
      return [elementId, resolvedValue];
    }
  );

  const processedFormValues = Object.fromEntries(transformedEntries);

  return { ["form"]: processedFormValues };
};
