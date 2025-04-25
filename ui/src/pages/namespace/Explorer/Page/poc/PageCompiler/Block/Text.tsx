import { TemplateString } from "./utils/TemplateString";
import { TextType } from "../../schema/blocks/text";

type TextProps = {
  blockProps: TextType;
};

export const Text = ({ blockProps: { label } }: TextProps) => (
  <p>
    <TemplateString value={label} />
  </p>
);
