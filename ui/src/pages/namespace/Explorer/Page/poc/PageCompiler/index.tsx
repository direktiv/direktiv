import { Block } from "./Block";
import { BlocksWrapper } from "./Block/utils/BlocksWrapper";
import { DirektivPagesType } from "../schema";

type PageCompilerProps = {
  page: DirektivPagesType;
};

export const PageCompiler = ({ page }: PageCompilerProps) => (
  <BlocksWrapper>
    {page.blocks.map((block, index) => (
      <Block key={index} block={block} />
    ))}
  </BlocksWrapper>
);
