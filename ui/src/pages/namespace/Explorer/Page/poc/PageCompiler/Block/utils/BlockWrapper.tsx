import { PropsWithChildren, Suspense } from "react";

import { Loading } from "./Loading";

export const BlockWrapper = ({ children }: PropsWithChildren) => (
  <div className="border p-3 border-dashed">
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </div>
);
