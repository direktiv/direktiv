import React, { FC, HTMLAttributes } from "react";

import { twMergeClsx } from "~/util/helpers";

export type AvatarProps = HTMLAttributes<HTMLDivElement> & {
  className?: string;
  children?: React.ReactNode;
};

const Avatar: FC<AvatarProps> = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ className, children, ...props }, ref) => (
    <div
      {...props}
      className={twMergeClsx(
        "flex h-7 w-7 items-center justify-center rounded-full text-xs",
        "bg-primary-500 text-gray-1 dark:text-gray-dark-1",
        className
      )}
      ref={ref}
    >
      {children ? children : ""}
    </div>
  )
);

Avatar.displayName = "Avatar";

export default Avatar;
