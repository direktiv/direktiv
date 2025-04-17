import { PropsWithChildren } from "react";

export const BlocksWrapper = ({ children }: PropsWithChildren) => (
  <div className="flex flex-col gap-3">{children}</div>
);
