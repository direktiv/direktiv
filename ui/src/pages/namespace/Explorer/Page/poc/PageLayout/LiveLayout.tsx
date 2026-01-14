import { Block } from "../PageCompiler/Block";
import { DirektivPagesType } from "../schema";
import { PageCompilerContextProvider } from "../PageCompiler/context/pageCompilerContext";
import { PagePreviewContainer } from "../BlockEditor/PagePreviewContainer";
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
      <div className="relative lg:flex lg:h-[calc(100vh-230px)] lg:flex-col">
        <PagePreviewContainer>
          <VisitorBlockList path={[]}>
            {page.blocks.map((block, index) => (
              <Block key={index} block={block} blockPath={[index]} />
            ))}
          </VisitorBlockList>
        </PagePreviewContainer>
      </div>
    </PageCompilerContextProvider>
  );
};
