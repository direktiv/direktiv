import { BlockWrapper } from "./utils/BlockWrapper";
import { HeadlineType } from "../../schema/blocks/headline";

export const Headline = ({ label, description }: HeadlineType) => (
  <BlockWrapper>
    <h1 className="text-xl">{label}</h1>
    {description && <p>{description}</p>}
  </BlockWrapper>
);
