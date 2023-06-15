import { FC, PropsWithChildren } from "react";

const EmptyList: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex flex-col items-center gap-y-5 p-10">
    <span className="text-center text-sm">{children}</span>
  </div>
);

export default EmptyList;
