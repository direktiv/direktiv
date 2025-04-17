import { Block } from "./Block";
import { DirektivPagesType } from "../schema";

type PageCompilerProps = {
  page: DirektivPagesType;
};

export const PageCompiler = ({ page }: PageCompilerProps) => (
  <>
    {page.blocks.map((block, index) => (
      <Block key={index} block={block} />
    ))}
  </>
);
