import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type AvatarProps = HTMLAttributes<HTMLDivElement> & {
  size?: "xs" | "sm" | "lg" | "xlg";
  src?: string;
  className?: string;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Avatar: FC<AvatarProps> = ({
  size = "lg",
  className,
  children,
  src,
  placeholder,
  ...props
}) => (
  <div className={clsx("avatar", placeholder && "placeholder")} {...props}>
    <div
      className={clsx(
        className,
        "rounded-full",
        size === "xlg" && "w-32",
        size === "lg" && "w-24",
        size === "sm" && "w-16",
        size === "xs" && "w-8",
        placeholder && "bg-neutral-focus text-neutral-content"
      )}
    >
      {placeholder ? (
        <span className="text-3xl">{placeholder}</span>
      ) : (
        <img
          src={
            src ||
            "https://daisyui.com/images/stock/photo-1534528741775-53994a69daeb.jpg"
          }
        />
      )}
      {children}
    </div>
  </div>
);

const AvatarWithForwaredRef = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ ...props }, ref) => <Avatar forwaredRef={ref} {...props} />
);

AvatarWithForwaredRef.displayName = "Button";

export default AvatarWithForwaredRef;
