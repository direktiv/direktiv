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

  // TODO: can we get rid of this type assertion?
  return [elementType as Block["type"], elementId] as const;
};

export const createFormContextVariables = (
  e: FormEvent<HTMLFormElement>,
  formName: string
): InjectedVariables => {
  const formData = new FormData(e.currentTarget);
  const formValues = Object.fromEntries(formData.entries());

  const transformedEntries = Object.entries(formValues).map(
    ([serializedKey, value]) => {
      const [, elementId] = decodeElementKey(serializedKey);
      return [elementId, value];
    }
  );

  const processedFormValues = Object.fromEntries(transformedEntries);

  return { form: { [formName]: processedFormValues } };
};
