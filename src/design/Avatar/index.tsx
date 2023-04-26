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
      "flex items-center justify-center rounded-full bg-gray-8",
      size === "xlg" && "h-32 w-32 text-2xl",
      size === "lg" && "h-24 w-24 text-xl",
      size === "sm" && "h-16 w-16 text-base",
      size === "xs" && "h-8 w-8 text-xs",
      className
    )}
  >
    {children}
  </div>
);

const AvatarWithForwaredRef = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ ...props }, ref) => <Avatar forwaredRef={ref} {...props} />
);

AvatarWithForwaredRef.displayName = "Button";

export default AvatarWithForwaredRef;
