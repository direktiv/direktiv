import { BlockType as Block } from "../../../../schema/blocks";

const separator = "::";

export const serializeFieldName = (
  elementType: Block["type"],
  elementId: string
) => [elementType, elementId].join(separator);

export const deserializeFieldName = (elementName: string) => {
  const [elementType, elementId] = elementName.split(separator, 2);

  if (!elementType || !elementId) throw new Error("invalid form element name");

  // TODO: can we get rid of this type assertion?
  return [elementType as Block["type"], elementId] as const;
};
