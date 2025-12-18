import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { ReactElement } from "react";
import { VisitorBlockList } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block/utils/BlockList";

type BlockListProps = {
  horizontal?: boolean;
  children: ReactElement[];
  path: BlockPathType;
};

export const BlockList = (props: BlockListProps) => (
  <div>
    <pre>Page server app</pre>
    <VisitorBlockList {...props} />
  </div>
);
