import { MouseEventHandler, ReactElement, cloneElement } from "react";

type StopPropagationProps = {
  children: ReactElement;
  onClick?: React.MouseEventHandler;
};

export const StopPropagation = ({
  children,
  onClick,
}: StopPropagationProps) => {
  const handleClick: MouseEventHandler = (event) => {
    event.stopPropagation();
    onClick?.(event);
  };

  return cloneElement(children, {
    onClick: (e: React.MouseEvent) => {
      handleClick(e);
      children.props.onClick?.(e);
    },
  });
};
