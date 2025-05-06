import { TemplateString } from "./utils/TemplateString";
import { TextType } from "../../schema/blocks/text";

type TextProps = {
  blockProps: TextType;
};

export const Text = ({ blockProps }: TextProps) => (
  <p>
    <TemplateString value={blockProps.content} />
  </p>
);
