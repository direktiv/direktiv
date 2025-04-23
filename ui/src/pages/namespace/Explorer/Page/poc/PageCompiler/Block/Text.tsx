import { TextType } from "../../schema/blocks/text";

type TextProps = {
  blockProps: TextType;
};

export const Text = ({ blockProps: { label } }: TextProps) => <p>{label}</p>;
