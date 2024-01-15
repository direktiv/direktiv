import { Check, Copy, XCircle } from "lucide-react";
import { ComponentPropsWithRef, FC, useEffect, useState } from "react";

import Button from "../Button";

type ButtonPropsType = ComponentPropsWithRef<typeof Button>;

const CopyButton: FC<{
  value: string;
  buttonProps?: ButtonPropsType & { "data-testid"?: string };
  children?: (copied: boolean) => React.ReactNode;
}> = ({ value, buttonProps: { onClick, ...buttonProps } = {}, children }) => {
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState(false);

  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (copied === true) {
      timeout = setTimeout(() => {
        setCopied(false);
      }, 1000);
    }
    return () => clearTimeout(timeout);
  }, [copied]);

  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (error === true) {
      timeout = setTimeout(() => {
        setError(false);
      }, 1000);
    }
    return () => clearTimeout(timeout);
  }, [error]);

  const getIcon = () => {
    if (error) {
      return <XCircle />;
    }
    if (copied) {
      return <Check />;
    }
    return <Copy />;
  };
  return (
    <Button
      variant="ghost"
      onClick={(e) => {
        if (navigator.clipboard) {
          navigator.clipboard.writeText(value);
          setCopied(true);
        } else {
          setError(true);
          console.warn(
            "Clipboard API is not available, you migh not be on HTTPS"
          );
        }
        onClick?.(e);
      }}
      {...buttonProps}
    >
      {getIcon()}
      {children && children(copied)}
    </Button>
  );
};

export default CopyButton;
