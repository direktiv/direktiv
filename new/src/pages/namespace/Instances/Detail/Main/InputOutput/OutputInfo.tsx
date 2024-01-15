import { FC, PropsWithChildren } from "react";

import Toolbar from "./Toolbar";

const OutputInfo: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex grow flex-col gap-5 pb-12">
    <Toolbar variant="output" />
    <div className="flex h-full flex-col items-center justify-center gap-y-5 p-10">
      <span className="text-center text-gray-11">{children}</span>
    </div>
  </div>
);

export default OutputInfo;
