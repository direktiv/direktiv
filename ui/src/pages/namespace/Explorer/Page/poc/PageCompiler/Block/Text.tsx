import { TemplateString } from "../primitives/TemplateString";
import { TextType } from "../../schema/blocks/text";

type TextProps = {
  blockProps: TextType;
};

export const Text = ({ blockProps }: TextProps) => (
  <div>
    <TemplateString value={blockProps.content} />
  </div>
);
