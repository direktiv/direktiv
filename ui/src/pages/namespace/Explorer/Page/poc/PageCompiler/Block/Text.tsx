import { BlockWrapper } from "./utils/BlockWrapper";
import { TextType } from "../../schema/blocks/text";

export const Text = ({ label }: TextType) => (
  <BlockWrapper>
    <p>{label}</p>
  </BlockWrapper>
);
