import { MouseEventHandler, ReactElement, cloneElement } from "react";

type StopPropagationProps = {
  children: ReactElement;
  asChild?: boolean;
  onClick?: React.MouseEventHandler;
};

export const StopPropagation = ({
  children,
  asChild,
  onClick,
}: StopPropagationProps) => {
  const handleClick: MouseEventHandler = (event) => {
    event.stopPropagation();
    onClick?.(event);
  };

  if (asChild) {
    return cloneElement(children, {
      onClick: (e: React.MouseEvent) => {
        handleClick(e);
        children.props.onClick?.(e);
      },
    });
  }

  return <div onClick={handleClick}>{children}</div>;
};
