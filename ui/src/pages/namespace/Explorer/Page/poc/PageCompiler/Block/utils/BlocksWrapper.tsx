import { PropsWithChildren, Suspense } from "react";

import { Loading } from "./Loading";

export const BlocksWrapper = ({ children }: PropsWithChildren) => (
  <div className="flex flex-col gap-3">
    <Suspense fallback={<Loading />}>{children}</Suspense>
  </div>
);
