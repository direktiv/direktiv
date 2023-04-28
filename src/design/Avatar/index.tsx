import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type AvatarProps = HTMLAttributes<HTMLDivElement> & {
  size?: "xs" | "sm" | "lg" | "xlg";
  className?: string;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Avatar: FC<AvatarProps> = ({
  size = "lg",
  className,
  children,
  ...props
}) => (
  <div
    {...props}
    className={clsx(
      "flex items-center justify-center rounded-full h-8 w-8 text-xs",
      "bg-primary-500 text-white",
      className
    )}
  >
    {children ? children : '??'}
  </div>
);

const AvatarWithForwaredRef = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ ...props }, ref) => <Avatar forwaredRef={ref} {...props} />
);

AvatarWithForwaredRef.displayName = "Avatar";

export default AvatarWithForwaredRef;
