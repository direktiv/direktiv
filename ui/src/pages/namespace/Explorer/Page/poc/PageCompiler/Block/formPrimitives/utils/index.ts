import { BlockType as Block } from "../../../../schema/blocks";
import { FormEvent } from "react";
import { InjectedVariables } from "../../../primitives/Variable/VariableContext";

const separator = "::";

export const encodeElementKey = (
  elementType: Block["type"],
  elementId: string
) => [elementType, elementId].join(separator);

const decodeElementKey = (elementName: string) => {
  const [elementType, elementId] = elementName.split(separator, 2);
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

export const createFormContextVariables = (
  e: FormEvent<HTMLFormElement>,
  formName: string
): InjectedVariables["form"] => {
  const formData = new FormData(e.currentTarget);
  const formValues = Object.fromEntries(formData.entries());

  const transformedEntries = Object.entries(formValues).map(
    ([serializedKey, value]) => {
      const [blockType, elementId] = decodeElementKey(serializedKey);
      const resolvedValue = resolveFormValue(blockType, value);
      return [elementId, resolvedValue];
    }
  );

  const processedFormValues = Object.fromEntries(transformedEntries);

  return { [formName]: processedFormValues };
};
