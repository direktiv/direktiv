import { Block } from "../PageCompiler/Block";
import { DirektivPagesType } from "../schema";
import { LocalDialogContainer } from "~/design/LocalDialog/container";
import { PageCompilerContextProvider } from "../PageCompiler/context/pageCompilerContext";
import { VisitorBlockList } from "../PageCompiler/Block/utils/BlockList";
import { useState } from "react";

export const LiveLayout = ({ page }: { page: DirektivPagesType }) => {
  const [scrollPos, setScrollPos] = useState(0);

  return (
    <PageCompilerContextProvider
      setPage={() => {}}
      page={page}
      scrollPos={scrollPos}
      setScrollPos={setScrollPos}
      mode="live"
    >
      <LocalDialogContainer className="mx-auto max-w-screen-lg">
        <VisitorBlockList path={[]}>
          {page.blocks.map((block, index) => (
            <Block key={index} block={block} blockPath={[index]} />
          ))}
        </VisitorBlockList>
      </LocalDialogContainer>
    </PageCompilerContextProvider>
  );
};
