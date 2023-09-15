import { Check, Copy } from "lucide-react";
import { ComponentProps, FC, useEffect, useState } from "react";

import Button from "../Button";

type ButtonPropsType = ComponentProps<typeof Button>;

const CopyButton: FC<{
  testid?: string;
  value: string;
  buttonProps?: ButtonPropsType;
  children?: (copied: boolean) => React.ReactNode;
}> = ({
  testid,
  value,
  buttonProps: { onClick, ...buttonProps } = {},
  children,
}) => {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (copied === true) {
      timeout = setTimeout(() => {
        setCopied(false);
      }, 1000);
    }
    return () => clearTimeout(timeout);
  }, [copied]);

  return (
    <Button
      data-testid={testid}
      variant="ghost"
      onClick={(e) => {
        navigator.clipboard.writeText(value);
        setCopied(true);
        onClick?.(e);
      }}
      {...buttonProps}
    >
      {copied ? <Check /> : <Copy />}
      {children && children(copied)}
    </Button>
  );
};

export default CopyButton;
