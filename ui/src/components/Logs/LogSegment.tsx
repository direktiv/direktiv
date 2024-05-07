import { FC, PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";

type LogSegmentProps = PropsWithChildren & {
  className?: string;
  display: boolean;
};

export const LogSegment: FC<LogSegmentProps> = ({
  display,
  className,
  children,
}) => {
  if (!display) return <></>;
  return <span className={twMergeClsx("mr-3", className)}>{children}</span>;
};
