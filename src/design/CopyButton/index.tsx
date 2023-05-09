import { ComponentProps, FC, useEffect, useState } from "react";
import { Copy, CopyCheck } from "lucide-react";

import Button from "../Button";

type ButtonPropsType = ComponentProps<typeof Button>;

const CopyButton: FC<{
  value: string;
  buttonProps?: ButtonPropsType;
  children?: (copied: boolean) => React.ReactNode;
}> = ({ value, buttonProps, children }) => {
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
      variant="ghost"
      onClick={() => {
        navigator.clipboard.writeText(value);
        setCopied(true);
      }}
      {...buttonProps}
    >
      {copied ? <CopyCheck /> : <Copy />}
      {children && children(copied)}
    </Button>
  );
};

export default CopyButton;
