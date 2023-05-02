import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type AvatarProps = HTMLAttributes<HTMLDivElement> & {
  className?: string;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Avatar: FC<AvatarProps> = ({ className, children, ...props }) => (
  <div
    {...props}
    className={clsx(
      "flex h-7 w-7 items-center justify-center rounded-full text-xs",
      "bg-primary-500 text-gray-1 dark:text-gray-dark-1",
      className
    )}
  >
    {children ? children : ""}
  </div>
);

const AvatarWithForwaredRef = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ ...props }, ref) => <Avatar forwaredRef={ref} {...props} />
);

AvatarWithForwaredRef.displayName = "Avatar";

export default AvatarWithForwaredRef;
