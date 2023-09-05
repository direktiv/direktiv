import { Check, Copy, Frown } from "lucide-react";
import { ComponentProps, FC, useEffect, useState } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../Tooltip";

import Button from "../Button";
import { ConditionalWrapper } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type ButtonPropsType = ComponentProps<typeof Button>;

const CopyButton: FC<{
  value: string;
  buttonProps?: ButtonPropsType;
  children?: (copied: boolean) => React.ReactNode;
}> = ({ value, buttonProps: { onClick, ...buttonProps } = {}, children }) => {
  const [copied, setCopied] = useState(false);
  const clipboardNotAvailable = !navigator.clipboard;
  const { t } = useTranslation();

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
    <ConditionalWrapper
      condition={clipboardNotAvailable}
      wrapper={(children) => (
        <TooltipProvider delayDuration={100}>
          <Tooltip>
            <TooltipTrigger asChild>
              <div>{children}</div>
            </TooltipTrigger>
            <TooltipContent>
              <div className="flex w-56 flex-col items-center gap-2 text-center">
                <Frown />
                <div>{t("components.copyButton.notSuported")}</div>
              </div>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    >
      <Button
        variant="ghost"
        onClick={(e) => {
          navigator.clipboard.writeText(value);
          setCopied(true);
          onClick?.(e);
        }}
        {...buttonProps}
        {...(clipboardNotAvailable ? { disabled: true } : {})}
      >
        {copied ? <Check /> : <Copy />}
        {children && children(copied)}
      </Button>
    </ConditionalWrapper>
  );
};

export default CopyButton;
