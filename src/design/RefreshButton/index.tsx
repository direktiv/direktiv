import { ComponentProps, FC, useEffect, useState } from "react";

import Button from "../Button";
import { RefreshCcw } from "lucide-react";

type ButtonPropsType = ComponentProps<typeof Button>;

const RefreshButton: FC<ButtonPropsType> = ({
  onClick,
  children,
  ...props
}) => {
  const [spinning, setSpinning] = useState(false);

  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (spinning === true) {
      timeout = setTimeout(() => {
        setSpinning(false);
      }, 500);
    }
    return () => clearTimeout(timeout);
  }, [spinning]);

  return (
    <Button
      onClick={(e) => {
        onClick?.(e);
        setSpinning(true);
      }}
      {...props}
    >
      <RefreshCcw className={spinning ? "animate-spin" : ""} />
      {children}
    </Button>
  );
};

export default RefreshButton;
