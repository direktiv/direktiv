import { Block } from "..";
import { BlocksWrapper } from "../utils/BlocksWrapper";
import { QueryProviderType } from "../../../schema/blocks/queryProvider";

export const QueryProvider = ({ blocks }: QueryProviderType) => (
  <BlocksWrapper>
    {blocks.map((block, index) => (
      <Block key={index} block={block} />
    ))}
  </BlocksWrapper>
);
