import { PropsWithChildren } from "react";

export const BlockWrapper = ({ children }: PropsWithChildren) => (
  <div className="border p-3 border-dashed">{children}</div>
);
