import { BlockType as Block } from "../../../../schema/blocks";
import { FormEvent } from "react";
import { LocalVariablesContent } from "../../../primitives/Variable/LocalVariables";

const keySeparator = "::";

export const encodeElementKey = (
  elementType: Block["type"],
  elementId: string
) => [elementType, elementId].join(keySeparator);

const decodeElementKey = (elementName: string) => {
  const [elementType, elementId] = elementName.split(keySeparator, 2);
  if (!elementType || !elementId) throw new Error("invalid form element name");
  return [elementType as Block["type"], elementId] as const;
};

const resolveFormValue = (
  blockType: Block["type"],
  value: FormDataEntryValue
) => {
  switch (blockType) {
    case "form-checkbox":
      return value === "true";
    case "form-number-input":
      return parseInt(String(value));
    case "form-string-input":
    case "form-date-input":
    case "form-select":
    case "form-textarea":
      return value;
    default:
      throw new Error("invalid form element type");
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
      const [blockType, elementId] = decodeElementKey(serializedKey);
      const resolvedValue = resolveFormValue(blockType, value);
      return [elementId, resolvedValue];
    }
  );

  const processedFormValues = Object.fromEntries(transformedEntries);

  return { ["form"]: processedFormValues };
};
